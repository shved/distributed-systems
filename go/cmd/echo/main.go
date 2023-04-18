package main

import (
	"bufio"
	"os"

	"github.com/shved/distributed-systems/go/pkg/node"
)

func main() {
	n := node.NewNode("", 0)

	n.RegisterHandler("echo", func(msg node.Message, msgID uint64) node.Message {
		return node.Message{
			Src:  msg.Dest,
			Dest: msg.Src,
			Body: node.Body{
				"type":        "echo_ok",
				"msg_id":      msgID,
				"in_reply_to": msg.Body["msg_id"].(float64),
				"echo":        msg.Body["echo"],
			},
		}
	})

	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		n.SpawnHandler(input, err)
	}
}
