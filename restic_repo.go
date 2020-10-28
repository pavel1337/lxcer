package main

import (
	"bytes"
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type ResticRepo struct {
	Path     string `yaml:"path"`
	Password string `yaml:"password"`
}

func (r *ResticRepo) Check() {
	var (
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)
	cmd := exec.Command("restic", "check")
	cmd.Env = r.setEnv()
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalln(Stderr.String())
	}
	log.Infof("restic repository %s is OK", r.Path)
}

func (r *ResticRepo) Backup(path string) error {
	var (
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)
	cmd := exec.Command("restic", "backup", path)
	cmd.Env = r.setEnv()
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (r *ResticRepo) setEnv() []string {
	var envs []string
	envs = append(envs, fmt.Sprintf("RESTIC_REPOSITORY=%s", r.Path))
	envs = append(envs, fmt.Sprintf("RESTIC_PASSWORD=%s", r.Password))
	return envs
}
