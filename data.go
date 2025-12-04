/*
 Copyright (c) 42Crunch Ltd. All rights reserved.
 Licensed under the GNU Affero General Public License version 3. See LICENSE.txt in the project root for license information.
*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

type JobRequest struct {
	Token           string
	Name            string
	ExpirationTime  int64
	PlatformService string
	ScandImage      string
	Env             map[string]string
}

type Job struct {
	Name           string
	ExpirationTime int64
	ScandImage     string
	Env            []v1.EnvVar
}

func sendResponse(w http.ResponseWriter, response map[string]interface{}) {
	result, err := json.Marshal(response)

	if err != nil {
		log.Printf("ERROR, failed to marshal response: %s", err)
		writeErrorMsg(err, w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(result))
}

func writeErrorMsg(err error, w http.ResponseWriter, status int) {
	json, _ := json.Marshal(map[string]string{
		"error": fmt.Sprint(err),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(json))
}

func readJobRequest(r *http.Request) (*Job, error) {
	if r.Header.Get("Content-Type") != "application/json" {
		return nil, errors.New("Content-Type must be application/json")
	}

	// set default values which can be overriden by the request
	job := JobRequest{
		Name:            fmt.Sprintf("scand-%s", uuid.New().String()),
		ExpirationTime:  expirationTimeInt,
		PlatformService: platformService,
		ScandImage:      scandImage,
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	defer r.Body.Close()
	err := decoder.Decode(&job)

	if err != nil {
		log.Println("ERROR, failed to launch a job, can't decode the request:", err)
		return nil, errors.New("invalid request")
	}

	if !isValidJobName(job.Name) {
		log.Println("ERROR, failed to launch a job, invalid job name:", job.Name)
		return nil, errors.New("invalid job name")
	}

	if !isValidUUID(job.Token) {
		log.Println("ERROR, failed to launch a job, no token provided or invalid token:", job.Token)
		return nil, errors.New("no token provided or invalid token")
	}

	if !isValidHostnameAndPort(job.PlatformService) {
		log.Println("ERROR, failed to launch a job, invalid platform host:", job.PlatformService)
		return nil, errors.New("invalid platform host")
	}

	if !isValidScandImage(job.ScandImage) {
		log.Println("ERROR, failed to launch a job, invalid scand image:", job.ScandImage)
		return nil, errors.New("invalid scand image")
	}

	if job.ExpirationTime <= 0 {
		log.Println("ERROR, failed to launch a job, invalid expiration time:", job.ExpirationTime)
		return nil, errors.New("invalid expiration time, must be greater than 0")
	}

	if job.ExpirationTime > maxExpirationTime {
		log.Println("ERROR, failed to launch a job, expiration time too long:", job.ExpirationTime)
		return nil, fmt.Errorf("expiration time too long, must be less than %d seconds", maxExpirationTime)
	}

	var envVars []v1.EnvVar
	envVars = append(envVars, newEnvVar("SCAN_TOKEN", job.Token))
	envVars = append(envVars, newEnvVar("PLATFORM_SERVICE", job.PlatformService))

	// Helper to add a proxy env vars only if they have a value
	addProxyEnv := func(name, value string) {
		if value == "" {
			return
		}
		envVars = append(envVars, newEnvVar(name, value))
	}

	// Deployment-time defaults (from scand-manager env)
	addProxyEnv("HTTP_PROXY", defaultHTTPProxy)
	addProxyEnv("HTTPS_PROXY", defaultHTTPSProxy)
	addProxyEnv("HTTP_PROXY_API", defaultHTTPProxyAPI)
	addProxyEnv("HTTPS_PROXY_API", defaultHTTPSProxyAPI)

	// Helper function to set or override an env var in envVars
	setOrOverrideEnv := func(name, value string) {
		for i := range envVars {
			if envVars[i].Name == name {
				envVars[i].Value = value
				return
			}
		}
		envVars = append(envVars, newEnvVar(name, value))
	}

	if job.Env != nil {
		for name, value := range job.Env {
			nameUpper := strings.ToUpper(name)
			if strings.HasPrefix(nameUpper, "SECURITY_") ||
				strings.HasPrefix(nameUpper, "SCAN42C_") ||
				strings.HasPrefix(nameUpper, "HTTPS_") ||
				strings.HasPrefix(nameUpper, "HTTP_") {

				// Per-job override (or addition)
				setOrOverrideEnv(name, value)
			} else {
				log.Println("ERROR, invalid env variable in the request, must start with 'SECURITY_, SCAN42C_, or HTTP(S)_' ", name)
				return nil, errors.New("invalid environment variable name, must start with 'SECURITY_, SCAN42C_, or HTTP(S)_'")
			}
		}
	}

	return &Job{
		Name:           job.Name,
		ExpirationTime: job.ExpirationTime,
		ScandImage:     job.ScandImage,
		Env:            envVars,
	}, nil
}

func newEnvVar(name string, value string) v1.EnvVar {
	return v1.EnvVar{
		Name:  name,
		Value: value,
	}
}

func getJobStatus(job *batchv1.Job) string {
	if job.Status.Active > 0 {
		return "active"
	} else if job.Status.Succeeded > 0 {
		return "succeeded"
	} else if job.Status.Failed > 0 {
		return "failed"
	} else {
		return "unknown"
	}
}
