package main

import (
	"flag"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Hosts       []string     `yaml:"hosts"`
	Blacklist   []string     `yaml:"blacklist"`
	ResticRepos []ResticRepo `yaml:"restic_repos"`
	// TmpFolder   string       `yaml:"tmp_folder"`
	Cleanup bool
	Local   bool
}

func loadConfig() *Config {
	fileName := flag.String("config", "", "Path to YAML config.")
	cleanup := flag.Bool("cleanup", false, "Delete local containers and images before processing")
	local := flag.Bool("local", false, "Backup local containers")

	flag.Parse()

	if *fileName == "" {
		log.Fatalf(`Please provide path to yaml config file by using "-config" flag`)
	}

	yamlFile, err := ioutil.ReadFile(*fileName)
	if err != nil {
		log.Fatalf("Error reading YAML file: %s", err)
	}

	var c Config
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Error parsing YAML file: %s", err)
	}

	// if c.TmpFolder == "" {
	// 	log.Fatalf(`Please provide "tmp_folder" in config`)
	// }

	c.Cleanup = *cleanup
	c.Local = *local
	// err = createFolder(c.TmpFolder)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	return &c
}

func createFolder(path string) error {
	return os.MkdirAll(path, 775)
}
