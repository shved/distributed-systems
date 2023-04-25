package main

import (
	"os"

	"github.com/shved/distributed-systems/go/pkg/node"
	"github.com/shved/distributed-systems/go/pkg/noderr"
)

type EchoBody struct {
	Type  string  `json:"type,omitemtpy"`
	MsgID float64 `json:"msg_id,omitemtpy"`
	Echo  string  `json:"echo,omitemtpy"`
}

type EchoOkBody struct {
	Type      string  `json:"type,omitemtpy"`
	MsgID     uint64  `json:"msg_id,omitemtpy"`
	InReplyTo float64 `json:"in_reply_to,omitemtpy"`
	Echo      string  `json:"echo,omitemtpy"`
}

func main() {
	n := node.New("", 0)

	n.RegisterHandler("echo", func(msg node.Message, msgID uint64) *node.Message {
		var body EchoBody

		if err := msg.ExtractBody(&body); err != nil {
			return node.WithErrorBody(msg, 0, noderr.Malformed)
		}

		return node.WithOkBody(msg, EchoOkBody{
			Type:      "echo_ok",
			MsgID:     msgID,
			InReplyTo: body.MsgID,
			Echo:      body.Echo,
		})
	})

	n.ListenAndServe(os.Stdin)
}
