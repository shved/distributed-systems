package main

import (
	"fmt"
	"os"

	"github.com/shved/distributed-systems/go/pkg/node"
	"github.com/shved/distributed-systems/go/pkg/noderr"
)

type GenerateBody struct {
	Type  string  `json:"type,omitempty"`
	MsgID float64 `json:"msg_id,omitempty"`
}

type GenerateOkBody struct {
	Type      string  `json:"type,omitempty"`
	MsgID     uint64  `json:"msg_id,omitempty"`
	InReplyTo float64 `json:"in_reply_to,omitempty"`
	ID        string  `json:"id,omitempty"`
}

func main() {
	n := node.New("", 0)

	n.RegisterHandler("generate", func(msg node.Message, msgID uint64) *node.Message {
		var body GenerateBody

		if err := msg.ExtractBody(&body); err != nil {
			return node.WithErrorBody(msg, 0, noderr.Malformed)
		}

		return node.WithOkBody(msg, GenerateOkBody{
			Type:      "generate_ok",
			MsgID:     msgID,
			InReplyTo: body.MsgID,
			ID:        fmt.Sprintf("%s-%d", n.NodeID(), msgID),
		})
	})

	n.ListenAndServe(os.Stdin)
}
