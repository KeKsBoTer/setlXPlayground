package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
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
	log.Println("Executing code...")
	start := time.Now()
	out, err := runAsFile(string(code))
	if err != nil {
		return nil, err
	}
	log.Printf("Finished after %s seconds\n", time.Since(start).String())
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

func runAsString(code string) ([]byte, error) {
	cmd := exec.Command("setlx", "-x", strings.Replace(code, "\n", "\n", -1))
	return cmd.CombinedOutput()
}

// Workaround
// create temp file, run it and then delte it
func runAsFile(code string) ([]Message, error) {
	f, err := ioutil.TempFile("tmp", "code")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	f.WriteString(code)

	cmd := exec.Command("setlX", f.Name())

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
	return messages, nil
}
