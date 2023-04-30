package main

import (
	"encoding/json"
	"log"
	"os"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	log.SetOutput(os.Stderr)

	n := maelstrom.NewNode()

	messages := make(map[float64]struct{})

	// { "src": "", "dest": "", "body": { "type": "generate" } }
	n.Handle("broadcast", func(msg maelstrom.Message) error {
		log.Println("handling broadcast")
		var request map[string]any
		if err := json.Unmarshal(msg.Body, &request); err != nil {
			return err
		}

		message := request["message"].(float64)
		log.Println("message received", message)

		if _, ok := messages[message]; !ok {
			messages[message] = struct{}{}
			for _, nodeId := range n.NodeIDs() {
				if nodeId == n.ID() {
					continue
				}
				propagate := make(map[string]any)

				propagate["type"] = "propagate"
				propagate["message"] = message

				n.Send(nodeId, propagate)
			}
		}

		response := make(map[string]any)
		response["type"] = "broadcast_ok"
		return n.Reply(msg, response)
	})

	n.Handle("propagate", func(msg maelstrom.Message) error {
		log.Println("handling propagate")
		var request map[string]any
		if err := json.Unmarshal(msg.Body, &request); err != nil {
			return err
		}

		message := request["message"].(float64)
		log.Println("message received", message)

		messages[message] = struct{}{}

		return nil
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		response := make(map[string]any)
		response["type"] = "read_ok"
		keys := make([]float64, 0, len(messages))
		for key := range messages {
			keys = append(keys, key)
		}
		response["messages"] = keys
		return n.Reply(msg, response)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		response := make(map[string]any)
		response["type"] = "topology_ok"
		return n.Reply(msg, response)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
