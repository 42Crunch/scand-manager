/*
 Copyright (c) 42Crunch Ltd. All rights reserved.
 Licensed under the GNU Affero General Public License version 3. See LICENSE.txt in the project root for license information.
*/

package main

import "regexp"

func isValidJobName(name string) bool {
	matched, err := regexp.MatchString("^scand-[0-9a-zA-Z-]{1,36}$", name)
	return err == nil && matched
}

func isValidUUID(scanToken string) bool {
	matched, err := regexp.MatchString("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$", scanToken)
	return err == nil && matched
}

func isValidHostnameAndPort(service string) bool {
	matched, err := regexp.MatchString("^[^\\:]+:[0-9]{3,5}$", service)
	return err == nil && matched
}

func isValidScandImage(image string) bool {
	matched, err := regexp.MatchString("^\\P{Cc}+${1,128}$", image)
	return err == nil && matched
}
