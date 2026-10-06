package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"butterfly/internal/app"
	"butterfly/internal/mcpbridge"
)

type caller interface {
	Call(context.Context, string, map[string]any) (map[string]any, error)
}
type Services struct {
	ctx          context.Context
	cancel       context.CancelFunc
	closeOnce    sync.Once
	client       caller
	bridge       *mcpbridge.Client
	demo         *exec.Cmd
	demoDone     chan struct{}
	proxy        *httputil.ReverseProxy
	root         string
	configured   bool
	mu           sync.Mutex
	errorMessage string
	queue        chan func()
	workerDone   chan struct{}
}

// Start reuses the installed plugin with a separate, private project data root.
func Start(ctx context.Context, dataDir string) *Services {
	ctx, cancel := context.WithCancel(ctx)
	s := &Services{ctx: ctx, cancel: cancel, queue: make(chan func(), 64), workerDone: make(chan struct{})}
	go func() {
		defer close(s.workerDone)
		for {
			select {
			case <-ctx.Done():
				return
			case fn := <-s.queue:
				if ctx.Err() == nil {
					fn()
				}
			}
		}
	}()
	var err error
	s.root, err = filepath.Abs(filepath.Join(dataDir, "hackalem"))
	if err != nil {
		s.errorMessage = "Не удалось определить приватный каталог MCP."
		return s
	}
	home := os.Getenv("HACKALEM_HOME")
	if home == "" {
		s.errorMessage = "Укажите HACKALEM_HOME для готовых MCP и Fly demo."
		return s
	}
	s.configured = true
	home, err = filepath.Abs(home)
	if err != nil {
		s.errorMessage = "Не удалось определить каталог установленного HackAlem."
		return s
	}
	if err := os.MkdirAll(s.root, 0700); err != nil {
		s.errorMessage = "Не удалось создать приватный каталог MCP."
		return s
	}
	client, err := mcpbridge.New(ctx, filepath.Join(home, "bin", "hackalem-mcp"), s.root)
	if err != nil {
		s.errorMessage = "MCP не запущен: проверьте установленный плагин."
		return s
	}
	s.bridge = client
	s.client = client
	result, err := client.Call(ctx, "system.context", nil)
	if !success(result, err) {
		s.errorMessage = "MCP не прошёл проверку контекста."
		client.Close()
		s.client = nil
		return s
	}
	// Let the OS choose an unused loopback port; never expose the plugin control API.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		s.errorMessage = "Не удалось выделить локальный порт Fly demo."
		return s
	}
	port := fmt.Sprint(l.Addr().(*net.TCPAddr).Port)
	l.Close()
	cmd := exec.CommandContext(ctx, filepath.Join(home, "bin", "hackalem-demo-api"))
	cmd.Dir = s.root
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	for _, v := range os.Environ() {
		k, _, _ := strings.Cut(v, "=")
		if !strings.HasPrefix(k, "HACKALEM_DEMO_") && k != "HACKALEM_REPO_ROOT" && k != "OPENAI_API_KEY" && k != "NVIDIA_API_KEY" && k != "POSTGRES_DSN" {
			cmd.Env = append(cmd.Env, v)
		}
	}
	cmd.Env = append(cmd.Env, "HACKALEM_REPO_ROOT="+s.root, "HACKALEM_DEMO_HOST=127.0.0.1", "HACKALEM_DEMO_PORT="+port, "HACKALEM_DEMO_TOKEN=", "OPENAI_API_KEY=", "NVIDIA_API_KEY=", "POSTGRES_DSN=")
	if cmd.Start() != nil {
		s.errorMessage = "MCP подключён, но Fly demo не запущен."
		return s
	}
	s.demo = cmd
	s.demoDone = make(chan struct{})
	go func() { _ = cmd.Wait(); close(s.demoDone) }()
	target, _ := url.Parse("http://127.0.0.1:" + port)
	httpClient := &http.Client{Timeout: 300 * time.Millisecond}
	ready := false
	for i := 0; i < 30; i++ {
		resp, e := httpClient.Get(target.String() + "/healthz")
		if e == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				ready = true
				break
			}
		}
		select {
		case <-ctx.Done():
			return s
		case <-s.demoDone:
			s.errorMessage = "Процесс Fly demo завершился до готовности."
			return s
		case <-time.After(50 * time.Millisecond):
		}
	}
	if !ready {
		s.errorMessage = "Fly demo не прошёл healthcheck."
		_ = cmd.Process.Kill()
		<-s.demoDone
		return s
	}
	if !flyVisualizationReady(ctx, httpClient, target) {
		s.errorMessage = "Fly demo не содержит совместимую страницу /mb-3d. Обновите бинарные файлы HackAlem."
		_ = cmd.Process.Kill()
		<-s.demoDone
		return s
	}
	s.proxy = httputil.NewSingleHostReverseProxy(target)
	director := s.proxy.Director
	s.proxy.Director = func(r *http.Request) {
		director(r)
		if r.URL.Path == "/fly-demo" {
			r.URL.Path = "/mb-3d"
		}
		r.Header.Del("Cookie")
		r.Header.Del("Authorization")
	}
	s.proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		http.Error(w, "Fly demo временно недоступен", 503)
	}
	s.proxy.ModifyResponse = func(resp *http.Response) error {
		if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
			return nil
		}
		content, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		resp.Body.Close()
		if err != nil {
			return err
		}
		html := strings.ReplaceAll(string(content), "https://cdn.jsdelivr.net/npm/three@0.160.0/build/three.module.js", "/vendor/three.module.js")
		html = strings.ReplaceAll(html, `<a href="/fly-body">← к мухе</a>`, `схема HackAlem Fly`)
		resp.Body = io.NopCloser(strings.NewReader(html))
		resp.ContentLength = int64(len(html))
		resp.Header.Set("Content-Length", strconv.Itoa(len(html)))
		return nil
	}
	return s
}

func flyVisualizationReady(ctx context.Context, client *http.Client, target *url.URL) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String()+"/mb-3d", nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK && strings.Contains(resp.Header.Get("Content-Type"), "text/html")
}

func success(result map[string]any, err error) bool {
	return err == nil && result != nil && result["status"] == "ok"
}
func (s *Services) Health() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	connected := s.connected()
	flyAvailable := s.proxy != nil && s.ctx.Err() == nil
	if s.demoDone != nil {
		select {
		case <-s.demoDone:
			flyAvailable = false
		default:
		}
	}
	errorMessage := s.errorMessage
	if s.ctx.Err() != nil {
		errorMessage = "Интеграции остановлены."
	} else if s.client != nil && !connected {
		errorMessage = "Соединение с MCP закрыто. Перезапустите приложение."
	} else if s.proxy != nil && !flyAvailable {
		errorMessage = "Процесс Fly demo остановлен. Перезапустите приложение."
	}
	return map[string]any{"configured": s.configured, "connected": connected, "fly_available": flyAvailable, "error": errorMessage}
}
func (s *Services) connected() bool {
	return s.client != nil && s.ctx.Err() == nil && (s.bridge == nil || s.bridge.Running())
}
func (s *Services) TelegramConfigured() bool {
	return s.connected() && strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")) != "" && strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_ID")) != ""
}
func (s *Services) Close() {
	s.closeOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		if s.demo != nil {
			_ = s.demo.Process.Kill()
			<-s.demoDone
		}
		if s.bridge != nil {
			_ = s.bridge.Close()
		}
		if s.workerDone != nil {
			<-s.workerDone
		}
	})
}
func (s *Services) enqueue(fn func()) {
	if !s.connected() {
		return
	}
	select {
	case s.queue <- fn:
	default:
		s.mu.Lock()
		s.errorMessage = "Очередь MCP заполнена; протокол сохранён, события Fly пропущены."
		s.mu.Unlock()
	}
}
func (s *Services) call(tool string, args map[string]any) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(s.ctx, 25*time.Second)
	defer cancel()
	return s.client.Call(ctx, tool, args)
}

const auditTask = "Create an audit report for a meeting protocol that was reviewed by a human."

func (s *Services) Created(meeting map[string]any) {
	s.enqueue(func() {
		_, _ = s.call("routing.fly_recommend", map[string]any{"task": auditTask, "available_capabilities": []string{"reports.create"}, "input_metadata": map[string]any{"type": "meeting_review", "mode": meeting["mode"]}})
	})
}
func (s *Services) Reviewed(meeting, event map[string]any) {
	s.enqueue(func() {
		// Only anonymous workflow metadata reaches Fly; meeting content stays private.
		result, err := s.call("reports.create", map[string]any{"title": "Butterfly review audit", "format": "md", "sections": []map[string]string{{"heading": "Human review", "body": fmt.Sprintf("Mode: %v. Status: completed. Action count: %d. No transcript or personal data included.", meeting["mode"], actionCount(meeting))}}})
		if !success(result, err) {
			return
		}
		_, _ = s.call("routing.fly_feedback", map[string]any{"task": auditTask, "chosen": []string{"reports.create"}, "source": "normal-teacher", "input_metadata": map[string]any{"type": "meeting_review", "mode": meeting["mode"]}})
	})
}
func actionCount(m map[string]any) int {
	switch a := m["actions"].(type) {
	case []any:
		return len(a)
	case []map[string]any:
		return len(a)
	}
	return 0
}

func (s *Services) Register(mux *http.ServeMux, application *app.App) {
	for _, path := range []string{"/fly-demo", "/fly-live", "/fly-events", "/fly-feedback", "/fly-neuron-events"} {
		mux.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) {
			if _, err := application.Owner(r); err != nil {
				http.Error(w, "Требуется сессия", 401)
				return
			}
			if s.proxy == nil {
				http.Error(w, "Fly demo недоступен. Настройте установленный HackAlem через HACKALEM_HOME.", 503)
				return
			}
			s.proxy.ServeHTTP(w, r)
		})
	}
	mux.HandleFunc("POST /api/meetings/{id}/notify", func(w http.ResponseWriter, r *http.Request) {
		owner, err := application.Owner(r)
		if err != nil {
			reply(w, 401, map[string]any{"detail": "Требуется сессия."})
			return
		}
		meeting, err := application.Get(owner, r.PathValue("id"))
		if err != nil {
			reply(w, 404, map[string]any{"detail": "Встреча не найдена."})
			return
		}
		if meeting["status"] != "completed" {
			reply(w, 409, map[string]any{"detail": "Сначала подтвердите протокол."})
			return
		}
		if !s.TelegramConfigured() {
			reply(w, 503, map[string]any{"detail": "Telegram не настроен."})
			return
		}
		// Exclusive durable reservation prevents duplicate sends, including uncertain
		// failures. A retry must be an explicit operational decision, never automatic.
		id, _ := meeting["id"].(string)
		name := filepath.Join(s.root, "telegram-"+id+".json")
		f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if os.IsExist(err) {
			reply(w, 409, map[string]any{"detail": "Отправка уже выполнялась. Проверьте чат перед повторением."})
			return
		}
		if err != nil {
			reply(w, 500, map[string]any{"detail": "Не удалось сохранить состояние отправки."})
			return
		}
		_, err = f.WriteString(`{"status":"sending"}`)
		closeErr := f.Close()
		if err != nil || closeErr != nil {
			reply(w, 500, map[string]any{"detail": "Не удалось сохранить состояние отправки."})
			return
		}
		text := "Butterfly: протокол подтверждён. Откройте приложение в том же браузере, чтобы скачать PDF/DOCX."
		if meeting["mode"] == "demo" {
			text = "ДЕМО · синтетическая встреча. " + text
		}
		result, err := s.call("telegram.send", map[string]any{"text": text})
		state := "uncertain"
		if success(result, err) {
			state = "sent"
		}
		_ = os.WriteFile(name, []byte(fmt.Sprintf(`{"status":%q}`, state)), 0600)
		if state != "sent" {
			reply(w, 502, map[string]any{"detail": "Telegram не подтвердил доставку. Проверьте чат; автоматического повтора нет."})
			return
		}
		reply(w, 200, map[string]any{"ok": true, "status": "sent"})
	})
}
func reply(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
