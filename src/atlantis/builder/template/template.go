/* Copyright 2014 Ooyala, Inc. All rights reserved.
 *
 * This file is licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
 * except in compliance with the License. You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is
 * distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and limitations under the License.
 */

package template

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

const RunitTemplate = `#!/bin/bash
cd /app
{ exec chpst -u user1 {{.Cmd}} | logger -p local{{.Num}}.info; } 2>&1 | logger -p local{{.Num}}.error
`

type CmdAndNum struct {
	Cmd string
	Num int
}

func WriteRunitScript(path string, cmd string, idx int) {
	tmpl := template.Must(template.New("runit").Parse(RunitTemplate))
	if fh, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0500); err != nil {
		panic(err)
	} else {
		if err := tmpl.Execute(fh, CmdAndNum{cmd, idx}); err != nil {
			panic(err)
		}
	}
}

const RsyslogAppTemplate = `# config for app{{.}}
$outchannel app{{.}}Info,/var/log/atlantis/app{{.}}/stdout.log,10485760,/etc/logrot
$outchannel app{{.}}Error,/var/log/atlantis/app{{.}}/stderr.log,10485760,/etc/logrot

local{{.}}.=info  :omfile:$app{{.}}Info
local{{.}}.=error :omfile:$app{{.}}Error
local{{.}}.=crit  :omfile:$app{{.}}Error
`

func WriteRsyslogAppConfig(path string, idx int) {
	tmpl := template.Must(template.New("rsyslog").Parse(RsyslogAppTemplate))
	if fh, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0500); err != nil {
		panic(err)
	} else {
		if err := tmpl.Execute(fh, idx); err != nil {
			panic(err)
		}
	}
}

func WriteRsyslogCustomConfig(path string, fac string, desc map[string]string) {
	name := desc["name"]
	delete(desc, "name")
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(`# config for %s on %s\n`, name, fac))
	for key, val := range desc {
		key = strings.ToLower(key)
		buffer.WriteString(fmt.Sprintf(`$outchannel %s%s,/var/log/atlantis/%s/%s.log,10485760,/etc/logrot\n`, fac, key, name, val))
		buffer.WriteString(fmt.Sprintf(`%s.=%s  :omfile:$%s%s\n`, fac, key, fac, key))
	}
	if fh, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0500); err != nil {
		panic(err)
	} else {
		if _, err := fh.Write(buffer.Bytes()); err != nil {
			panic(err)
		}
	}
}

const SetupTemplate = `#!/bin/bash -x
{{range .SetupCommands}}
{{.}}
{{end}}
`

func WriteSetupScript(path string, manifest interface{}) {
	tmpl := template.Must(template.New("setup").Parse(SetupTemplate))
	if fh, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0500); err != nil {
		panic(err)
	} else {
		if err := tmpl.Execute(fh, manifest); err != nil {
			panic(err)
		}
	}
}
