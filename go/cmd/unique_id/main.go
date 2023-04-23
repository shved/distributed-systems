package main

import (
	"fmt"
	"os"

	"github.com/shved/distributed-systems/go/pkg/node"
)

type GenerateBody struct {
	Type  string  `json:"type"`
	MsgID float64 `json:"msg_id"`
}

type GenerateOkBody struct {
	Type      string  `json:"type"`
	MsgID     uint64  `json:"msg_id"`
	InReplyTo float64 `json:"in_reply_to"`
	ID        string  `json:"id"`
}

func main() {
	n := node.NewNode("", 0)

	n.RegisterHandler("generate", func(msg node.Message, msgID uint64) node.Message {
		var req GenerateBody

		if errMsg, err := msg.ExtractBody(&req); err != nil {
			return errMsg
		}

		return node.WithOkBody(msg, GenerateOkBody{
			Type:      "generate_ok",
			MsgID:     msgID,
			InReplyTo: req.MsgID,
			ID:        fmt.Sprintf("%s-%d", n.NodeID(), msgID),
		})
	})

	n.ListenAndServe(os.Stdin)
}
