package configutil

// TODO: move to common repo

import (
	"flag"
	"log"
)

func GetConfigPath() string {
	configPath := flag.String("config", "", "Path to the config file")
	flag.Parse()

	if *configPath == "" {
		log.Fatal("config path is empty. usage: app -config /path/to/config.yaml")
	}

	return *configPath
}
