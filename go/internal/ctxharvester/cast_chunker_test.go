package ctxharvester

import (
	"strings"
	"testing"
)

func TestCastChunkGo(t *testing.T) {
	goCode := `package main

import (
	"fmt"
	"os"
)

// A main function
func main() {
	fmt.Println("Hello world")
}

type MyStruct struct {
	Val int
}

func (m *MyStruct) Process() {
	println(m.Val)
}
`
	metadata := map[string]interface{}{"filename": "main.go"}
	ch := NewContextHarvester(nil)
	chunks := ch.Harvest(SourceActiveFile, goCode, metadata)

	if len(chunks) == 0 {
		t.Fatal("expected chunks, got none")
	}

	// Verify that the contextual headers got prepended
	for _, chunk := range chunks {
		if !strings.Contains(chunk.Content, "[cAST Context: go | File: main.go]") {
			t.Errorf("chunk missing expected header comment, got: %s", chunk.Content)
		}
		if !strings.Contains(chunk.Content, "package main") {
			t.Errorf("chunk missing package declaration in header, got: %s", chunk.Content)
		}
	}
}

func TestCastChunkPython(t *testing.T) {
	pyCode := `import os
import sys

class Agent:
    def __init__(self, name):
        self.name = name

    def act(self):
        print(f"Agent {self.name} acting")

def helper_func():
    return 42
`
	metadata := map[string]interface{}{"path": "/project/agent.py"}
	ch := NewContextHarvester(nil)
	chunks := ch.Harvest(SourceActiveFile, pyCode, metadata)

	if len(chunks) == 0 {
		t.Fatal("expected Python chunks, got none")
	}

	for _, chunk := range chunks {
		if !strings.Contains(chunk.Content, "[cAST Context: python | File: agent.py]") {
			t.Errorf("chunk missing expected python header comment, got: %s", chunk.Content)
		}
		if !strings.Contains(chunk.Content, "#   import os") {
			t.Errorf("chunk missing import statement in header context, got: %s", chunk.Content)
		}
	}
}

func TestCastChunkJSFamily(t *testing.T) {
	tsCode := `import { useState } from 'react';

export function Counter() {
	const [count, setCount] = useState(0);
	return <button onClick={() => setCount(count + 1)}>{count}</button>;
}

export class Helper {
	static log(msg: string) {
		console.log(msg);
	}
}
`
	metadata := map[string]interface{}{"filename": "Counter.tsx"}
	ch := NewContextHarvester(nil)
	chunks := ch.Harvest(SourceActiveFile, tsCode, metadata)

	if len(chunks) == 0 {
		t.Fatal("expected TS chunks, got none")
	}

	for _, chunk := range chunks {
		if !strings.Contains(chunk.Content, "[cAST Context: js/ts | File: Counter.tsx]") {
			t.Errorf("chunk missing expected ts/js header comment, got: %s", chunk.Content)
		}
	}
}
