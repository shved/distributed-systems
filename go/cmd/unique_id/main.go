package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/shved/distributed-systems/go/pkg/node"
)

func main() {
	n := node.NewNode("", 0)

	n.RegisterHandler("generate", func(msg node.Message, msgID uint64) node.Message {
		return node.Message{
			Src:  msg.Dest,
			Dest: msg.Src,
			Body: node.Body{
				"type":        "generate_ok",
				"msg_id":      msgID,
				"in_reply_to": msg.Body["msg_id"].(float64),
				"id":          fmt.Sprintf("%s-%d", n.NodeID(), msgID),
			},
		}
	})

	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		n.SpawnHandler(input, err)
	}
}
