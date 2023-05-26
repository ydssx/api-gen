package main

import (
	"flag"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// 定义处理链的接口
type HandlerChain interface {
	SetNext(HandlerChain) HandlerChain // 设置下一个处理者
	Handle(*APIGenBuilder)             // 处理请求
}

// 解析类型的处理者
type ParseTypesHandler struct {
	next HandlerChain
}

func (h *ParseTypesHandler) SetNext(next HandlerChain) HandlerChain {
	h.next = next
	return next
}

func (h *ParseTypesHandler) Handle(data *APIGenBuilder) {
	// 解析类型的逻辑
	data.typeInfo = parseTypes(data.cfg.TypeFile, data.api)

	// 调用下一个处理者
	if h.next != nil {
		h.next.Handle(data)
	}
}

// 生成逻辑函数的处理者
type GenLogicFuncHandler struct {
	next HandlerChain
}

func (h *GenLogicFuncHandler) SetNext(next HandlerChain) HandlerChain {
	h.next = next
	return next
}

func (h *GenLogicFuncHandler) Handle(data *APIGenBuilder) {
	// 生成逻辑函数的逻辑
	data.logicFunc = genLogicFunc(data.cfg.Logic.File, data.typeInfo)

	// 调用下一个处理者
	if h.next != nil {
		h.next.Handle(data)
	}
}

// 生成处理函数的处理者
type GenHandlerFuncHandler struct {
	next HandlerChain
}

func (h *GenHandlerFuncHandler) SetNext(next HandlerChain) HandlerChain {
	h.next = next
	return next
}

func (h *GenHandlerFuncHandler) Handle(data *APIGenBuilder) {
	// 生成处理函数的逻辑
	data.handlerFunc = genHandlerFunc(data.cfg.Handler.File, data.typeInfo, data.logicFunc, data.cfg)

	// 调用下一个处理者
	if h.next != nil {
		h.next.Handle(data)
	}
}

// 添加路由的处理者
type AddRouterHandler struct {
	next HandlerChain
}

func (h *AddRouterHandler) SetNext(next HandlerChain) HandlerChain {
	h.next = next
	return next
}

func (h *AddRouterHandler) Handle(data *APIGenBuilder) {
	// 添加路由的逻辑
	err := addRouter(data.cfg.Router.File, data.cfg.Router.GroupFunc, data.typeInfo, data.handlerFunc)
	if err != nil {
		log.Fatal(err)
	}

	// 不需要调用下一个处理者，这是处理链的最后一个处理者
}

func loadConfig() Config {
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
	return cfg
}

func main1() {
	cfg := loadConfig()

	// 创建处理链
	parseTypesHandler := &ParseTypesHandler{}
	genLogicFuncHandler := &GenLogicFuncHandler{}
	genHandlerFuncHandler := &GenHandlerFuncHandler{}
	addRouterHandler := &AddRouterHandler{}

	// 设置处理链的顺序
	parseTypesHandler.SetNext(genLogicFuncHandler).SetNext(genHandlerFuncHandler).SetNext(addRouterHandler)

	for _, api := range cfg.ApiPath {
		// 调用处理链的头部处理者
		bd := &APIGenBuilder{cfg: cfg, api: api}
		parseTypesHandler.Handle(bd)
	}
}
