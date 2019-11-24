package main

import (
	"fmt"
	"os"

	mccat "github.com/heat1024/mccat/memcache-cat"
)

var url string

func init() {
	url = "localhost:11211"

	if len(os.Args) == 2 {
		url = os.Args[1]
	}

	if url == "help" || url == "-h" {
		Usage()
		os.Exit(0)
	}
}

// Usage show simple manual
func Usage() {
	fmt.Println("How to use mccat(memcached cat)")
	fmt.Println("mccat [URL:PORT] (default server : localhost:11211)")
	fmt.Println()
	fmt.Println("  --help [-h]               : show usage")
}

func main() {
	mccat.Start(url)

	os.Exit(0)
}
