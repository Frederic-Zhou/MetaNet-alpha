package main

import (
	"bufio"
	"bytes"
	"os"

	"github.com/Frederic-Zhou/MetaNet-alpha/NAT2P/client/network"
)

func main() {

	//network.Register("ZETA", "0.0.0.0:9999", "0.0.0.0:9998")

	go network.Listener("0.0.0.0:9999")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {

		network.Sender(bytes.NewReader(scanner.Bytes()),
			network.DataType_Text,
			"0.0.0.0:9999",
			"0.0.0.0:9998")
	}

}
