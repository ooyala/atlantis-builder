package main

import (
	atlantis "atlantis/types"
	"flag"
	"os"
	"text/template"
)

func main() {
	ifileStr := flag.String("i", "", "template file to read")
	ofileStr := flag.String("o", "", "out file to write")

	cfg, err := atlantis.LoadAppConfig()
	if err != nil {
		panic(err)
	}

	tmpl, err := template.ParseFiles(*ifileStr)
	if err != nil {
		panic(err)
	}

	ofile, err := os.Create(*ofileStr)
	if err != nil {
		panic(err)
	}

	if err := tmpl.Execute(ofile, cfg); err != nil {
		panic(err)
	}
	ofile.Close()
}
