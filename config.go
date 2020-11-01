package main

import (
	"flag"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Hosts        []string     `yaml:"hosts"`
	Blacklist    []string     `yaml:"blacklist"`
	ResticRepos  []ResticRepo `yaml:"restic_repos"`
	LocalWorkers int          `yaml:"local_workers"`
	// TmpFolder   string       `yaml:"tmp_folder"`
	Cleanup      bool
	Concurrently bool
	Local        bool
}

var (
	fileName     = flag.String("config", "", "Path to YAML config.")
	cleanup      = flag.Bool("cleanup", false, "Delete local containers and images before processing")
	concurrently = flag.Bool("concurrently", false, "Backup concurrently")
	local        = flag.Bool("local", false, "Backup local containers")
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
	})
}

func loadConfig() *Config {
	flag.Parse()

	if *fileName == "" {
		log.Fatalf(`Please provide path to yaml config file by using "-config" flag`)
	}

	c := readConfig(*fileName)
	c.Cleanup = *cleanup
	c.Local = *local
	c.Concurrently = *concurrently

	return &c
}

func createFolder(path string) error {
	return os.MkdirAll(path, 775)
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
