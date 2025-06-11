/*
 Copyright (c) 42Crunch Ltd. All rights reserved.
 Licensed under the GNU Affero General Public License version 3. See LICENSE.txt in the project root for license information.
*/

package main

import (
	"log"
	"os"
	"strconv"
)

func readEnvConfig() {
	namespace = os.Getenv("NAMESPACE")
	if namespace == "" {
		log.Fatal("ERROR, no NAMESPACE env variable is set")
	}

	platformService = os.Getenv("PLATFORM_SERVICE")
	if platformService == "" {
		log.Printf("No PLATFORM_SERVICE env variable is set, using default: %s", defaultplatformService)
		platformService = defaultplatformService
	}

	scandImage = os.Getenv("SCAND_IMAGE")
	if scandImage == "" {
		log.Printf("No SCAND_IMAGE env variable is set, using default: %s", defaultScandImage)
		scandImage = defaultScandImage
	}

	expirationTime := os.Getenv("EXPIRATION_TIME")
	if expirationTime == "" {
		log.Printf("No EXPIRATION_TIME env variable is set, using default: %d", defaultExpirationTime)
		expirationTimeInt = defaultExpirationTime
	} else {
		var err error
		expirationTimeInt, err = strconv.ParseInt(expirationTime, 10, 32)
		if err != nil {
			log.Fatalf("ERROR, invalid EXPIRATION_TIME env var: '%s': %s", expirationTime, err)
		}

		if expirationTimeInt < 0 {
			log.Fatalf("ERROR, invalid EXPIRATION_TIME env var: '%d', must be >= 0", expirationTimeInt)
		}

		if expirationTimeInt > maxExpirationTime {
			log.Fatalf("ERROR, invalid EXPIRATION_TIME env var: '%d', must be <= %d", expirationTimeInt, maxExpirationTime)
		}
	}
}
