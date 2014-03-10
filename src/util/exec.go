package util

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

func EchoExec(cmd *exec.Cmd) []byte {
	// make streaming copies of stdout
	var buf bytes.Buffer
	outWriter := io.MultiWriter(&buf, os.Stdout)

	cmd.Stderr = os.Stderr
	cmd.Stdout = outWriter

	if err := cmd.Start(); err != nil {
		panic(err)
	}
	cmd.Wait()

	return buf.Bytes()
}
