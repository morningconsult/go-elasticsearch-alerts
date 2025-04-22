// Copyright 2019 The Morning Consult, LLC or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//         https://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/morningconsult/go-elasticsearch-alerts/command"
)

const banner = "Go Elasticsearch Alerts version %v, commit %v, built %v\n"

var (
	// version indicates which version of the binary is running.
	version = "dev"

	// commit is the commit hash of the git repository from which
	// the binary is built.
	commit = "none"

	// date is the date on which the binary was built.
	date = "unknown"
)

func main() {
	var versionFlag bool
	flag.BoolVar(&versionFlag, "version", false, "print version and exit")
	flag.Parse()

	// Exit safely when version is used
	if versionFlag {
		fmt.Printf(banner, version, commit, date)
		os.Exit(0)
	}

	os.Exit(command.Run())
}
