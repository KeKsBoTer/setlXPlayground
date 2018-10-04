package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Message contains a std-output
// It holds the text and to which output is was printed
type Message struct {
	// Text is the message as string
	Text string

	// Kind is the channel the message was printed to
	// "stdout" or "stderr"
	Kind string

	// Delay is the time elapsed since the last message
	Delay int
}

// Response contains a list of events.
// Every event is a message.
type Response struct {
	Events []Message `json:"events"`
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

	start time.Time
}

func (w *writer) Write(bytes []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	newMessage := Message{
		Kind:  w.source,
		Text:  string(bytes),
		Delay: int(time.Since(w.start) / time.Millisecond),
	}
	*w.writes = append(*w.writes, newMessage)
	return len(bytes), nil
}

const maxExecutionTime = 3 * time.Minute

// Workaround
// create temp file, run it and then delte it
func runCode(code string) ([]Message, error) {
	code += "\n return 0;"
	ctx, cancel := context.WithTimeout(context.Background(), maxExecutionTime)
	defer cancel()
	cmd := exec.CommandContext(ctx, "java", "-Djava.security.manager", "-Djava.security.policy=java.policy", "-cp", "setlx/setlX.jar", "org.randoom.setlx.pc.ui.SetlX", "-x", code)

	var mu sync.Mutex
	var messages []Message
	startTime := time.Now()

	cmd.Stderr = &writer{
		source: "stderr",
		mu:     &mu,
		writes: &messages,
		start:  startTime,
	}
	cmd.Stdout = &writer{
		source: "stdout",
		mu:     &mu,
		writes: &messages,
		start:  startTime,
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
	} else {
		last := &messages[len(messages)-1]
		if strings.HasPrefix(last.Text, "Internal error.") {
			last.Text = "Internal error (did you try to read or write from the disk?)\n"
		}

		duration := time.Since(startTime) / time.Millisecond
		messages = append(messages, Message{
			Text:  fmt.Sprintf("----------------------\nFinished in %.2f seconds\n", float64(duration)/1000),
			Kind:  "stdout",
			Delay: 0,
		})
	}
	return messages, nil
}
