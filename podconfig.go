/*
 Copyright (c) 42Crunch Ltd. All rights reserved.
 Licensed under the GNU Affero General Public License version 3. See LICENSE.txt in the project root for license information.
*/

package main

import (
	"io/ioutil"
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

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Can't read pod config file '%s': %s", filename, err)
	}

	podconfig = &v1.PodSpec{}
	err = yaml.Unmarshal(data, podconfig)
	if err != nil {
		log.Fatalf("Can't parse pod config file '%s': %s", filename, err)
	}

	if podconfig.Affinity == nil {
		log.Fatal("Affinity is required in pod config file.")
	}
}
