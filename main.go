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
	fmt.Println("--------------------------------------------------------------------")
	fmt.Println("- when connect to tcp server (default)")
	fmt.Println("   $ mccat [tcp://]URL:PORT (default : localhost:11211)")
	fmt.Println("- when connect to unix socket")
	fmt.Println("   $ mccat [unix://]PATH")
	fmt.Println()
	fmt.Println("  --help [-h]               : show usage")
}

func main() {
	historyFile := os.Getenv("HOME") + "/.mccat_history"

	// connect to memcached server
	fmt.Printf("connect to memcached server [%s]\n", url)

	nc, err := mccat.New(url, historyFile)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("cannot connect to server [%s]: %s\n", url, err.Error()))

		os.Exit(1)
	}
	defer nc.Close(false)

	nc.Start()

	os.Exit(0)
}
