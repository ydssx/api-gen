package main

import (
	"flag"

	"github.com/ydssx/api-gen/gen"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "c", "config.yaml", "path to config file")
	flag.Parse()

	gen.NewAPIGenBuilder().WithConfig(configFile).Build()
}
