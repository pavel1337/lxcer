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

func (c *Container) Delete() error {
	var (
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)

	cmd := exec.Command("lxc", "delete", fmt.Sprintf("%s", c.Name), "--force")
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

	cmd := exec.Command("lxc", "delete", fmt.Sprintf("%s:", host), fmt.Sprintf("%s", c.Name), "--force")
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

	cmd := exec.Command("lxc", "delete", fmt.Sprintf("%s:", host), fmt.Sprintf("%s/%s", c.Name, sn))
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

func (c *Container) Publish() error {
	var (
		Stdout bytes.Buffer
		Stderr bytes.Buffer
	)
	cmd := exec.Command("lxc", "publish", fmt.Sprintf("%s", c.Name), "--alias", c.Name, "--compression", "none", "--force")
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
	cmd := exec.Command("lxc", "image", "export", fmt.Sprintf("%s", c.Name), fmt.Sprintf("%s", c.Name))
	fmt.Println(cmd.String())
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
	cmd := exec.Command("lxc", "image", "delete", fmt.Sprintf("%s", c.Name))
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
	cmd := exec.Command("zstd", fmt.Sprintf("%s.tar", c.Name), "--rsyncable", "-o", fmt.Sprintf("%s.tar.zst", c.Name))
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
