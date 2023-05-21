package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/ydssx/api-gen/docs"
	"github.com/ydssx/api-gen/router"
	"gopkg.in/yaml.v2"
)

type Config struct {
	BasicPath string   `yaml:"basicPath"`
	ApiPath   []string `yaml:"apiPath"`
	TypeFile  string   `yaml:"typeFile"`

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
	initRouter()
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

	typeInfo := parseTypes(cfg.TypeFile, cfg.ApiPath[0])

	logicFunc := genLogicFunc(cfg.Logic.File, typeInfo)

	handlerFunc := genHandlerFunc(cfg.Handler.File, cfg.BasicPath, typeInfo, logicFunc)

	err = addRouter(cfg.Router.File, cfg.Router.GroupFunc, typeInfo, handlerFunc)
	if err != nil {
		log.Fatal(err)
	}
}

func initRouter() {
	r := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"
	v1 := r.Group("/api/v1")
	{
		eg := v1.Group("/example")
		{
			eg.GET("/helloworld", Helloworld)
		}
	}

	{
		user := v1.Group("/user")
		router.UserRouter(user)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	r.Run(":8080")
}

// @BasePath /api/v1

// PingExample godoc
// @Summary ping example
// @Schemes
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /example/helloworld [get]
func Helloworld(g *gin.Context) {
	g.JSON(http.StatusOK, "helloworld")
}
