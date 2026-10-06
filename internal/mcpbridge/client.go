// Package mcpbridge connects Butterfly to the existing HackAlem MCP process.
// Protocol flow adapted from HackAlem Agent Kit's cmd/hackalem-rehearse/main.go
// (MCP client, lines 278–412). The supplied checkout contains no LICENSE file;
// no license attribution is inferred. This bridge adds bounded I/O and isolation.
package mcpbridge

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	protocolVersion = "2025-06-18"
	maxMessageBytes = 2 << 20
	requestTimeout  = 30 * time.Second
)

var allowedTools = map[string]bool{
	"system.context": true, "system.capabilities": true,
	"routing.fly_recommend": true, "routing.fly_feedback": true,
	"reports.create": true, "telegram.send": true, "telegram.validate": true,
}

type readResult struct {
	line []byte
	err  error
}

// Client owns one initialized subprocess. Calls are serialized using a
// context-aware mutex so canceled callers do not wait behind another tool call.
type Client struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	lines     chan readResult
	done      chan struct{}
	exited    chan struct{}
	mutex     chan struct{}
	closeOnce sync.Once
	nextID    uint64
}

// New starts the existing binary against a dedicated data root. An empty value
// (rather than omission) prevents inherited AI keys from being loaded as defaults.
func New(ctx context.Context, binary, repoRoot string) (*Client, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if binary == "" || repoRoot == "" {
		return nil, errors.New("MCP binary and dedicated root are required")
	}
	cmd := exec.Command(binary, "--repo-root", repoRoot)
	cmd.Dir = repoRoot
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if key != "OPENAI_API_KEY" && key != "NVIDIA_API_KEY" && key != "POSTGRES_DSN" {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	cmd.Env = append(cmd.Env, "OPENAI_API_KEY=", "NVIDIA_API_KEY=", "POSTGRES_DSN=")
	// The plugin can include configuration in stderr. Retain zero bytes, and
	// never include its raw diagnostics or credential-bearing URLs in errors.
	cmd.Stderr = io.Discard
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("open MCP input: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("open MCP output: %w", err)
	}
	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		return nil, fmt.Errorf("start MCP: %w", err)
	}
	c := &Client{cmd: cmd, stdin: stdin, lines: make(chan readResult, 1), done: make(chan struct{}), exited: make(chan struct{}), mutex: make(chan struct{}, 1)}
	c.mutex <- struct{}{}
	go func() { _ = cmd.Wait(); close(c.exited) }()
	go c.read(stdout)
	if err := c.acquire(ctx); err != nil {
		c.Close()
		return nil, err
	}
	raw, err := c.request(ctx, "initialize", map[string]any{
		"protocolVersion": protocolVersion,
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "butterfly", "version": "0.1.0"},
	})
	if err == nil {
		var handshake struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		if json.Unmarshal(raw, &handshake) != nil || handshake.ProtocolVersion != protocolVersion {
			err = errors.New("MCP server returned an incompatible initialization")
		} else {
			err = c.write(ctx, []byte(`{"jsonrpc":"2.0","method":"notifications/initialized"}`))
		}
	}
	c.release()
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("initialize MCP: %w", err)
	}
	go func() {
		select {
		case <-ctx.Done():
			_ = c.Close()
		case <-c.exited:
			_ = c.Close()
		case <-c.done:
		}
	}()
	return c, nil
}

func (c *Client) read(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), maxMessageBytes)
	for scanner.Scan() {
		line := append([]byte(nil), scanner.Bytes()...)
		select {
		case c.lines <- readResult{line: line}:
		case <-c.done:
			return
		}
	}
	err := scanner.Err()
	if err == nil {
		err = io.EOF
	}
	select {
	case c.lines <- readResult{err: err}:
	case <-c.done:
	}
}

func (c *Client) acquire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return errors.New("MCP client is closed")
	case <-c.mutex:
		if err := ctx.Err(); err != nil {
			c.release()
			return err
		}
		select {
		case <-c.done:
			c.release()
			return errors.New("MCP client is closed")
		default:
		}
		return nil
	}
}

func (c *Client) release() { c.mutex <- struct{}{} }

func (c *Client) write(ctx context.Context, line []byte) error {
	if len(line) >= maxMessageBytes {
		return errors.New("MCP request exceeds size limit")
	}
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	result := make(chan error, 1)
	go func() { _, err := c.stdin.Write(append(line, '\n')); result <- err }()
	select {
	case err := <-result:
		if err != nil {
			c.Close()
			return errors.New("MCP process input closed")
		}
		return nil
	case <-ctx.Done():
		c.Close()
		return ctx.Err()
	case <-c.done:
		return errors.New("MCP client is closed")
	}
}

func (c *Client) request(ctx context.Context, method string, params map[string]any) (json.RawMessage, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	c.nextID++
	id := c.nextID
	line, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params})
	if err != nil {
		return nil, errors.New("MCP arguments are not JSON serializable")
	}
	if err = c.write(ctx, line); err != nil {
		return nil, err
	}
	for {
		select {
		case <-ctx.Done():
			c.Close()
			return nil, ctx.Err()
		case <-c.done:
			return nil, errors.New("MCP client is closed")
		case result := <-c.lines:
			if result.err != nil {
				c.Close()
				return nil, errors.New("MCP process output closed or exceeded size limit")
			}
			var reply struct {
				JSONRPC string          `json:"jsonrpc"`
				ID      json.RawMessage `json:"id"`
				Method  string          `json:"method"`
				Result  json.RawMessage `json:"result"`
				Error   *struct {
					Code int `json:"code"`
				} `json:"error"`
			}
			if json.Unmarshal(result.line, &reply) != nil || reply.JSONRPC != "2.0" {
				c.Close()
				return nil, errors.New("MCP returned invalid JSON-RPC")
			}
			if len(reply.ID) == 0 && reply.Method != "" {
				continue
			}
			if string(reply.ID) != strconv.FormatUint(id, 10) {
				c.Close()
				return nil, errors.New("MCP response ID does not match request")
			}
			if reply.Error != nil {
				return nil, fmt.Errorf("MCP RPC error %d", reply.Error.Code)
			}
			if len(reply.Result) == 0 || string(reply.Result) == "null" {
				c.Close()
				return nil, errors.New("MCP response has no result")
			}
			return reply.Result, nil
		}
	}
}

// Call exposes only the explicitly selected project capabilities. It returns
// HackAlem's envelope unchanged, including partial/error statuses and trace IDs.
func (c *Client) Call(ctx context.Context, tool string, args map[string]any) (map[string]any, error) {
	if !allowedTools[tool] {
		return nil, errors.New("MCP tool is not allowed by Butterfly")
	}
	if err := c.acquire(ctx); err != nil {
		return nil, err
	}
	defer c.release()
	if args == nil {
		args = map[string]any{}
	}
	raw, err := c.request(ctx, "tools/call", map[string]any{"name": tool, "arguments": args})
	if err != nil {
		return nil, err
	}
	var result struct {
		IsError bool `json:"isError"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if json.Unmarshal(raw, &result) != nil {
		return nil, errors.New("MCP returned an invalid tool result")
	}
	for _, block := range result.Content {
		if block.Type != "text" {
			continue
		}
		var envelope map[string]any
		if json.Unmarshal([]byte(block.Text), &envelope) != nil {
			return nil, errors.New("MCP returned an invalid HackAlem envelope")
		}
		if status, ok := envelope["status"].(string); !ok || status == "" {
			return nil, errors.New("MCP envelope has no status")
		}
		if result.IsError {
			return envelope, errors.New("MCP tool reported an error")
		}
		return envelope, nil
	}
	return nil, errors.New("MCP tool returned no text envelope")
}

// Running reports whether the initialized transport and child are still open.
func (c *Client) Running() bool {
	if c == nil {
		return false
	}
	select {
	case <-c.done:
		return false
	case <-c.exited:
		return false
	default:
		return true
	}
}

// Close is idempotent and interrupts pending reads/writes before reaping the child.
func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	c.closeOnce.Do(func() { close(c.done); _ = c.stdin.Close(); _ = c.cmd.Process.Kill(); <-c.exited })
	return nil
}
