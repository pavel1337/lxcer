package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

var (
	StatusRunning int = 103
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		// DisableColors: true,
	})
}

func main() {
	config := loadConfig()

	for _, r := range config.ResticRepos {
		r.Check()
	}

	if config.Cleanup {
		cleanup()
	}

	if config.Local {
		localBackup(config)
		return
	}

	hh := toHosts(config.Hosts)
	for _, h := range hh {
		h.Backup(config)
	}
}

func localBackup(config *Config) {
	cc, err := listContainersLocal()
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range filterContainers(cc, config.Blacklist) {
		if len(c.Snapshots) > 0 {
			err := c.DeleteSnapshots()
			if err != nil {
				log.Error(err)
				continue
			}
			log.Infof("Local snapshots deleted for %s", c.Name)
		}

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

func cleanup() {
	cc, err := listContainersLocal()
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range cc {
		err = c.Delete()
		if err != nil {
			log.Error(err)
			continue
		}
		log.Infof("Local container %s deleted", c.Name)
	}

	images, err := listImagesLocal()
	if err != nil {
		log.Fatal(err)
	}
	for _, image := range images {
		err = image.Delete()
		if err != nil {
			log.Error(err)
			continue
		}
		log.Infof("Local image %s deleted", image.Fingerprint)

	}
}

func listContainersLocal() ([]Container, error) {
	var (
		cc     []Container
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)

	cmd := exec.Command("lxc", "list", "--format", "json")
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		return nil, errors.New(Stderr.String())
	}
	err = json.Unmarshal(Stdout.Bytes(), &cc)
	if err != nil {
		return nil, err
	}
	return cc, nil
}

func listImagesLocal() ([]Image, error) {
	var (
		images []Image
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)

	cmd := exec.Command("lxc", "image", "list", "--format", "json")
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		return nil, errors.New(Stderr.String())
	}
	err = json.Unmarshal(Stdout.Bytes(), &images)
	if err != nil {
		return nil, err
	}
	return images, nil
}

func filterContainers(icc []Container, blacklist []string) []Container {
	var cc []Container
	for _, c := range icc {
		if c.StatusCode != StatusRunning {
			continue
		}
		if contains(blacklist, c.Name) {
			continue
		}
		cc = append(cc, c)
	}
	return cc
}
