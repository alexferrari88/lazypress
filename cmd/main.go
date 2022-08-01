package main

import (
	"flag"

	"github.com/alexferrari88/lazypress"
)

func main() {
	port := flag.Int("port", 3444, "port to listen on")
	flag.Parse()
	lazypress.InitServer(*port)
}
