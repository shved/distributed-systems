package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"os"
)

type ErrorCode int

const (
	Timeout            ErrorCode = 0  // "Timed out"
	NodeNotFound       ErrorCode = 1  // "Node not found"
	NotSupported       ErrorCode = 10 // "Operation not supported"
	Unavailable        ErrorCode = 11 // "Temporary unavailable"
	Malformed          ErrorCode = 12 // "Malformed request"
	Crashed            ErrorCode = 13 // "Crashed"
	Aborted            ErrorCode = 14 // "Aborted"
	KeyNotExist        ErrorCode = 20 // "Key does not exist"
	KeyAlreadyExist    ErrorCode = 21 // "Key already exist"
	PreconditionFailed ErrorCode = 22 // "Precondition failed"
	TXConflict         ErrorCode = 30 // "Transaction conflict"
)

// Besides few fields Message body could be anything.
type RawMessage map[string]any

type Message struct {
	Src  string     `json:"src"`
	Dest string     `json:"dest"`
	Body RawMessage `json:"body"`
}

type ErrorMessage struct {
	Src  string    `json:"src"`
	Dest string    `json:"dest"`
	Body ErrorBody `json:"body"`
}

type ErrorBody struct {
	Type      string    `json:"type"`
	InReplyTo float64   `json:"in_reply_to"`
	Code      ErrorCode `json:"code"`
	Text      string    `json:"text"`
}

type Node struct {
	id       string
	msgIDSeq int
	log      *log.Logger
	output   *log.Logger
}

func (n *Node) setID(id string) {
	if n.id != "" {
		return
	}

	n.id = id
}

func (n *Node) spawnHandler(input string, readErr error) {
	go func(input string, readErr error) {
		n.log.Println("Received", input)

		if readErr != nil {
			resp := renderError("", "", 0, Malformed)
			n.send(resp)
			return
		}

		message, err := parseMessage(input)
		if err != nil {
			resp := renderError("", "", 0, Malformed)
			n.send(resp)
			return
		}

		switch message.Body["type"] {
		case "init":
			resp := n.handleInit(message)
			n.send(resp)
		case "echo":
			if message.Dest != n.id {
				resp := renderError(n.id, message.Src, message.Body["msg_id"].(float64), NodeNotFound)
				n.send(resp)
				return
			}

			resp := n.handleEcho(message)
			n.send(resp)
		default:
			resp := renderError(n.id, message.Src, message.Body["msg_id"].(float64), NotSupported)
			n.send(resp)
		}
	}(input, readErr)
}

func renderError(from, to string, inReply float64, code ErrorCode) ErrorMessage {
	return ErrorMessage{
		Src:  from,
		Dest: to,
		Body: ErrorBody{
			Type:      "error",
			InReplyTo: inReply,
			Code:      code,
			Text:      code.String(),
		},
	}
}

func (n *Node) handleEcho(in Message) Message {
	n.msgIDSeq += 1

	return Message{
		Src:  in.Dest,
		Dest: in.Src,
		Body: RawMessage{
			"type":        "echo_ok",
			"msg_id":      n.msgIDSeq,
			"in_reply_to": in.Body["msg_id"].(float64),
			"echo":        in.Body["echo"],
		},
	}
}

func (n *Node) handleInit(in Message) Message {
	n.setID(in.Body["node_id"].(string))
	n.msgIDSeq += 1

	return Message{
		Src:  in.Dest,
		Dest: in.Src,
		Body: RawMessage{
			"type":        "init_ok",
			"msg_id":      n.msgIDSeq,
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

func main() {
	node := Node{
		log:    log.New(os.Stderr, "", 0),
		output: log.New(os.Stdout, "", 0),
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		node.spawnHandler(input, err)
	}
}

func parseMessage(input string) (Message, error) {
	var msg Message
	if err := json.Unmarshal([]byte(input), &msg); err != nil {
		return Message{}, errors.New("invalid json")
	}

	// TODO Leginimately validate message here before pass it further (including client and node ids fromat).

	return msg, nil
}

func (e ErrorCode) String() string {
	switch e {
	case Timeout:
		return "Timed out"
	case NodeNotFound:
		return "Node not found"
	case NotSupported:
		return "Operation not supported"
	case Unavailable:
		return "Temporary unavailable"
	case Malformed:
		return "Malformed request"
	case Crashed:
		return "Crashed"
	case Aborted:
		return "Aborted"
	case KeyNotExist:
		return "Key does not exist"
	case KeyAlreadyExist:
		return "Key already exist"
	case PreconditionFailed:
		return "Precondition failed"
	case TXConflict:
		return "Transaction conflict"
	default:
		panic("unreachable")
	}
}
