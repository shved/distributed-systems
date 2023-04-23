package node

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync/atomic"

	"github.com/shved/distributed-systems/go/pkg/noderr"
)

type Message struct {
	Src  string          `json:"src,omitempty"`
	Dest string          `json:"dest,omitempty"`
	Body json.RawMessage `json:"body,omitempty"`
}

type MsgProbe struct {
	Body struct {
		Type  string  `json:"type,omitempy"`
		MsgID float64 `json:"msg_id,omitempy"`
	} `json:"body,omitempy"`
}

type InitBody struct {
	Type    string   `json:"type,omitempy"`
	MsgID   uint64   `json:"msg_id,omitempy"`
	NodeID  string   `json:"node_id,omitempy"`
	NodeIDs []string `json:"node_ids,omitempy"`
}

type InitOkBody struct {
	Type      string `json:"type,omitempy"`
	MsgID     uint64 `json:"msg_id,omitempy"`
	InReplyTo uint64 `json:"in_reply_to,omitempy"`
}

type HandlerFunc func(msg Message, msgID uint64) Message

type Node struct {
	nodeID string
	msgID  uint64
	nodes  []string

	log      *log.Logger
	output   io.Writer // Should be syncronised.
	handlers map[string]HandlerFunc
	pool     chan struct{}
}

func New(nodeID string, msgID uint64) *Node {
	n := &Node{
		nodeID:   nodeID,
		msgID:    msgID,
		nodes:    []string{},
		log:      log.New(os.Stderr, "", 0),
		output:   log.New(os.Stdout, "", 0).Writer(),
		handlers: map[string]HandlerFunc{},
		pool:     make(chan struct{}, 100),
	}

	n.RegisterHandler("init", func(msg Message, msgID uint64) Message {
		var body InitBody

		if err := msg.ExtractBody(&body); err != nil {
			return WithErrorBody(msg, 0, noderr.Malformed)
		}

		n.setID(body.NodeID)
		n.nodes = body.NodeIDs

		return WithOkBody(msg, InitOkBody{
			Type:      "init_ok",
			MsgID:     msgID,
			InReplyTo: body.MsgID,
		})
	})

	return n
}

func (n *Node) RegisterHandler(method string, fn HandlerFunc) {
	if _, ok := n.handlers[method]; ok {
		panic(fmt.Sprintf("handler for '%s' action defined twice", method))
	}

	n.handlers[method] = fn
}

func (n *Node) ListenAndServe(input *os.File) {
	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		n.SpawnHandler(input, err)
	}
}

func (n *Node) SpawnHandler(input string, readErr error) {
	n.pool <- struct{}{}

	go func(input string, readErr error) {
		defer func() {
			<-n.pool
		}()

		n.log.Printf("Received %s", input)

		if readErr != nil {
			resp := WithErrorBody(Message{}, 0, noderr.Malformed)
			n.Send(resp)
			return
		}

		message, probe, err := parseMessage(input)
		if err != nil {
			resp := WithErrorBody(Message{}, 0, noderr.Malformed)
			n.Send(resp)
			return
		}

		if n.nodeID != "" && message.Dest != n.nodeID {
			resp := WithErrorBody(message, 0, noderr.Malformed)
			n.Send(resp)
			return
		}

		handler, ok := n.handlers[probe.Body.Type]
		if !ok {
			resp := WithErrorBody(message, probe.Body.MsgID, noderr.NotSupported)
			n.Send(resp)
			return
		}

		newMsgID := n.incMsgID()
		resp := handler(message, newMsgID)
		n.Send(resp)
	}(input, readErr)
}

func (n *Node) setID(id string) {
	if n.nodeID == "" {
		n.nodeID = id
	}
}

func (n *Node) NodeID() string {
	return n.nodeID
}

func (n *Node) Nodes() []string {
	return n.nodes
}

func (n *Node) setNodes(nodes []string) {
	n.nodes = nodes
}

func (n *Node) incMsgID() uint64 {
	return atomic.AddUint64(&n.msgID, 1)
}

func (n *Node) Send(message Message) {
	// TODO Could be some premarshaled error stub added to send. Just to not skip the error here.
	resp, _ := json.Marshal(message)
	resp = append(resp, '\n')
	n.output.Write(resp)
	n.log.Printf("Sent %#v\n", message)
}

func parseMessage(input string) (Message, MsgProbe, error) {
	var msg Message
	if err := json.Unmarshal([]byte(input), &msg); err != nil {
		return Message{}, MsgProbe{}, errors.New("invalid json")
	}

	var msgProbe MsgProbe
	if err := json.Unmarshal([]byte(input), &msgProbe); err != nil {
		return Message{}, MsgProbe{}, errors.New("invalid json")
	}

	return msg, msgProbe, nil
}

func (m *Message) ExtractBody(body any) error {
	if err := json.Unmarshal(m.Body, body); err != nil {
		return err
	}

	return nil
}

func WithErrorBody(msg Message, inReply float64, code noderr.ErrorCode) Message {
	errBody := noderr.ErrorBody{
		Type:      "error",
		InReplyTo: inReply,
		Code:      code,
		Text:      code.String(),
	}

	body, _ := json.Marshal(errBody)

	return Message{
		Src:  msg.Dest,
		Dest: msg.Src,
		Body: body,
	}
}

func WithOkBody(msg Message, body any) Message {
	resp, _ := json.Marshal(body)

	return Message{
		Src:  msg.Dest,
		Dest: msg.Src,
		Body: resp,
	}
}
