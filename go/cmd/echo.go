package main

import (
	"bufio"
	"os"

	"github.com/shved/distributed-systems/go/pkg/node"
)

func main() {
	n := node.NewNode()
	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		n.SpawnHandler(input, err)
	}
}
