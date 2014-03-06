package template

import (
	"os"
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

const RsyslogTemplate = `# config for app{{.}}
$template			logFormat,"%msg%\n"
$ActionFileDefaultTemplate	logFormat

$outchannel app{{.}}Info,/var/log/atlantis/app{{.}}/stdout.log,10485760,/etc/logrot
$outchannel app{{.}}Error,/var/log/atlantis/app{{.}}/stderr.log,10485760,/etc/logrot

local{{.}}.=info  :omfile:$app{{.}}Info
local{{.}}.=error :omfile:$app{{.}}Error
`

func WriteRsyslogConfig(path string, idx int) {
	tmpl := template.Must(template.New("rsyslog").Parse(RsyslogTemplate))
	if fh, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0500); err != nil {
		panic(err)
	} else {
		if err := tmpl.Execute(fh, idx); err != nil {
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
