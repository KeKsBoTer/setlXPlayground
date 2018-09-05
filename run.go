package main

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"sync"
	"time"
)

type Message struct {
	Text  string
	Kind  string
	Delay int
}

type Response struct {
	Events []Message
}

func run(code []byte) ([]byte, error) {
	if len(code) == 0 {
		return nil, errors.New("Empty code")
	}
	out, err := runCode(string(code))
	if err != nil {
		return nil, err
	}
	r := Response{
		Events: out,
	}
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type writer struct {
	source string

	mu     *sync.Mutex
	writes *[]Message
}

func (w *writer) Write(bytes []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	*w.writes = append(*w.writes, Message{
		Kind:  w.source,
		Text:  string(bytes),
		Delay: int(time.Now().Unix()),
	})
	return len(bytes), nil
}

const maxExecutionTime = 3 * time.Minute

// Workaround
// create temp file, run it and then delte it
func runCode(code string) ([]Message, error) {
	code += "\n return 0;"
	ctx, cancel := context.WithTimeout(context.Background(), maxExecutionTime)
	defer cancel()
	cmd := exec.CommandContext(ctx, "java", "-cp", "setlx/setlX.jar", "org.randoom.setlx.pc.ui.SetlX", "-x", code)

	var mu sync.Mutex
	var messages []Message

	cmd.Stderr = &writer{
		source: "stderr",
		mu:     &mu,
		writes: &messages,
	}
	cmd.Stdout = &writer{
		source: "stdout",
		mu:     &mu,
		writes: &messages,
	}
	cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		messages = append(messages, Message{
			Text:  "Programm exceeded maximum execution time",
			Kind:  "stderr",
			Delay: 0,
		})
	}
	if len(messages) == 0 {
		messages = append(messages, Message{})
	}
	return messages, nil
}
