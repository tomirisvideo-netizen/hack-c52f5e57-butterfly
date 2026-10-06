package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"butterfly/internal/app"
)

type recordedCaller struct {
	mu    sync.Mutex
	calls []string
	err   error
}

func (c *recordedCaller) Call(_ context.Context, name string, _ map[string]any) (map[string]any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, name)
	return map[string]any{"status": "ok"}, c.err
}
func (c *recordedCaller) count() int { c.mu.Lock(); defer c.mu.Unlock(); return len(c.calls) }

func serveTest(mux http.Handler, method, path string, cookie *http.Cookie, body io.Reader) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, body)
	if cookie != nil {
		r.AddCookie(cookie)
	}
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	return w
}
func testSession(t *testing.T, mux http.Handler) *http.Cookie {
	t.Helper()
	w := serveTest(mux, "POST", "/api/session", nil, nil)
	if w.Code != 200 {
		t.Fatalf("session %d: %s", w.Code, w.Body.String())
	}
	return w.Result().Cookies()[0]
}
func completedDemo(t *testing.T, mux http.Handler, cookie *http.Cookie) string {
	t.Helper()
	w := serveTest(mux, "POST", "/api/demo", cookie, nil)
	var m map[string]any
	if w.Code != 201 || json.Unmarshal(w.Body.Bytes(), &m) != nil {
		t.Fatalf("demo %d: %s", w.Code, w.Body.String())
	}
	id := m["id"].(string)
	for _, a := range m["actions"].([]any) {
		a.(map[string]any)["status"] = "confirmed"
	}
	body, _ := json.Marshal(map[string]any{"actions": m["actions"]})
	if w = serveTest(mux, "PATCH", "/api/meetings/"+id, cookie, bytes.NewReader(body)); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w = serveTest(mux, "POST", "/api/meetings/"+id+"/confirm", cookie, nil); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	return id
}
func setupRuntime(t *testing.T, c *recordedCaller) (*Services, *http.ServeMux) {
	t.Helper()
	t.Setenv("TELEGRAM_BOT_TOKEN", "mock-token-never-sent")
	t.Setenv("TELEGRAM_CHAT_ID", "mock-chat")
	a, err := app.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	s := &Services{ctx: context.Background(), client: c, root: t.TempDir()}
	mux := http.NewServeMux()
	a.Register(mux)
	s.Register(mux, a)
	return s, mux
}

func TestNotifyOwnershipReviewAndConcurrentDedupe(t *testing.T) {
	c := &recordedCaller{}
	_, mux := setupRuntime(t, c)
	alice, bob := testSession(t, mux), testSession(t, mux)
	id := completedDemo(t, mux, alice)
	path := "/api/meetings/" + id + "/notify"
	if w := serveTest(mux, "POST", path, nil, nil); w.Code != 401 {
		t.Fatal(w.Code)
	}
	if w := serveTest(mux, "POST", path, bob, nil); w.Code != 404 {
		t.Fatal(w.Code)
	}
	if c.count() != 0 {
		t.Fatal("unauthorized send reached MCP")
	}
	var wg sync.WaitGroup
	statuses := make(chan int, 8)
	for range 8 {
		wg.Add(1)
		go func() { defer wg.Done(); statuses <- serveTest(mux, "POST", path, alice, nil).Code }()
	}
	wg.Wait()
	close(statuses)
	sent := 0
	for code := range statuses {
		if code == 200 {
			sent++
		} else if code != 409 {
			t.Fatalf("unexpected notify status %d", code)
		}
	}
	if sent != 1 || c.count() != 1 {
		t.Fatalf("sent=%d MCP calls=%d", sent, c.count())
	}
}

func TestNotifyUncertainFailureDoesNotRetry(t *testing.T) {
	c := &recordedCaller{err: errors.New("mock timeout")}
	_, mux := setupRuntime(t, c)
	cookie := testSession(t, mux)
	id := completedDemo(t, mux, cookie)
	path := "/api/meetings/" + id + "/notify"
	if w := serveTest(mux, "POST", path, cookie, nil); w.Code != 502 {
		t.Fatal(w.Code)
	}
	if w := serveTest(mux, "POST", path, cookie, nil); w.Code != 409 {
		t.Fatal(w.Code)
	}
	if c.count() != 1 {
		t.Fatal("uncertain send repeated")
	}
}

func TestNotifyRequiresConfiguredChannelAndReview(t *testing.T) {
	c := &recordedCaller{}
	_, mux := setupRuntime(t, c)
	cookie := testSession(t, mux)
	w := serveTest(mux, "POST", "/api/demo", cookie, nil)
	var m map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &m)
	if w = serveTest(mux, "POST", "/api/meetings/"+m["id"].(string)+"/notify", cookie, nil); w.Code != 409 {
		t.Fatal(w.Code)
	}
	id := completedDemo(t, mux, cookie)
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	if w = serveTest(mux, "POST", "/api/meetings/"+id+"/notify", cookie, nil); w.Code != 503 {
		t.Fatal(w.Code)
	}
	if c.count() != 0 {
		t.Fatal("invalid request reached MCP")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFlyProxyAllowsOnlyVisualizationRoutesWithSession(t *testing.T) {
	c := &recordedCaller{}
	s, mux := setupRuntime(t, c)
	target, _ := url.Parse("http://unused.invalid")
	s.proxy = httputil.NewSingleHostReverseProxy(target)
	calls := 0
	s.proxy.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("demo"))}, nil
	})
	cookie := testSession(t, mux)
	for _, path := range []string{"/fly-demo", "/fly-live", "/fly-events", "/fly-feedback", "/fly-neuron-events"} {
		if w := serveTest(mux, "GET", path, nil, nil); w.Code != 401 {
			t.Fatalf("%s without session: %d", path, w.Code)
		}
		if w := serveTest(mux, "GET", path, cookie, nil); w.Code != 200 {
			t.Fatalf("%s: %d", path, w.Code)
		}
	}
	for _, path := range []string{"/control", "/calls", "/inputs", "/artifacts", "/routing", "/fly-demo/control"} {
		if w := serveTest(mux, "GET", path, cookie, nil); w.Code != 404 {
			t.Fatalf("control path %s exposed: %d", path, w.Code)
		}
	}
	if w := serveTest(mux, "POST", "/fly-demo", cookie, nil); w.Code != 405 {
		t.Fatal(w.Code)
	}
	if calls != 5 {
		t.Fatalf("unexpected upstream requests: %d", calls)
	}
}

func TestCloseStopsWorkerWithoutCancelingParent(t *testing.T) {
	t.Setenv("HACKALEM_HOME", "")
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := Start(parent, t.TempDir())
	s.Close()
	s.Close()
	select {
	case <-s.workerDone:
	case <-time.After(time.Second):
		t.Fatal("Close left the integration worker running")
	}
	if parent.Err() != nil {
		t.Fatal("Close canceled the caller's context")
	}
}

func TestHealthAndTelegramReflectCanceledRuntime(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "mock-token-never-sent")
	t.Setenv("TELEGRAM_CHAT_ID", "mock-chat")
	ctx, cancel := context.WithCancel(context.Background())
	s := &Services{ctx: ctx, client: &recordedCaller{}, proxy: &httputil.ReverseProxy{}, configured: true}
	cancel()
	health := s.Health()
	if health["connected"] != false || health["fly_available"] != false || health["error"] == "" {
		t.Fatalf("stopped runtime advertised healthy integrations: %#v", health)
	}
	if s.TelegramConfigured() {
		t.Fatal("stopped runtime advertised available Telegram")
	}
}

func TestFlyCompatibilityRequiresVisualizationPage(t *testing.T) {
	for _, tc := range []struct {
		name        string
		status      int
		contentType string
		want        bool
	}{
		{"current page", http.StatusOK, "text/html; charset=utf-8", true},
		{"old binary without page", http.StatusNotFound, "text/plain", false},
		{"incompatible response", http.StatusOK, "application/json", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/mb-3d" {
					t.Errorf("unexpected compatibility probe: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", tc.contentType)
				w.WriteHeader(tc.status)
			}))
			defer server.Close()
			target, _ := url.Parse(server.URL)
			if got := flyVisualizationReady(context.Background(), server.Client(), target); got != tc.want {
				t.Fatalf("visualization readiness = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestInstalledHackAlemRuntime(t *testing.T) {
	home := os.Getenv("HACKALEM_TEST_HOME")
	if home == "" {
		t.Skip("set HACKALEM_TEST_HOME for installed MCP/Fly smoke")
	}
	t.Setenv("HACKALEM_HOME", home)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := Start(ctx, t.TempDir())
	defer s.Close()
	if health := s.Health(); health["connected"] != true || health["fly_available"] != true {
		t.Fatalf("installed integrations did not start: %#v", health)
	}
	// Only synthetic metadata and a local report are sent to the private MCP root.
	// Telegram is deliberately excluded from this installed-runtime smoke test.
	for _, call := range []struct {
		tool string
		args map[string]any
	}{
		{"system.capabilities", nil},
		{"routing.fly_recommend", map[string]any{"task": auditTask, "available_capabilities": []string{"reports.create"}, "input_metadata": map[string]any{"type": "meeting_review", "mode": "demo"}}},
		{"reports.create", map[string]any{"title": "Synthetic Butterfly compatibility test", "format": "md", "sections": []map[string]string{{"heading": "Local smoke", "body": "Synthetic metadata only."}}}},
		{"routing.fly_feedback", map[string]any{"task": auditTask, "chosen": []string{"reports.create"}, "source": "normal-teacher", "input_metadata": map[string]any{"type": "meeting_review", "mode": "demo"}}},
	} {
		result, err := s.call(call.tool, call.args)
		if !success(result, err) {
			t.Fatalf("installed %s: status=%v, err=%v", call.tool, result["status"], err)
		}
	}
	a, err := app.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	a.Register(mux)
	s.Register(mux, a)
	cookie := testSession(t, mux)
	for _, path := range []string{"/fly-demo", "/fly-events", "/fly-feedback", "/fly-neuron-events"} {
		w := serveTest(mux, "GET", path, cookie, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("installed Fly %s returned %d", path, w.Code)
		}
		if path == "/fly-demo" && (!strings.Contains(w.Body.String(), "/vendor/three.module.js") || strings.Contains(w.Body.String(), "https://cdn.jsdelivr.net/")) {
			t.Fatal("Fly demo did not use the local Three.js module")
		}
	}
	cancel()
	select {
	case <-s.demoDone:
	case <-time.After(time.Second):
		t.Fatal("canceling runtime left the Fly process running")
	}
	s.Close()
	if health := s.Health(); health["connected"] != false || health["fly_available"] != false {
		t.Fatalf("closed runtime advertised healthy integrations: %#v", health)
	}
}
