package main

import (
	"flag"
	"log"
	"os"
	"path"

	"github.com/alexferrari88/lazypress"
)

func main() {
	port := flag.Int("port", 3444, "port to listen on")
	flag.Parse()
	// locate chrome executable path
	dir, dirError := os.Getwd()
	if dirError != nil {
		log.Fatalln(dirError)
	}
	chromePath := path.Join(dir, "chrome-linux", "chrome")
	lazypress.InitServer(*port, chromePath)
}
