package mcpbridge

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if mode := os.Getenv("MCPBRIDGE_TEST_HELPER"); mode != "" {
		helper(mode)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func helper(mode string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var request struct {
			ID     int            `json:"id"`
			Method string         `json:"method"`
			Params map[string]any `json:"params"`
		}
		if json.Unmarshal(scanner.Bytes(), &request) != nil {
			os.Exit(2)
		}
		if request.Method == "notifications/initialized" {
			continue
		}
		if request.Method == "initialize" {
			protocol := protocolVersion
			if mode == "bad-init" {
				protocol = "unsupported"
			}
			if request.Params["protocolVersion"] != protocolVersion {
				os.Exit(3)
			}
			respond(request.ID, map[string]any{"protocolVersion": protocol})
			continue
		}
		switch mode {
		case "invalid":
			fmt.Println("not-json")
			continue
		case "wrong-id":
			respond(request.ID+1, map[string]any{})
			continue
		case "exit":
			os.Exit(4)
		case "timeout":
			time.Sleep(10 * time.Second)
			continue
		case "oversized":
			fmt.Println(strings.Repeat("x", maxMessageBytes+1))
			continue
		case "rpc-error":
			fmt.Printf(`{"jsonrpc":"2.0","id":%d,"error":{"code":-1,"message":"SECRET-NOT-LEAKED"}}`+"\n", request.ID)
			continue
		case "tool-error":
			respond(request.ID, map[string]any{"isError": true, "content": []any{map[string]any{"type": "text", "text": `{"status":"ok","data":{"error":"SECRET-NOT-LEAKED"}}`}}})
			continue
		}
		fmt.Println(`{"jsonrpc":"2.0","method":"notifications/progress","params":{}}`)
		if mode == "bad-envelope" {
			respond(request.ID, map[string]any{"content": []any{map[string]any{"type": "text", "text": "SECRET-NOT-LEAKED"}}})
			continue
		}
		envelope, _ := json.Marshal(map[string]any{"status": "success", "trace_id": fmt.Sprintf("trace-%d", request.ID), "data": map[string]any{
			"ai_keys_cleared": os.Getenv("OPENAI_API_KEY") == "" && os.Getenv("NVIDIA_API_KEY") == "",
			"tool":            request.Params["name"],
		}})
		respond(request.ID, map[string]any{"content": []any{map[string]any{"type": "text", "text": string(envelope)}}})
	}
}

func respond(id int, result any) {
	_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"jsonrpc": "2.0", "id": id, "result": result})
}

func startHelper(t *testing.T, mode string) *Client {
	t.Helper()
	t.Setenv("MCPBRIDGE_TEST_HELPER", mode)
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	client, err := New(context.Background(), binary, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })
	return client
}

func TestPersistentClientClearsAIKeysPreservesEnvelopeAndRejectsOtherTools(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "fake-sensitive-openai-value")
	t.Setenv("NVIDIA_API_KEY", "fake-sensitive-nvidia-value")
	client := startHelper(t, "normal")
	for _, tool := range []string{"system.context", "routing.fly_recommend", "routing.fly_feedback"} {
		envelope, err := client.Call(context.Background(), tool, map[string]any{"task": "meeting review"})
		if err != nil {
			t.Fatal(err)
		}
		if envelope["status"] != "success" || envelope["trace_id"] == nil {
			t.Fatalf("invalid envelope: %#v", envelope)
		}
		data := envelope["data"].(map[string]any)
		if data["ai_keys_cleared"] != true || data["tool"] != tool {
			t.Fatalf("unexpected data: %#v", data)
		}
	}
	if _, err := client.Call(context.Background(), "system.exec", nil); err == nil {
		t.Fatal("arbitrary tool accepted")
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Call(context.Background(), "system.context", nil); err == nil {
		t.Fatal("closed client accepted call")
	}
}

func TestConcurrentCallsAreSerialized(t *testing.T) {
	client := startHelper(t, "normal")
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			if _, err := client.Call(context.Background(), "system.context", nil); err != nil {
				t.Errorf("concurrent call: %v", err)
			}
		})
	}
	wg.Wait()
}

func TestInvalidProcessRepliesFailCleanly(t *testing.T) {
	for _, mode := range []string{"invalid", "wrong-id", "exit", "oversized", "rpc-error", "tool-error", "bad-envelope"} {
		t.Run(mode, func(t *testing.T) {
			client := startHelper(t, mode)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_, err := client.Call(ctx, "system.context", nil)
			if err == nil {
				t.Fatal("invalid response accepted")
			}
			if strings.Contains(err.Error(), "SECRET-NOT-LEAKED") {
				t.Fatal("raw process text leaked")
			}
		})
	}
}

func TestTimeoutClosesProcessAndMakesFutureCallsFail(t *testing.T) {
	client := startHelper(t, "timeout")
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	_, err := client.Call(ctx, "system.context", nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("want deadline error, got %v", err)
	}
	select {
	case <-client.exited:
	default:
		t.Fatal("timed-out child was not reaped")
	}
	if _, err := client.Call(context.Background(), "system.context", nil); err == nil {
		t.Fatal("timed-out client reused")
	}
}

func TestInitializationAndCanceledContextsFail(t *testing.T) {
	t.Setenv("MCPBRIDGE_TEST_HELPER", "bad-init")
	binary, _ := os.Executable()
	if _, err := New(context.Background(), binary, t.TempDir()); err == nil {
		t.Fatal("incompatible protocol accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := New(ctx, binary, t.TempDir()); !errors.Is(err, context.Canceled) {
		t.Fatalf("want canceled, got %v", err)
	}
}

func TestOwnerCancellationReapsIdleProcess(t *testing.T) {
	t.Setenv("MCPBRIDGE_TEST_HELPER", "normal")
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client, err := New(ctx, binary, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	cancel()
	select {
	case <-client.exited:
	case <-time.After(time.Second):
		t.Fatal("canceling owner left idle MCP process running")
	}
}

func TestInstalledHackAlemHandshake(t *testing.T) {
	binary := os.Getenv("HACKALEM_MCP_TEST_BINARY")
	if binary == "" {
		t.Skip("set HACKALEM_MCP_TEST_BINARY for installed-plugin smoke")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := New(ctx, binary, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	envelope, err := client.Call(ctx, "system.context", nil)
	if err != nil {
		t.Fatal(err)
	}
	if envelope["status"] == "error" || envelope["data"] == nil {
		t.Fatalf("plugin handshake failed: status=%v", envelope["status"])
	}
}
