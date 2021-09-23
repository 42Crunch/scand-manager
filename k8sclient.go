/*
 Copyright (c) 42Crunch Ltd. All rights reserved.
 Licensed under the GNU Affero General Public License version 3. See LICENSE.txt in the project root for license information.
*/

package main

import (
	"context"
	"errors"
	"log"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func connectk8sClient() {
	var err error

	clientset, err = createk8sClient()
	if err != nil {
		log.Fatalf("ERROR, failed to create k8s client: %s", err)
	}
}

func createk8sClient() (*kubernetes.Clientset, error) {
	// get the config from inside of the cluster
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	cli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func getk8sJob(job Job) *batchv1.Job {
	//int 32 parameters
	backoffLimit := int32(0)
	tTLSecondsAfterFinished := int32(job.ExpirationTime)

	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   job.Name,
			Labels: map[string]string{"jobgroup": "scand-job"},
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  job.Name,
							Image: job.ScandImage,
							Env:   job.Env,
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &tTLSecondsAfterFinished,
		},
	}
}

func getPod(jobName string) (string, error) {
	requirement, _ := labels.NewRequirement("job-name", selection.Equals, []string{jobName})
	selector := labels.NewSelector()
	selector = selector.Add(*requirement)

	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return "", err
	}

	if len(pods.Items) > 0 {
		return pods.Items[0].GetName(), nil
	}

	return "", errors.New("matching pod is not found")
}
