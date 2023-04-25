package main

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/shved/distributed-systems/go/pkg/node"
	"github.com/shved/distributed-systems/go/pkg/noderr"
)

type TopologyBody struct {
	Type     string              `json:"type,omitemtpy"`
	MsgID    float64             `json:"msg_id,omitemtpy"`
	Topology map[string][]string `json:"topology,omitemtpy"`
}

type TopologyOkBody struct {
	Type      string  `json:"type,omitemtpy"`
	MsgID     uint64  `json:"msg_id,omitemtpy"`
	InReplyTo float64 `json:"in_reply_to,omitemtpy"`
}

type BroadcastBody struct {
	Type    string  `json:"type,omitemtpy"`
	MsgID   float64 `json:"msg_id,omitemtpy"`
	Message int     `json:"message,omitemtpy"`
}

type BroadcastOkBody struct {
	Type      string  `json:"type,omitemtpy"`
	MsgID     uint64  `json:"msg_id,omitemtpy"`
	InReplyTo float64 `json:"in_reply_to,omitemtpy"`
}

type ReadBody struct {
	Type  string  `json:"type,omitemtpy"`
	MsgID float64 `json:"msg_id,omitemtpy"`
}

type ReadOkBody struct {
	Type      string  `json:"type,omitemtpy"`
	MsgID     uint64  `json:"msg_id,omitemtpy"`
	InReplyTo float64 `json:"in_reply_to,omitemtpy"`
	Messages  []int   `json:"messages,omitemtpy"`
}

type extraState struct {
	mu             sync.Mutex
	topology       map[string][]string
	neighbours     []string
	messagesHashed map[int]struct{}
	messages       []int
}

func main() {
	n := node.New("", 0)

	state := &extraState{
		neighbours:     n.Nodes(),
		messagesHashed: make(map[int]struct{}),
		messages:       make([]int, 0),
		topology:       make(map[string][]string),
	}

	n.RegisterHandler("topology", func(msg node.Message, msgID uint64) *node.Message {
		var body TopologyBody

		if err := msg.ExtractBody(&body); err != nil {
			return node.WithErrorBody(msg, 0, noderr.Malformed)
		}

		// state.setTopology(body.Topology, n)

		return node.WithOkBody(msg, TopologyOkBody{
			Type:      "topology_ok",
			MsgID:     msgID,
			InReplyTo: body.MsgID,
		})
	})

	n.RegisterHandler("broadcast", func(msg node.Message, msgID uint64) *node.Message {
		var body BroadcastBody

		if err := msg.ExtractBody(&body); err != nil {
			return node.WithErrorBody(msg, 0, noderr.Malformed)
		}

		if body.MsgID == 0 {
			return nil
		}

		state.appendKnownMessages(body.Message)

		broadcast(n, msg.Src, body.Message, state.neighbours)

		return node.WithOkBody(msg, BroadcastOkBody{
			Type:      "broadcast_ok",
			MsgID:     msgID,
			InReplyTo: body.MsgID,
		})
	})

	n.RegisterHandler("read", func(msg node.Message, msgID uint64) *node.Message {
		var body ReadBody

		if err := msg.ExtractBody(&body); err != nil {
			return node.WithErrorBody(msg, 0, noderr.Malformed)
		}

		return node.WithOkBody(msg, ReadOkBody{
			Type:      "read_ok",
			MsgID:     msgID,
			InReplyTo: body.MsgID,
			Messages:  state.knownMessages(),
		})

	})

	n.ListenAndServe(os.Stdin)
}

// Send a bunch of fire-and-forget type of messages that are just printed to stdout.
func broadcast(n *node.Node, src string, message int, neighbours []string) {
	body := BroadcastBody{
		Type:    "broadcast",
		Message: message,
	}

	rawBody, _ := json.Marshal(body)

	msg := &node.Message{
		Src:  n.NodeID(),
		Body: rawBody,
	}

	for _, neighbour := range neighbours {
		if neighbour != src { // Skip send the message back to src.
			msg.Dest = neighbour
			n.Send(msg)
		}

	}
}

func (s *extraState) setTopology(t map[string][]string, n *node.Node) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.topology = t
	if neighbours, ok := t[n.NodeID()]; ok {
		s.neighbours = neighbours
	}
}

func (s *extraState) appendKnownMessages(message int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.messagesHashed[message]; !ok {
		s.messagesHashed[message] = struct{}{}
		s.messages = append(s.messages, message)
	}
}

func (s *extraState) knownMessages() []int {
	return s.messages
}
