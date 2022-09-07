package main

import (
	"guide"
	"os"
)

func main() {
	guide.RunServer(os.Args[1:], os.Stdout)
}
