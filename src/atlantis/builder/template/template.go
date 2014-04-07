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
