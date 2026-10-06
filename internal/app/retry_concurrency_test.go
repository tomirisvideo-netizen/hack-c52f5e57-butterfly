package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRetryDoesNotBlockOtherMeetings(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		payload, _ := json.Marshal(map[string]any{"summary": "Проверено", "actions": []any{}})
		writeJSON(w, 200, map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"content": string(payload)}}}})
	}))
	defer provider.Close()
	t.Setenv("NEMOTRON_BASE_URL", provider.URL+"/v1")
	t.Setenv("NEMOTRON_MODEL", "test-model")
	a, mux := testServer(t)
	cookie := session(t, mux)
	owner := hashToken(cookie.Value)
	meeting := newMeeting("Повтор", "ru", "2026-10-07", "manual")
	meeting["status"] = "failed"
	meeting["transcript"] = []any{map[string]any{"id": "s1", "text": "Обсудили план."}}
	if err := a.create(owner, meeting); err != nil {
		t.Fatal(err)
	}
	other := newMeeting("Другая встреча", "ru", "2026-10-07", "manual")
	other["status"] = "needs_review"
	if err := a.create(owner, other); err != nil {
		t.Fatal(err)
	}
	path := "/api/meetings/" + text(meeting["id"])
	retried := make(chan *httptest.ResponseRecorder, 1)
	go func() { retried <- request(t, mux, "POST", path+"/retry", nil, cookie, "") }()
	defer func() {
		close(release)
		select {
		case got := <-retried:
			if got.Code != 200 {
				t.Errorf("retry failed after provider completed: %d", got.Code)
			}
		case <-time.After(3 * time.Second):
			t.Error("retry did not finish after releasing provider")
		}
	}()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("retry did not reach provider")
	}
	checks := []struct {
		method, path, body string
		status             int
	}{
		{"PATCH", "/api/meetings/" + text(other["id"]), `{"summary":"Сохранено во время распознавания"}`, 200},
		{"POST", path + "/retry", "", 409},
		{"PATCH", path, `{"summary":"Нельзя перезаписать обработку"}`, 409},
		{"DELETE", path, "", 409},
	}
	for _, check := range checks {
		done := make(chan *httptest.ResponseRecorder, 1)
		go func() {
			done <- request(t, mux, check.method, check.path, strings.NewReader(check.body), cookie, "application/json")
		}()
		select {
		case got := <-done:
			if got.Code != check.status {
				t.Fatalf("%s %s: status %d, want %d", check.method, check.path, got.Code, check.status)
			}
		case <-time.After(time.Second):
			t.Fatalf("%s %s blocked behind slow provider", check.method, check.path)
		}
	}
}
