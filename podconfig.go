/*
 Copyright (c) 42Crunch Ltd. All rights reserved.
 Licensed under the GNU Affero General Public License version 3. See LICENSE.txt in the project root for license information.
*/

package main

import (
	"log"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
)

func readPodConfig(filename string) {
	if !strings.HasSuffix(filename, ".yaml") && !strings.HasSuffix(filename, ".yml") {
		log.Fatalf("Pod config file '%s' must be a YAML file.", filename)
	}

	if _, err := os.Stat(filename); err != nil {
		log.Fatalf("Can't find pod config file '%s': %s", filename, err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Can't read pod config file '%s': %s", filename, err)
	}

	podconfig = &v1.PodSpec{} // Initialize podconfig

	err = yaml.Unmarshal(data, podconfig)
	if err != nil {
		log.Fatalf("Can't parse pod config file '%s': %s", filename, err)
	}

	// Check and print the values if found
	if podconfig.Affinity != nil {
		log.Printf("Affinity: %+v\n", podconfig.Affinity)
	}

	if len(podconfig.ImagePullSecrets) > 0 {
		log.Printf("ImagePullSecrets: %+v\n", podconfig.ImagePullSecrets)
	}

	if podconfig.Affinity == nil && len(podconfig.ImagePullSecrets) == 0 {
		log.Printf("Neither Affinity nor ImagePullSecrets could be found in %s.", filename)
	}
}
