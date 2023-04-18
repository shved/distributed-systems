package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync/atomic"

	"github.com/shved/distributed-systems/go/pkg/nodeerr"
)

// Besides few fields Message body could be anything.
type Body map[string]any

type Message struct {
	Src  string `json:"src,omitempty"`
	Dest string `json:"dest,omitempty"`
	Body Body   `json:"body,omitempty"`
}

type HandlerFunc func(msg Message, msgID uint64) Message

type Node struct {
	nodeID string
	msgID  uint64
	nodes  []string

	log      *log.Logger
	output   *log.Logger
	handlers map[string]HandlerFunc
	pool     chan struct{}

	// To use for a specific calls needs.
	ExtraState any
}

func NewNode(nodeID string, msgID uint64) *Node {
	n := &Node{
		nodeID:     nodeID,
		msgID:      msgID,
		nodes:      []string{},
		log:        log.New(os.Stderr, "", 0),
		output:     log.New(os.Stdout, "", 0),
		handlers:   map[string]HandlerFunc{},
		pool:       make(chan struct{}, 200),
		ExtraState: make(map[string]any),
	}

	n.RegisterHandler("init", func(msg Message, msgID uint64) Message {
		n.setID(msg.Body["node_id"].(string))
		// TODO n.setNodes()

		return Message{
			Src:  msg.Dest,
			Dest: msg.Src,
			Body: Body{
				"type":        "init_ok",
				"msg_id":      msgID,
				"in_reply_to": msg.Body["msg_id"].(float64),
			},
		}
	})

	return n
}

func (n *Node) RegisterHandler(method string, fn HandlerFunc) {
	if _, ok := n.handlers[method]; ok {
		panic(fmt.Sprintf("handler for '%s' action defined twice", method))
	}

	n.handlers[method] = fn
}

func (n *Node) setID(id string) {
	if n.nodeID == "" {
		n.nodeID = id
	}
}

func (n *Node) NodeID() string {
	return n.nodeID
}

func (n *Node) setNodes(nodes []string) {
	n.nodes = nodes
}

func (n *Node) incMsgID() uint64 {
	return atomic.AddUint64(&n.msgID, 1)
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

		if n.nodeID != "" && message.Dest != n.nodeID {
			resp := renderError(n.nodeID, message.Src, message.Body["msg_id"].(float64), nodeerr.NodeNotFound)
			n.send(resp)
			return
		}

		msgType := message.Body["type"].(string)
		handler, ok := n.handlers[msgType]
		if !ok {
			resp := renderError(n.nodeID, message.Src, message.Body["msg_id"].(float64), nodeerr.NotSupported)
			n.send(resp)
			return
		}

		newMsgID := n.incMsgID()
		resp := handler(message, newMsgID)
		n.send(resp)
	}(input, readErr)
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
