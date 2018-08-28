package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Message struct {
	Text  []byte
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
	lines := strings.Split(string(out), "\n")
	r := Response{
		Events: make([]Message, len(lines)),
	}
	for i, l := range lines {
		r.Events[i] = Message{
			Text:  []byte(l),
			Kind:  "stdout",
			Delay: 0,
		}
	}
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func runAsString(code string) ([]byte, error) {
	cmd := exec.Command("setlx", "-x", strings.Replace(code, "\n", "\n", -1))
	return cmd.CombinedOutput()
}

// Workaround
// create temp file, run it and then delte it
func runAsFile(code string) ([]byte, error) {
	f, err := ioutil.TempFile("tmp", "code")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	f.WriteString(code)

	cmd := exec.Command("setlx", f.Name())
	return cmd.CombinedOutput()
}
