package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Hosts             []string     `yaml:"hosts"`
	Blacklist         []string     `yaml:"blacklist"`
	BackupResticRepos []ResticRepo `yaml:"backup_restic_repos"`
	RestoreResticRepo ResticRepo   `yaml:"restore_restic_repo"`
	LocalWorkers      int          `yaml:"local_workers"`
	ActionType        string
	Cleanup           bool
	Concurrently      bool
	Local             bool
	ContList          contList
}

type RestoreContainer struct {
	Name        string
	RestoreName string
}

type contList map[string]string

var (
	logLevel           = flag.String("log-level", "error", "Remote host name to restore containers to")
	remoteHost         = flag.String("remote-host", "", "Remote host name to restore containers to")
	restoreContainerAs = flag.String("as", "", "Restore-name of the container")
	restoreContainer   = flag.String("container", "", "Name of the container to restore")
	restoreList        = flag.String("restore-list", "", "Path to list in format container_name:container_restore_name")
	fileName           = flag.String("config", "", "Path to YAML config.")
	actionType         = flag.String("a", "", "Action to take (backup or restore)")
	cleanup            = flag.Bool("cleanup", false, "Delete local containers and images before processing")
	concurrently       = flag.Bool("concurrently", false, "Backup concurrently")
	local              = flag.Bool("local", false, "Backup local containers")
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
	})
}

func loadConfig() *Config {
	flag.Parse()

	lv, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalln(err)
	}
	log.SetLevel(lv)

	if *fileName == "" {
		log.Fatalf(`Please provide path to yaml config file by using -config flag`)
	}

	if *actionType == "" {
		log.Fatalf(`Please provide action type with -a flag (backup or restore)`)
	}

	c := readConfig(*fileName)

	if *restoreList != "" {
		cl, err := LoadContainerList(*restoreList)
		if err != nil {
			log.Fatalln(err)
		}
		c.ContList = cl
	}

	c.ActionType = *actionType
	c.Cleanup = *cleanup
	c.Local = *local
	c.Concurrently = *concurrently

	return &c
}

func readConfig(path string) Config {
	yamlFile, err := ioutil.ReadFile(*fileName)
	if err != nil {
		log.Fatalf("Error reading YAML file: %s", err)
	}

	var c Config
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Error parsing YAML file: %s", err)
	}

	return c
}

func LoadContainerList(path string) (contList, error) {
	var cl contList = make(map[string]string)

	buf, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	snl := bufio.NewScanner(buf)
	for snl.Scan() {
		ss := strings.Split(snl.Text(), ":")
		if len(ss) == 2 {
			cl[ss[0]] = ss[1]
		}
	}
	err = snl.Err()
	if err != nil {
		return nil, err
	}

	if err = buf.Close(); err != nil {
		return nil, err
	}
	return cl, nil
}
