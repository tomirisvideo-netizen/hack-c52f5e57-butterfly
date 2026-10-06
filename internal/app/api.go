package app

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

func (a *App) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/session", a.session)
	mux.HandleFunc("GET /api/meetings", a.listMeetings)
	mux.HandleFunc("POST /api/meetings", a.createMeeting)
	mux.HandleFunc("POST /api/demo", a.demo)
	mux.HandleFunc("GET /api/meetings/{id}", a.readMeeting)
	mux.HandleFunc("PATCH /api/meetings/{id}", a.editMeeting)
	mux.HandleFunc("DELETE /api/meetings/{id}", a.deleteMeeting)
	mux.HandleFunc("POST /api/meetings/{id}/confirm", a.confirmMeeting)
	mux.HandleFunc("POST /api/meetings/{id}/retry", a.retryMeeting)
	mux.HandleFunc("GET /api/meetings/{id}/audio", a.audio)
	mux.HandleFunc("GET /api/meetings/{id}/export", a.export)
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func fail(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"detail": message})
}
func (a *App) authenticated(w http.ResponseWriter, r *http.Request) (string, bool) {
	owner, err := a.Owner(r)
	if err != nil {
		fail(w, 401, err.Error())
		return "", false
	}
	return owner, true
}
func (a *App) owned(w http.ResponseWriter, r *http.Request) (string, map[string]any, bool) {
	owner, ok := a.authenticated(w, r)
	if !ok {
		return "", nil, false
	}
	meeting, err := a.Get(owner, r.PathValue("id"))
	if err != nil {
		fail(w, 404, ErrNotFound.Error())
		return "", nil, false
	}
	return owner, meeting, true
}
func (a *App) saveHTTP(w http.ResponseWriter, owner string, meeting map[string]any) bool {
	if err := a.Save(owner, meeting); err != nil {
		fail(w, 500, "Не удалось сохранить протокол.")
		return false
	}
	return true
}
func (a *App) session(w http.ResponseWriter, r *http.Request) {
	token := ""
	if cookie, err := r.Cookie(cookieName); err == nil {
		if _, err = a.Owner(r); err == nil {
			token = cookie.Value
		}
	}
	if token == "" {
		token = randomID()
		a.mu.Lock()
		owner := hashToken(token)
		a.state.Sessions[owner] = true
		err := a.persistLocked()
		if err != nil {
			delete(a.state.Sessions, owner)
		}
		a.mu.Unlock()
		if err != nil {
			fail(w, 500, "Не удалось создать сессию.")
			return
		}
	}
	sameSite := http.SameSiteLaxMode
	switch strings.ToLower(os.Getenv("BUTTERFLY_COOKIE_SAMESITE")) {
	case "none":
		sameSite = http.SameSiteNoneMode
	case "strict":
		sameSite = http.SameSiteStrictMode
	}
	secure := sameSite == http.SameSiteNoneMode || os.Getenv("BUTTERFLY_SECURE_COOKIE") == "1" || strings.EqualFold(os.Getenv("BUTTERFLY_SECURE_COOKIE"), "true") || r.TLS != nil
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: token, Path: "/", MaxAge: 30 * 24 * 3600, HttpOnly: true, Secure: secure, SameSite: sameSite})
	writeJSON(w, 200, map[string]any{"ok": true})
}
func (a *App) listMeetings(w http.ResponseWriter, r *http.Request) {
	owner, ok := a.authenticated(w, r)
	if !ok {
		return
	}
	meetings := a.List(owner)
	sort.Slice(meetings, func(i, j int) bool { return text(meetings[i]["created_at"]) > text(meetings[j]["created_at"]) })
	if len(meetings) > 100 {
		meetings = meetings[:100]
	}
	for i, m := range meetings {
		meetings[i] = publicMeeting(m)
	}
	writeJSON(w, 200, map[string]any{"meetings": meetings})
}
func newMeeting(title, language, date, mode string) map[string]any {
	return map[string]any{"id": randomID(), "title": title, "created_at": timestamp(), "date": date, "language": language, "status": "processing", "mode": mode, "summary": "", "transcript": []any{}, "actions": []any{}, "events": []any{}}
}

var allowedExtensions = map[string]bool{".wav": true, ".mp3": true, ".m4a": true, ".ogg": true, ".flac": true, ".webm": true, ".mp4": true, ".aac": true, ".opus": true}

func (a *App) createMeeting(w http.ResponseWriter, r *http.Request) {
	owner, ok := a.authenticated(w, r)
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes+(1<<20))
	var err error
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		err = r.ParseMultipartForm(1 << 20)
		if r.MultipartForm != nil {
			defer r.MultipartForm.RemoveAll()
		}
	} else {
		err = r.ParseForm()
	}
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			fail(w, 413, "Максимальный размер аудио — 50 МБ.")
		} else {
			fail(w, 422, "Некорректная форма загрузки.")
		}
		return
	}
	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		title = "Новая встреча"
	}
	if utf8.RuneCountInString(title) > 200 {
		fail(w, 422, "Название длиннее 200 символов.")
		return
	}
	language := r.FormValue("language")
	if language == "" {
		language = "auto"
	}
	if language != "auto" && language != "ru" && language != "kk" && language != "mixed" {
		fail(w, 422, "Неподдерживаемый язык.")
		return
	}
	date := r.FormValue("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	if _, err = time.Parse("2006-01-02", date); err != nil {
		fail(w, 422, "Дата должна иметь формат YYYY-MM-DD.")
		return
	}
	transcript := strings.TrimSpace(r.FormValue("transcript"))
	if utf8.RuneCountInString(transcript) > maxTranscriptChars {
		fail(w, 413, "Транскрипт превышает 300 000 символов.")
		return
	}
	var file multipart.File
	var header *multipart.FileHeader
	fileErr := http.ErrMissingFile
	if r.MultipartForm != nil {
		file, header, fileErr = r.FormFile("file")
	}
	if fileErr != nil && fileErr != http.ErrMissingFile {
		fail(w, 422, "Не удалось прочитать аудио.")
		return
	}
	if fileErr == nil {
		defer file.Close()
	}
	if transcript == "" && fileErr == http.ErrMissingFile {
		fail(w, 422, "Загрузите аудио или добавьте текст встречи.")
		return
	}
	mode := "live"
	if transcript != "" {
		mode = "manual"
	}
	meeting := newMeeting(title, language, date, mode)
	uploaded := ""
	if fileErr == nil {
		suffix := strings.ToLower(filepath.Ext(header.Filename))
		if !allowedExtensions[suffix] {
			fail(w, 415, "Поддерживаются WAV, MP3, M4A, OGG, FLAC, WebM, MP4, AAC и Opus.")
			return
		}
		name := text(meeting["id"]) + suffix
		uploaded = filepath.Join(a.dataDir, "uploads", name)
		out, openErr := os.OpenFile(uploaded, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if openErr != nil {
			fail(w, 500, "Не удалось сохранить аудио.")
			return
		}
		size, copyErr := io.Copy(out, io.LimitReader(file, maxUploadBytes+1))
		closeErr := out.Close()
		if copyErr != nil || closeErr != nil || size == 0 || size > maxUploadBytes {
			_ = os.Remove(uploaded)
			status := 422
			message := "Аудиофайл пуст или недоступен."
			if size > maxUploadBytes {
				status = 413
				message = "Максимальный размер аудио — 50 МБ."
			}
			fail(w, status, message)
			return
		}
		meeting["_audio_file"] = name
		meeting["audio_url"] = "/api/meetings/" + text(meeting["id"]) + "/audio"
	}
	if transcript != "" {
		segments := []any{}
		for _, line := range strings.Split(transcript, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				segments = append(segments, map[string]any{"id": randomID(), "start": nil, "end": nil, "speaker": "Текст пользователя", "text": line})
			}
		}
		meeting["transcript"] = segments
	}
	addEvent(meeting, "created", "Встреча загружена и сохранена.")
	if err = a.create(owner, meeting); err != nil {
		if uploaded != "" {
			_ = os.Remove(uploaded)
		}
		fail(w, 500, "Не удалось сохранить встречу.")
		return
	}
	if a.OnCreated != nil {
		a.OnCreated(publicMeeting(meeting))
	}
	if err = a.process(r, owner, meeting); err != nil {
		fail(w, 500, "Не удалось сохранить результат обработки.")
		return
	}
	writeJSON(w, 201, publicMeeting(meeting))
}
func (a *App) demo(w http.ResponseWriter, r *http.Request) {
	owner, ok := a.authenticated(w, r)
	if !ok {
		return
	}
	fixture := map[string]any{}
	if err := json.Unmarshal([]byte(demoJSON), &fixture); err != nil {
		fail(w, 500, "Демо недоступно.")
		return
	}
	meeting := newMeeting(text(fixture["title"]), text(fixture["language"]), text(fixture["date"]), "demo")
	for key, value := range fixture {
		meeting[key] = value
	}
	meeting["status"] = "needs_review"
	addEvent(meeting, "demo_loaded", "Синтетический пример: заранее подготовленные данные, не результат работы моделей.")
	addEvent(meeting, "review_required", "Три поручения ожидают проверки; у одного ответственный и срок не определены.")
	if err := a.create(owner, meeting); err != nil {
		fail(w, 500, "Не удалось сохранить демо.")
		return
	}
	if a.OnCreated != nil {
		a.OnCreated(publicMeeting(meeting))
	}
	writeJSON(w, 201, publicMeeting(meeting))
}
func (a *App) readMeeting(w http.ResponseWriter, r *http.Request) {
	_, meeting, ok := a.owned(w, r)
	if ok {
		writeJSON(w, 200, publicMeeting(meeting))
	}
}
func (a *App) editMeeting(w http.ResponseWriter, r *http.Request) {
	a.mutationMu.Lock()
	defer a.mutationMu.Unlock()
	owner, meeting, ok := a.owned(w, r)
	if !ok {
		return
	}
	if meeting["status"] == "processing" {
		fail(w, 409, "Дождитесь окончания обработки.")
		return
	}
	var patch map[string]json.RawMessage
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20))
	if err := decoder.Decode(&patch); err != nil {
		fail(w, 422, "Некорректные изменения.")
		return
	}
	if len(patch) == 0 {
		fail(w, 422, "Укажите резюме или поручения.")
		return
	}
	for key := range patch {
		if key != "summary" && key != "actions" {
			fail(w, 422, "Неизвестное поле изменения.")
			return
		}
	}
	if raw, exists := patch["summary"]; exists {
		var summary string
		if err := json.Unmarshal(raw, &summary); err != nil || string(raw) == "null" || utf8.RuneCountInString(summary) > 30000 {
			fail(w, 422, "Некорректное резюме.")
			return
		}
		meeting["summary"] = summary
	}
	if raw, exists := patch["actions"]; exists {
		actions, err := parseActions(raw, entries(meeting["transcript"]))
		if err != nil {
			fail(w, 422, err.Error())
			return
		}
		meeting["actions"] = actions
	}
	meeting["status"] = "needs_review"
	addEvent(meeting, "edited", "Пользователь отредактировал протокол. Требуется итоговое подтверждение.")
	if a.saveHTTP(w, owner, meeting) {
		writeJSON(w, 200, publicMeeting(meeting))
	}
}
func (a *App) confirmMeeting(w http.ResponseWriter, r *http.Request) {
	a.mutationMu.Lock()
	defer a.mutationMu.Unlock()
	owner, meeting, ok := a.owned(w, r)
	if !ok {
		return
	}
	if meeting["status"] == "completed" {
		writeJSON(w, 200, publicMeeting(meeting))
		return
	}
	if meeting["status"] != "needs_review" {
		fail(w, 409, "Сначала подготовьте и проверьте протокол.")
		return
	}
	for _, item := range entries(meeting["actions"]) {
		if object(item)["status"] != "confirmed" {
			fail(w, 409, "Подтвердите каждое поручение, включая неизвестные ответственные и сроки.")
			return
		}
	}
	meeting["status"] = "completed"
	event := addEvent(meeting, "confirmed", "Пользователь подтвердил протокол и каждое поручение.")
	if !a.saveHTTP(w, owner, meeting) {
		return
	}
	if a.OnReview != nil {
		a.OnReview(publicMeeting(meeting), clone(event))
	}
	writeJSON(w, 200, publicMeeting(meeting))
}
func (a *App) retryMeeting(w http.ResponseWriter, r *http.Request) {
	a.mutationMu.Lock()
	owner, meeting, ok := a.owned(w, r)
	if !ok {
		a.mutationMu.Unlock()
		return
	}
	if meeting["mode"] == "demo" || meeting["status"] == "processing" || meeting["status"] == "completed" {
		a.mutationMu.Unlock()
		fail(w, 409, "Для этой встречи повторная обработка недоступна.")
		return
	}
	// Reserve this meeting under the mutation lock, then release it before
	// waiting for ASR/LLM so other meetings can still be reviewed or deleted.
	err := a.beginProcessing(owner, meeting)
	a.mutationMu.Unlock()
	if err != nil {
		fail(w, 500, "Не удалось сохранить обработку.")
		return
	}
	if err := a.finishProcessing(r, owner, meeting); err != nil {
		fail(w, 500, "Не удалось сохранить обработку.")
		return
	}
	writeJSON(w, 200, publicMeeting(meeting))
}
func (a *App) audio(w http.ResponseWriter, r *http.Request) {
	_, meeting, ok := a.owned(w, r)
	if !ok {
		return
	}
	name := text(meeting["_audio_file"])
	if name == "" || filepath.Base(name) != name {
		fail(w, 404, "У встречи нет аудио.")
		return
	}
	path := filepath.Join(a.dataDir, "uploads", name)
	file, err := os.Open(path)
	if err != nil {
		fail(w, 404, "Аудиофайл недоступен.")
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		fail(w, 404, "Аудиофайл недоступен.")
		return
	}
	kind := mime.TypeByExtension(filepath.Ext(name))
	if kind == "" {
		kind = "application/octet-stream"
	}
	w.Header().Set("Content-Type", kind)
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, name, info.ModTime(), file)
}
func (a *App) export(w http.ResponseWriter, r *http.Request) {
	_, meeting, ok := a.owned(w, r)
	if !ok {
		return
	}
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "pdf"
	}
	media := map[string]string{"json": "application/json", "pdf": "application/pdf", "docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document"}
	kind, valid := media[format]
	if !valid {
		fail(w, 422, "Неподдерживаемый формат экспорта.")
		return
	}
	if format != "json" && meeting["status"] != "completed" {
		fail(w, http.StatusConflict, "Сначала подтвердите протокол перед экспортом PDF/DOCX.")
		return
	}
	var payload []byte
	var err error
	if format == "json" {
		payload, err = json.MarshalIndent(publicMeeting(meeting), "", "  ")
	} else if a.Export != nil {
		payload, err = a.Export(publicMeeting(meeting), format)
	} else {
		fail(w, 503, "Экспорт в этом формате ещё не подключён.")
		return
	}
	if err != nil {
		fail(w, 500, "Не удалось сформировать документ.")
		return
	}
	w.Header().Set("Content-Type", kind)
	w.Header().Set("Content-Disposition", `attachment; filename="butterfly-`+text(meeting["id"])+`.`+format+`"`)
	w.Header().Set("Cache-Control", "private, no-store")
	_, _ = w.Write(payload)
}
func (a *App) deleteMeeting(w http.ResponseWriter, r *http.Request) {
	a.mutationMu.Lock()
	defer a.mutationMu.Unlock()
	owner, meeting, ok := a.owned(w, r)
	if !ok {
		return
	}
	if meeting["status"] == "processing" {
		fail(w, 409, "Дождитесь окончания обработки.")
		return
	}
	if err := a.delete(owner, text(meeting["id"])); err != nil {
		fail(w, 500, "Не удалось удалить встречу.")
		return
	}
	name := text(meeting["_audio_file"])
	if name != "" && filepath.Base(name) == name {
		_ = os.Remove(filepath.Join(a.dataDir, "uploads", name))
	}
	w.WriteHeader(204)
}
