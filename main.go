package main

import "github.com/Frederic-Zhou/MetaNet-alpha/node"

func main() {
	node.InitFiles("validator")
	node.Start()
}
