/*
 Copyright (c) 42Crunch Ltd. All rights reserved.
 Licensed under the GNU Affero General Public License version 3. See LICENSE.txt in the project root for license information.
*/

package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
)

const defaultplatformService = "services.us.42crunch.cloud:8001"
const defaultScandImage = "42crunch/scand-agent:latest"
const defaultScandImagepullPolicy = "IfNotPresent"
const defaultExpirationTime = 86400 // expire completed jobs after 24h

var namespace string
var platformService string
var scandImage string
var scandImage_pullPolicy string
var expirationTimeInt int64
var clientset *kubernetes.Clientset
var podconfig *v1.PodSpec

func main() {

	podconfigFile := flag.String("podconfig", "", "pod configuration file")

	flag.Parse()

	readEnvConfig()

	if *podconfigFile != "" {
		readPodConfig(*podconfigFile)
	}

	connectk8sClient()

	log.Printf("Starting scand manager: NAMESPACE: '%s', PLATFORM_HOST: '%s', SCAND_IMAGE: '%s', EXPIRATION_TIME: '%d'",
		namespace, platformService, scandImage, expirationTimeInt)

	if podconfig != nil {
		log.Printf("Using pod configuration file: %s", *podconfigFile)
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/job/{name}", jobStatus).Methods("GET")
	r.HandleFunc("/api/job", jobList).Methods("GET")
	r.HandleFunc("/api/job", jobLaunch).Methods("POST")
	r.HandleFunc("/api/job/{name}", jobDelete).Methods("DELETE")
	r.HandleFunc("/api/logs/{name}", jobLogs).Methods("GET")

	http.Handle("/", r)
	srv := &http.Server{
		Handler: r,
		Addr:    ":8090",
		//timeouts
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func jobLaunch(w http.ResponseWriter, r *http.Request) {

	job, err := readJobRequest(r)
	if err != nil {
		writeErrorMsg(err, w, http.StatusBadRequest)
		return
	}

	jobsClient := clientset.BatchV1().Jobs(namespace)

	_, err = jobsClient.Create(context.TODO(), getk8sJob(*job), metav1.CreateOptions{})
	if err != nil {
		log.Printf("ERROR, failed to create to create the job '%s': %s", job.Name, err)
		writeErrorMsg(err, w, http.StatusBadRequest)
		return
	}

	log.Println("Launched job:", job.Name)

	sendResponse(w, map[string]interface{}{
		"status": "started",
		"name":   job.Name,
	})
}

func jobStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	name := vars["name"]

	if !isValidJobName(name) {
		log.Println("ERROR, failed get job status, invalid job name:", name)
		writeErrorMsg(errors.New("invalid job name"), w, http.StatusBadRequest)
		return
	}

	job, err := clientset.BatchV1().Jobs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Printf("ERROR, unable to retrieve job status %s: %s", name, err)
		writeErrorMsg(err, w, http.StatusBadRequest)
		return
	}

	sendResponse(w, map[string]interface{}{
		"status": getJobStatus(job),
		"name":   name,
	})
}

func jobList(w http.ResponseWriter, r *http.Request) {
	requirement, _ := labels.NewRequirement("jobgroup", selection.Equals, []string{"scand-job"})
	selector := labels.NewSelector()
	selector = selector.Add(*requirement)

	jobs, err := clientset.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		log.Printf("ERROR, unable to retrieve job list: %s", err)
		writeErrorMsg(err, w, http.StatusBadRequest)
		return
	}

	var statuses = make([]map[string]string, 0)

	for _, job := range jobs.Items {
		statuses = append(statuses, map[string]string{
			"name":   job.Name,
			"status": getJobStatus(&job),
		})
	}

	sendResponse(w, map[string]interface{}{
		"jobs": statuses,
	})
}

func jobDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	name := vars["name"]

	if !isValidJobName(name) {
		log.Println("ERROR, failed to delete a job, invalid job name:", name)
		writeErrorMsg(errors.New("invalid job name"), w, http.StatusBadRequest)
		return
	}

	background := metav1.DeletePropagationBackground
	if err := clientset.BatchV1().Jobs(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &background,
	}); err != nil {
		log.Printf("ERROR, unable to delete job %s: %s", name, err)
		writeErrorMsg(err, w, http.StatusBadRequest)
		return
	}

	log.Println("Deleted job:", name)

	sendResponse(w, map[string]interface{}{
		"status": "deleted",
		"name":   name,
	})
}

func jobLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	name := vars["name"]

	if !isValidJobName(name) {
		log.Println("ERROR, failed get job logs, invalid job name:", name)
		writeErrorMsg(errors.New("invalid job name"), w, http.StatusBadRequest)
		return
	}

	podName, err := getPod(name)
	if err != nil {
		log.Printf("ERROR, unable to find pod for the job %s: %s", name, err)
		writeErrorMsg(err, w, http.StatusBadRequest)
		return
	}

	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{})

	podLogs, err := req.Stream(context.Background())
	if err != nil {
		log.Printf("ERROR, unable to retrieve the logs for the job %s: %s", name, err)
		writeErrorMsg(err, w, http.StatusBadRequest)
		return
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		log.Printf("ERROR, unable to retrieve the logs for the job %s: %s", name, err)
		writeErrorMsg(err, w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(buf.Bytes())
}
