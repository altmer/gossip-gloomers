package main

import (
	"encoding/json"
	"log"
	"os"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func main() {
	log.SetOutput(os.Stderr)

	n := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(n)

	n.Handle("send", func(msg maelstrom.Message) error {
		log.Println("handling send")
		var request map[string]any
		if err := json.Unmarshal(msg.Body, &request); err != nil {
			return err
		}

		key := request["key"].(string)
		message := int(request["msg"].(float64))

		response := make(map[string]any)
		response["type"] = "send_ok"
		return n.Reply(msg, response)
	})

	n.Handle("poll", func(msg maelstrom.Message) error {
		response := make(map[string]any)
		response["type"] = "poll_ok"
		response["msgs"] = make(map[string][2][2]int)
		return n.Reply(msg, response)
	})

	n.Handle("commit_offsets", func(msg maelstrom.Message) error {
		log.Println("handling commit_offsets")
		var request map[string]any
		if err := json.Unmarshal(msg.Body, &request); err != nil {
			return err
		}

		response := make(map[string]any)
		response["type"] = "commit_offsets_ok"
		return n.Reply(msg, response)
	})

	n.Handle("list_committed_offsets", func(msg maelstrom.Message) error {
		log.Println("handling list_committed_offsets")
		var request map[string]any
		if err := json.Unmarshal(msg.Body, &request); err != nil {
			return err
		}

		response := make(map[string]any)
		response["type"] = "list_committed_offsets_ok"
		return n.Reply(msg, response)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
