package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

type Host struct {
	Name       string
	Containers []Container
}

func toHosts(strings []string) []Host {
	var hh []Host
	for _, s := range strings {
		hh = append(hh, Host{Name: s})
	}
	return hh
}

func (h *Host) Backup(config *Config) {
	log := log.WithField("host", h.Name)
	err := h.GetContainers()
	if err != nil {
		log.Error(err)
	}
	cc := filterContainers(h.Containers, config.Blacklist)
	for _, c := range cc {
		log = log.WithField("container", c.Name)

		t := time.Now()
		if c.SnapshotExists(sn) {
			err := c.DeleteSnapshotRemote(sn, h.Name)
			if err != nil {
				log.Error(err)
				continue
			}
			log.WithField("spent", time.Since(t)).Infof("Delete snapshot")
		}

		t = time.Now()
		err = c.CreateSnapshotRemote(sn, h.Name)
		if err != nil {
			log.Error(err)
			continue
		}
		log.WithField("spent", time.Since(t)).Infof("Create remote snapshot %s", sn)

		t = time.Now()
		err = c.PublishRemote(sn, h.Name)
		if err != nil {
			log.Error(err)
			continue
		}
		log.WithField("spent", time.Since(t)).Infof("Publish remote snapshot %s as image", sn)

		t = time.Now()
		err = c.DeleteSnapshotRemote(sn, h.Name)
		if err != nil {
			log.Error(err)
			continue
		}
		log.WithField("spent", time.Since(t)).Infof("Delete remote snapshot %s", sn)

		t = time.Now()
		err = c.ExportImage()
		if err != nil {
			log.Error(err)
			continue
		}
		log.WithField("spent", time.Since(t)).Infof("Export image as %s.tar", c.Name)

		t = time.Now()
		err = c.DeleteImage()
		if err != nil {
			log.Error(err)
			continue
		}
		log.WithField("spent", time.Since(t)).Infof("Delete image")

		t = time.Now()
		err = c.CompressWithZst()
		if err != nil {
			log.Error(err)
			continue
		}
		log.WithField("spent", time.Since(t)).Infof("Compress %s.tar to %s.tar.zst", c.Name, c.Name)

		t = time.Now()
		err = c.DeleteImageTar()
		if err != nil {
			log.Error(err)
			continue
		}
		log.WithField("spent", time.Since(t)).Infof("Delete %s.tar", c.Name)

		for _, r := range config.ResticRepos {
			t = time.Now()
			err := r.Backup(fmt.Sprintf("%s.tar.zst", c.Name))
			if err != nil {
				log.Error(err)
				continue
			}
			log.WithField("spent", time.Since(t)).Infof("Backup %s.tar.zst to %s", c.Name, r.Path)
		}

		t = time.Now()
		err = c.DeleteImageTarZst()
		if err != nil {
			log.Error(err)
			continue
		}
		log.WithField("spent", time.Since(t)).Infof("Delete %s.tar.zst", c.Name)

	}

}

func (h *Host) GetContainers() error {
	var (
		cc     []Container
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)

	cmd := exec.Command("lxc", "list", fmt.Sprintf("%s:", h.Name), "--format", "json")
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(Stderr.String())
	}
	err = json.Unmarshal(Stdout.Bytes(), &cc)
	if err != nil {
		return err
	}
	for _, c := range cc {
		c.Host = h.Name
		h.Containers = append(h.Containers, c)
	}
	return nil
}
