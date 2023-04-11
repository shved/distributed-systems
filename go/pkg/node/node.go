package node

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync/atomic"

	"github.com/shved/distributed-systems/go/pkg/nodeerr"
)

const (
	Init = "init"
	Echo = "echo"
)

// Besides few fields Message body could be anything.
type Body map[string]any

type Message struct {
	Src  string `json:"src,omitempty"`
	Dest string `json:"dest,omitempty"`
	Body Body   `json:"body,omitempty"`
}

type HandlerFunc func(msg Message) Message

type Node struct {
	id    string
	msgID uint64

	log      *log.Logger
	output   *log.Logger
	handlers map[string]HandlerFunc
	pool     chan struct{}
}

func NewNode() *Node {
	return &Node{
		log:      log.New(os.Stderr, "", 0),
		output:   log.New(os.Stdout, "", 0),
		handlers: map[string]HandlerFunc{},
		pool:     make(chan struct{}, 100),
	}
}

func (n *Node) RegisterHandler(method string, fn HandlerFunc) {
	n.handlers[method] = fn
}

func (n *Node) setID(id string) {
	if n.id == "" {
		n.id = id
	}
}

func (n *Node) MsgID() uint64 {
	return n.msgID
}

func (n *Node) incMsgID() {
	atomic.AddUint64(&n.msgID, 1)
}

func (n *Node) SpawnHandler(input string, readErr error) {
	n.pool <- struct{}{}

	go func(input string, readErr error) {
		defer func() {
			<-n.pool
		}()

		n.log.Printf("Received %s", input)

		if readErr != nil {
			resp := renderError("", "", 0, nodeerr.Malformed)
			n.send(resp)
			return
		}

		message, err := parseMessage(input)
		if err != nil {
			resp := renderError("", "", 0, nodeerr.Malformed)
			n.send(resp)
			return
		}

		if n.id != "" && message.Dest != n.id {
			resp := renderError(n.id, message.Src, message.Body["msg_id"].(float64), nodeerr.NodeNotFound)
			n.send(resp)
			return
		}

		msgType := message.Body["type"].(string)

		if msgType == Init {
			resp := n.init(message)
			n.send(resp)
			return
		}

		handler, ok := n.handlers[msgType]
		if !ok {
			resp := renderError(n.id, message.Src, message.Body["msg_id"].(float64), nodeerr.NotSupported)
			n.send(resp)
			return
		}

		n.incMsgID()
		resp := handler(message)
		n.send(resp)
	}(input, readErr)
}

func (n *Node) echo(in Message) Message {
	n.incMsgID()

	return Message{
		Src:  in.Dest,
		Dest: in.Src,
		Body: Body{
			"type":        "echo_ok",
			"msg_id":      n.msgID,
			"in_reply_to": in.Body["msg_id"].(float64),
			"echo":        in.Body["echo"],
		},
	}
}

func (n *Node) init(in Message) Message {
	n.setID(in.Body["node_id"].(string))
	n.incMsgID()

	return Message{
		Src:  in.Dest,
		Dest: in.Src,
		Body: Body{
			"type":        "init_ok",
			"msg_id":      n.msgID,
			"in_reply_to": in.Body["msg_id"].(float64),
		},
	}
}

func (n *Node) send(message any) {
	// TODO Could be some premarshaled error stub added to send. Just to not skip the error here.
	resp, _ := json.Marshal(message)
	n.output.Println(string(resp))
	n.log.Printf("Sent %#v\n", message)
}

func parseMessage(input string) (Message, error) {
	var msg Message
	if err := json.Unmarshal([]byte(input), &msg); err != nil {
		return Message{}, errors.New("invalid json")
	}

	// TODO Legitimately validate message here before pass it further.

	return msg, nil
}

func renderError(from, to string, inReply float64, code nodeerr.ErrorCode) Message {
	return Message{
		Src:  from,
		Dest: to,
		Body: Body{
			"type":        "error",
			"in_reply_to": inReply,
			"code":        code,
			"text":        code.String(),
		},
	}
}
