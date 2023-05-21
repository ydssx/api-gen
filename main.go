package main

import (
	"flag"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Group    string   `yaml:"group"`
	ApiPath  []string `yaml:"apiPath"`
	TypeFile string   `yaml:"typeFile"`

	Logic struct {
		File     string `yaml:"file"`
		Receiver string `yaml:"receiver"`
	} `yaml:"logic"`

	Handler struct {
		File string `yaml:"file"`
	} `yaml:"handler"`

	Router struct {
		File      string `yaml:"file"`
		GroupFunc string `yaml:"groupFunc"`
	} `yaml:"router"`
}

func main() {
	// initRouter()
	var configFile string
	flag.StringVar(&configFile, "c", "config.yaml", "path to config file")
	flag.Parse()

	configData, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(configData, &cfg); err != nil {
		log.Fatalf("failed to parse config file: %v", err)
	}
	
	for _, api := range cfg.ApiPath {
		typeInfo := parseTypes(cfg.TypeFile, api)

		logicFunc := genLogicFunc(cfg.Logic.File, typeInfo)

		handlerFunc := genHandlerFunc(cfg.Handler.File, typeInfo, logicFunc)

		err = addRouter(cfg.Router.File, cfg.Router.GroupFunc, typeInfo, handlerFunc)
		if err != nil {
			log.Fatal(err)
		}

	}
}
