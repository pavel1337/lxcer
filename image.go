package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
)

type Image struct {
	Fingerprint string `json:"fingerprint"`
}

func (i *Image) Delete() error {
	var (
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)

	cmd := exec.Command("lxc", "image", "delete", fmt.Sprintf("%s", i.Fingerprint))
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(Stderr.String())
	}
	return nil
}
