package main

import (
	"encoding/json"
	"io"
	"os/exec"
	"sync"
)

type config struct {
	Name            string              `json:"name"`
	Command         []string            `json:"command"`
	Env             map[string]string   `json:"env,omitempty"`
	Root            string              `json:"root"`
	File            string              `json:"file"`
	LanguageID      string              `json:"languageId"`
	TimeoutMS       int                 `json:"timeoutMs,omitempty"`
	WaitDiagnostics *bool               `json:"waitDiagnostics,omitempty"`
	Positions       map[string]position `json:"positions,omitempty"`
}

type position struct {
	Line      uint32 `json:"line"`
	Character uint32 `json:"character"`
}

type rpcMessage struct {
	ID     any             `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  any             `json:"error,omitempty"`
}

type client struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	pending map[int]chan rpcMessage
	diagCh  chan struct{}
	done    chan error
	mu      sync.Mutex
	nextID  int
}

type metric struct {
	Name string  `json:"name"`
	MS   float64 `json:"ms"`
	N    int     `json:"n,omitempty"`
}

type runResult struct {
	Name    string   `json:"name"`
	Count   int      `json:"count"`
	Runs    int      `json:"runs"`
	Warmup  int      `json:"warmup"`
	Metrics []metric `json:"metrics"`
}

type runFailure struct {
	Config string `json:"config"`
	Error  string `json:"error"`
}

type batchResult struct {
	Results []runResult  `json:"results"`
	Errors  []runFailure `json:"errors,omitempty"`
}

type options struct {
	count  int
	runs   int
	warmup int
}
