package main

import (
	"flag"
	"log"
	"os"
	"path"

	"github.com/alexferrari88/lazypress"
)

func main() {
	dir, dirError := os.Getwd()
	if dirError != nil {
		log.Fatalln(dirError)
	}
	port := flag.Int("port", 3444, "port to listen on")
	chromePath := flag.String("chrome", path.Join(dir, "chrome-linux", "chrome"), "path to chrome")
	flag.Parse()

	lazypress.InitServer(*port, *chromePath)
}
