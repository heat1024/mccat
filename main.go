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
	// connect to memcached server
	fmt.Printf("connect to memcached server [%s]\n", url)

	nc, err := mccat.New(url)
	if err != nil {
		fmt.Printf("cannot connect to server [%s]\n", url)

		os.Exit(1)
	}
	defer nc.Close()

	nc.Start(url)

	os.Exit(0)
}
