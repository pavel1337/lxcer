package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	StatusRunning int    = 103
	sn            string = "ssnet"
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

	if config.Concurrently {
		remoteBackupsConcurrently(toHosts(config.Hosts), config)
		return
	}

	for _, h := range toHosts(config.Hosts) {
		h.Backup(config)
	}
}

func localBackupsConcurrently(config *Config) {
	ch := handleSnapshotsLocal(config)
	ch = handleImages(ch, config.LocalWorkers)
	ch = handleTars(ch, config.LocalWorkers)
	for _, r := range config.ResticRepos {
		ch = backupToRepo(ch, r)
	}
	deleteTarZst(ch)
}

func remoteBackupsConcurrently(hh []Host, config *Config) {
	ch := handleSnapshotsRemote(hh, config)
	ch = handleImages(ch, config.LocalWorkers)
	ch = handleTars(ch, config.LocalWorkers)
	for _, r := range config.ResticRepos {
		ch = backupToRepo(ch, r)
	}
	deleteTarZst(ch)
}

// INFO[0334] Create remote snapshot ssnet                  container=kong-mz host=mbuzi spent=1.705760833s
// INFO[0356] Publish remote snapshot ssnet as image        container=kong-mz host=mbuzi spent=21.809066878s
// INFO[0357] Delete remote snapshot ssnet                  container=kong-mz host=mbuzi spent=1.095481593s

// INFO[0359] Export image as kong-mz.tar                   container=kong-mz host=mbuzi spent=2.551460824s
// INFO[0359] Delete image%!(EXTRA string=kong-mz)          container=kong-mz host=mbuzi spent=153.95267ms

// INFO[0362] Compress kong-mz.tar to kong-mz.tar.zst       container=kong-mz host=mbuzi spent=3.112826626s
// INFO[0363] Delete kong-mz.tar                            container=kong-mz host=mbuzi spent=113.553161ms

// INFO[0364] Backup kong-mz.tar.zst to restic_repos/one    container=kong-mz host=mbuzi spent=1.204783044s
// INFO[0364] Delete kong-mz.tar.zst                        container=kong-mz host=mbuzi spent=53.12294ms

func backupConcurrently(ch chan Container, config *Config, w int) {
	for c := range ch {
		log := log.WithFields(log.Fields{
			"host":      c.Host,
			"container": c.Name,
		})
		wg := sync.WaitGroup{}
		for _, r := range config.ResticRepos {
			wg.Add(1)
			go func(r ResticRepo) {
				defer wg.Done()
				t := time.Now()
				err := r.Backup(fmt.Sprintf("%s.tar.zst", c.Name))
				if err != nil {
					log.Error(err)
				}
				log.WithField("spent", time.Since(t)).Infof("Backup %s.tar.zst to %s", c.Name, r.Path)
			}(r)
		}

		wg.Wait()
		t := time.Now()
		err := c.DeleteImageTarZst()
		if err != nil {
			log.Error(err)
			continue
		}
		log.WithField("spent", time.Since(t)).Infof("Delete %s.tar.zst", c.Name)
	}
}

func handleTars(ch chan Container, w int) chan Container {
	zstCh := make(chan Container, w)

	go func() {
		wg := sync.WaitGroup{}
		for i := 0; i < w; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for c := range ch {
					log := log.WithFields(log.Fields{
						"host":      c.Host,
						"container": c.Name,
					})

					t := time.Now()
					err := c.CompressWithZst()
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
					zstCh <- c
				}
			}()
		}
		wg.Wait()
		close(zstCh)
	}()

	return zstCh
}

func handleImages(ch chan Container, w int) chan Container {
	tarCh := make(chan Container, w)

	go func() {
		wg := sync.WaitGroup{}

		for i := 0; i < w; i++ {
			wg.Add(1)
			go func(ch chan Container) {
				defer wg.Done()
				for c := range ch {
					log := log.WithFields(log.Fields{
						"host":      c.Host,
						"container": c.Name,
					})

					t := time.Now()
					err := c.ExportImage()
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

					tarCh <- c
				}
			}(ch)
		}

		wg.Wait()
		close(tarCh)
	}()

	return tarCh
}

func handleSnapshots(hh []Host, config *Config) chan Container {
	ch := make(chan Container, len(hh))
	go func() {
		wg := sync.WaitGroup{}
		for _, h := range hh {
			wg.Add(1)
			go func(h Host) {
				defer wg.Done()
				log := log.WithField("host", h.Name)
				err := h.GetContainers()
				if err != nil {
					log.Error(err)
				}
				cc := filterContainers(h.Containers, config.Blacklist)
				for _, c := range cc {
					t := time.Now()
					log = log.WithField("container", c.Name)
					if len(c.Snapshots) > 0 {
						err := c.DeleteSnapshotsRemote(h.Name)
						if err != nil {
							log.Error(err)
							continue
						}
						log.WithField("spent", time.Since(t)).Infof("Delete remote snapshots")
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
					ch <- c
				}
			}(h)
		}
		wg.Wait()
		close(ch)

	}()
	return ch
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
