package privacylens

import "fmt"

type Message struct {
	Role    string `json:"role" yaml:"role"`
	Content string `json:"content" yaml:"content"`
}

func NormalizeMessages(input any) ([]Message, error) {
	switch v := input.(type) {
	case string:
		return []Message{{Role: "user", Content: v}}, nil
	case []Message:
		return v, nil
	case map[string]any:
		raw, ok := v["messages"]
		if !ok {
			return nil, fmt.Errorf("unrecognized object format: expected key 'messages'")
		}
		arr, ok := raw.([]Message)
		if ok {
			return arr, nil
		}
		list, ok := raw.([]any)
		if !ok {
			return nil, fmt.Errorf("messages key must be an array")
		}
		out := make([]Message, 0, len(list))
		for _, item := range list {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			role, _ := m["role"].(string)
			content, _ := m["content"].(string)
			out = append(out, Message{Role: role, Content: content})
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported message format")
	}
}
