package utility

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Config struct {
	Parsers struct {
		GeoNode struct {
			Url string `json:"url"`
		} `json:"geoNode"`
	} `json:"parsers"`
	Proxy struct {
		Domain string `json:"domain"`
	} `json:"proxy"`
	Mysql struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Database string `json:"database"`
	} `json:"mysql"`
}

var DefaultConfig *Config

func init() {
	initializeConfigs()
}

func getDataFromYaml(path string) (*Config, error) {
	if buf, err := os.ReadFile(path); err != nil {
		return nil, err
	} else {
		c := new(Config)
		if err = json.Unmarshal(buf, c); err != nil {
			return nil, fmt.Errorf("in file %q: %v", path, err)
		}
		return c, nil
	}
}

func readConfig(path string) (*Config, error) {
	var (
		configData *Config
		errors     error
	)
	if configData, errors = getDataFromYaml(path); errors != nil {
		log.Fatal(errors)
		return nil, errors
	}
	return configData, nil
}

func initializeConfigs() *Config {
	var err error
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	DefaultConfig, _ = readConfig(dir + "/config/config.json")
	if err != nil {
		panic(err)
	}
	return DefaultConfig
}
