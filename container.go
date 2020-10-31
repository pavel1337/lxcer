package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type Container struct {
	Name       string     `json:"name"`
	StatusCode int        `json:"status_code"`
	Snapshots  []Snapshot `json:"snapshots"`
	Host       string
}

type Snapshot struct {
	Name string `json:"name"`
}

func (c *Container) DeleteSnapshots() error {
	for _, s := range c.Snapshots {
		err := c.DeleteSnapshot(s.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Container) DeleteSnapshotsRemote(host string) error {
	for _, s := range c.Snapshots {
		err := c.DeleteSnapshotRemote(s.Name, host)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Container) SnapshotExists(sn string) bool {
	var snn []string
	for _, s := range c.Snapshots {
		snn = append(snn, s.Name)
	}
	if contains(snn, sn) {
		return true
	}
	return false
}

func (c *Container) Delete() error {
	var (
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)

	cmd := exec.Command("lxc", "delete", "%s", "--force")
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(Stderr.String())
	}
	return nil
}

func (c *Container) DeleteRemote(host string) error {
	var (
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)

	cmd := exec.Command("lxc", "delete", fmt.Sprintf("%s:", host), c.Name)
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(Stderr.String())
	}
	return nil
}

func (c *Container) DeleteSnapshot(sn string) error {
	var (
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)

	cmd := exec.Command("lxc", "delete", fmt.Sprintf("%s/%s", c.Name, sn))
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(Stderr.String())
	}
	return nil
}

func (c *Container) DeleteSnapshotRemote(sn string, host string) error {
	var (
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)
	cmd := exec.Command("lxc", "delete", fmt.Sprintf("%s:%s/%s", host, c.Name, sn))
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(Stderr.String())
	}
	return nil
}

func (c *Container) CreateSnapshotLocal(sn string) error {
	var (
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)
	cmd := exec.Command("lxc", "snapshot", fmt.Sprintf("%s", c.Name), sn)
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(Stderr.String())
	}
	return nil
}

func (c *Container) CreateSnapshotRemote(sn string, host string) error {
	var (
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)
	cmd := exec.Command("lxc", "snapshot", fmt.Sprintf("%s:%s", host, c.Name), sn)
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(Stderr.String())
	}
	return nil
}

func (c *Container) CopySnapshot(sn string, host string) error {
	var (
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)
	cmd := exec.Command("lxc", "copy", fmt.Sprintf("%s:%s/%s", host, c.Name, sn), c.Name)
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(Stderr.String())
	}
	return nil
}

func (c *Container) PublishRemote(sn, host string) error {
	var (
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)
	cmd := exec.Command("lxc", "publish", fmt.Sprintf("%s:%s/%s", host, c.Name, sn), "--alias", c.Name, "--compression", "none")
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(Stderr.String())
	}
	return nil
}

func (c *Container) PublishContainer() error {
	var (
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)
	cmd := exec.Command("lxc", "publish", c.Name, "--alias", c.Name, "--compression", "none")
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(Stderr.String())
	}
	return nil
}

func (c *Container) PublishSnapshot(sn string) error {
	var (
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)
	cmd := exec.Command("lxc", "publish", fmt.Sprintf("%s/%s", c.Name, sn), "--alias", c.Name, "--compression", "none")
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(Stderr.String())
	}
	return nil
}

func (c *Container) ExportImage() error {
	var (
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)
	cmd := exec.Command("lxc", "image", "export", c.Name, c.Name)
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(Stderr.String())
	}
	return nil
}

func (c *Container) DeleteImage() error {
	var (
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)
	cmd := exec.Command("lxc", "image", "delete", c.Name)
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(Stderr.String())
	}
	return nil
}

func (c *Container) CompressWithZst() error {
	var (
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)
	// zstd c1.tar --rsyncable -o c1.tar.zst
	cmd := exec.Command("zstd", fmt.Sprintf("%s.tar", c.Name), "-T0", "--rsyncable", "-o", fmt.Sprintf("%s.tar.zst", c.Name))
	cmd.Stdout = &Stdout
	cmd.Stderr = &Stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(Stderr.String())
	}
	return nil
}

func (c *Container) DeleteImageTar() error {
	return os.Remove(fmt.Sprintf("%s.tar", c.Name)) // remove a single file
}

func (c *Container) DeleteImageTarZst() error {
	return os.Remove(fmt.Sprintf("%s.tar.zst", c.Name)) // remove a single file
}
