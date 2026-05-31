package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var errRPC = errors.New("rpc error")

func startClient(ctx context.Context, cfg config) (*client, error) {
	cmd := exec.CommandContext(ctx, cfg.Command[0], cfg.Command[1:]...)
	cmd.Dir = cfg.Root
	cmd.Env = os.Environ()
	for key, value := range cfg.Env {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	c := &client{
		cmd:     cmd,
		stdin:   stdin,
		pending: make(map[int]chan rpcMessage),
		diagCh:  make(chan struct{}, 16),
		done:    make(chan error, 1),
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start %q: %w", cfg.Command[0], err)
	}
	go c.readLoop(stdout)
	go func() {
		err := cmd.Wait()
		if err != nil && stderr.Len() > 0 {
			err = fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		c.done <- err
	}()
	return c, nil
}

func (c *client) request(ctx context.Context, method string, params any) (rpcMessage, error) {
	c.mu.Lock()
	c.nextID++
	id := c.nextID
	ch := make(chan rpcMessage, 1)
	c.pending[id] = ch
	c.mu.Unlock()

	if err := c.write(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params}); err != nil {
		return rpcMessage{}, err
	}
	select {
	case msg := <-ch:
		if msg.Error != nil {
			return rpcMessage{}, fmt.Errorf("%w: %v", errRPC, msg.Error)
		}
		return msg, nil
	case err := <-c.done:
		if err == nil {
			err = errors.New("server exited")
		}
		return rpcMessage{}, err
	case <-ctx.Done():
		return rpcMessage{}, ctx.Err()
	}
}

func (c *client) notify(method string, params any) error {
	return c.write(map[string]any{"jsonrpc": "2.0", "method": method, "params": params})
}

func (c *client) write(value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := fmt.Fprintf(c.stdin, "Content-Length: %d\r\n\r\n", len(data)); err != nil {
		return err
	}
	_, err = c.stdin.Write(data)
	return err
}

func (c *client) readLoop(stdout io.Reader) {
	reader := bufio.NewReader(stdout)
	for {
		data, err := readPacket(reader)
		if err != nil {
			return
		}
		var msg rpcMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		if msg.Method == "textDocument/publishDiagnostics" {
			select {
			case c.diagCh <- struct{}{}:
			default:
			}
			continue
		}
		id, ok := numericID(msg.ID)
		if !ok {
			continue
		}
		c.mu.Lock()
		ch := c.pending[id]
		delete(c.pending, id)
		c.mu.Unlock()
		if ch != nil {
			ch <- msg
		}
	}
}

func readPacket(reader *bufio.Reader) ([]byte, error) {
	length := -1
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok || !strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			continue
		}
		length, err = strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return nil, err
		}
	}
	if length < 0 {
		return nil, errors.New("missing Content-Length")
	}
	data := make([]byte, length)
	_, err := io.ReadFull(reader, data)
	return data, err
}

func (c *client) waitDiagnostics(ctx context.Context) error {
	select {
	case <-c.diagCh:
		return nil
	case err := <-c.done:
		if err == nil {
			err = errors.New("server exited before diagnostics")
		}
		return err
	case <-ctx.Done():
		return fmt.Errorf("wait diagnostics: %w", ctx.Err())
	}
}

func (c *client) close() {
	_ = c.stdin.Close()
	if c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}
}

func numericID(id any) (int, bool) {
	switch v := id.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	default:
		return 0, false
	}
}
