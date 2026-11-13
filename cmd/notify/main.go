package main

import (
	"flag"

	"github.com/vishenosik/notify/internal/config"
)

func main() {
	config.Flags()
	flag.Parse()

	New().Run()
}
