package main

import (
	"flag"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "c", "config.yaml", "path to config file")
	flag.Parse()

	NewAPIGenBuilder().WithConfig(configFile).Build()
}
