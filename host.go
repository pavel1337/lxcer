package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

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
	err := h.GetContainers()
	if err != nil {
		log.Error(err)
	}
	cc := filterContainers(h.Containers, config.Blacklist)
	for _, c := range cc {
		if len(c.Snapshots) > 0 {
			err := c.DeleteSnapshotsRemote(h.Name)
			if err != nil {
				log.Error(err)
				continue
			}
			log.Infof("Remote snapshots deleted for %s:%s", h.Name, c.Name)
		}

		err = c.CreateSnapshotRemote("ssnet", h.Name)
		if err != nil {
			log.Error(err)
			continue
		}
		log.Infof("Snapshot %s created for %s:%s", "ssnet", h.Name, c.Name)

		err = c.CopySnapshot("ssnet", h.Name)
		if err != nil {
			log.Error(err)
			continue
		}
		log.Infof("Snapshot %s copied from %s:%s", "ssnet", h.Name, c.Name)

		err = c.Publish()
		if err != nil {
			log.Error(err)
			continue
		}
		log.Infof("Container %s published as image", c.Name)

		err = c.ExportImage()
		if err != nil {
			log.Error(err)
			continue
		}
		log.Infof("Image of container %s exported to %s.tar", c.Name, c.Name)

		err = c.DeleteImage()
		if err != nil {
			log.Error(err)
			continue
		}
		log.Infof("Image of container %s delete", c.Name)

		err = c.CompressWithZst()
		if err != nil {
			log.Error(err)
			continue
		}
		log.Infof("Exported image of %s compressed to %s.tar.zst", c.Name, c.Name)

		err = c.DeleteImageTar()
		if err != nil {
			log.Error(err)
			continue
		}
		log.Infof("Uncompressed image of %s deleted from %s.tar", c.Name, c.Name)

		for _, r := range config.ResticRepos {
			err := r.Backup(fmt.Sprintf("%s.tar.zst", c.Name))
			if err != nil {
				log.Error(err)
				continue
			}
			log.Infof("Compressed %s.tar.zst has been backed up to %s", c.Name, r.Path)
		}

		err = c.DeleteImageTarZst()
		if err != nil {
			log.Error(err)
			continue
		}
		log.Infof("Compressed image of %s deleted from %s.tar.zst", c.Name, c.Name)

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
	h.Containers = cc
	return nil
}
