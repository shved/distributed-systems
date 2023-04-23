package main

import (
	"os"

	"github.com/shved/distributed-systems/go/pkg/node"
)

type EchoBody struct {
	Type  string  `json:"type"`
	MsgID float64 `json:"msg_id"`
	Echo  string  `json:"echo"`
}

type EchoOkBody struct {
	Type      string  `json:"type"`
	MsgID     uint64  `json:"msg_id"`
	InReplyTo float64 `json:"in_reply_to"`
	Echo      string  `json:"echo"`
}

func main() {
	n := node.NewNode("", 0)

	n.RegisterHandler("echo", func(msg node.Message, msgID uint64) node.Message {
		var req EchoBody

		if errMsg, err := msg.ExtractBody(&req); err != nil {
			return errMsg
		}

		return node.WithOkBody(msg, EchoOkBody{
			Type:      "echo_ok",
			MsgID:     msgID,
			InReplyTo: req.MsgID,
			Echo:      req.Echo,
		})
	})

	n.ListenAndServe(os.Stdin)
}
