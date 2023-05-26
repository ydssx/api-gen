package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
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

type APIGenBuilder struct {
	cfg         Config
	typeInfo    TypeInfo
	logicFunc   FuncInfo
	handlerFunc FuncInfo
	api         string
}

func NewAPIGenBuilder() *APIGenBuilder {
	return &APIGenBuilder{}
}

func (b *APIGenBuilder) WithConfig(configFile string) *APIGenBuilder {
	configData, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	if err := yaml.Unmarshal(configData, &b.cfg); err != nil {
		log.Fatalf("failed to parse config file: %v", err)
	}

	return b
}

func (b *APIGenBuilder) WithTypeInfo(typeFile, apiPath string) *APIGenBuilder {
	b.typeInfo = parseTypes(typeFile, apiPath)
	return b
}

func (b *APIGenBuilder) WithLogicFunc(logicFile string) *APIGenBuilder {
	b.logicFunc = genLogicFunc(logicFile, b.typeInfo)
	return b
}

func (b *APIGenBuilder) WithHandlerFunc(handlerFile string) *APIGenBuilder {
	b.handlerFunc = genHandlerFunc(handlerFile, b.typeInfo, b.logicFunc, b.cfg)
	return b
}

func (b *APIGenBuilder) AddRouter(routerFile, groupFunc string) error {
	return addRouter(routerFile, groupFunc, b.typeInfo, b.handlerFunc)
}

func (b *APIGenBuilder) Build() {
	for _, api := range b.cfg.ApiPath {
		err := b.WithTypeInfo(b.cfg.TypeFile, api).
			WithLogicFunc(b.cfg.Logic.File).
			WithHandlerFunc(b.cfg.Handler.File).
			AddRouter(b.cfg.Router.File, b.cfg.Router.GroupFunc)
		if err != nil {
			log.Fatal(err)
		}
	}
}
