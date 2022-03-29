package main

import (
	"bufio"
	"bytes"
	"flag"
	"os"

	"github.com/Frederic-Zhou/MetaNet-alpha/NAT2P/network"
)

func main() {
	laddr := flag.String("l", "0.0.0.0:9999", "0.0.0.0:9999")
	raddr := flag.String("r", "0.0.0.0:9998", "0.0.0.0:9998")

	flag.Parse()

	go network.Listener(*laddr)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {

		network.Sender(
			bytes.NewReader(scanner.Bytes()),
			network.DataType_Text, *laddr, *raddr)
	}

}
