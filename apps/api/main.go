package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type server struct {
	db       *pgxpool.Pool
	storage  string
	maxBytes int64
}
type job struct {
	ID        string  `json:"id"`
	Filename  string  `json:"filename"`
	Status    string  `json:"status"`
	Stage     string  `json:"stage"`
	CreatedAt string  `json:"createdAt"`
	Progress  int     `json:"progress"`
	Error     *string `json:"error,omitempty"`
}
type segment struct {
	ID      string  `json:"id"`
	Speaker string  `json:"speaker"`
	Text    string  `json:"text"`
	Start   float64 `json:"start"`
	End     float64 `json:"end"`
}

func main() {
	dbURL := env("DATABASE_URL", "postgres://voxtrace:voxtrace@localhost:5432/voxtrace?sslmode=disable")
	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err = db.Ping(context.Background()); err != nil {
		slog.Error("database unavailable", "error", err)
		os.Exit(1)
	}
	s := &server{db: db, storage: env("STORAGE_DIR", "../../storage/uploads"), maxBytes: 2 << 30}
	_ = os.MkdirAll(s.storage, 0750)
	r := chi.NewRouter()
	r.Use(cors, requestLog)
	r.Get("/api/health", s.health)
	r.Post("/api/jobs", s.createJob)
	r.Get("/api/jobs", s.listJobs)
	r.Get("/api/jobs/{id}", s.getJob)
	r.Get("/api/jobs/{id}/transcript", s.transcript)
	r.Get("/api/jobs/{id}/export", s.exportText)
	r.Post("/api/jobs/{id}/retry", s.retry)
	r.Post("/api/jobs/{id}/cancel", s.cancel)
	r.Delete("/api/jobs/{id}", s.deleteJob)
	addr := env("HTTP_ADDR", ":8080")
	slog.Info("VoxTrace API ready", "address", addr)
	if err = http.ListenAndServe(addr, r); err != nil {
		panic(err)
	}
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func (s *server) health(w http.ResponseWriter, r *http.Request) {
	write(w, 200, map[string]string{"status": "ok"})
}
func (s *server) createJob(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, s.maxBytes)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		write(w, 400, map[string]string{"error": "invalid or oversized upload"})
		return
	}
	f, h, err := r.FormFile("audio")
	if err != nil {
		write(w, 400, map[string]string{"error": "audio is required"})
		return
	}
	defer f.Close()
	if err = validateAudio(h); err != nil {
		write(w, 415, map[string]string{"error": err.Error()})
		return
	}
	id := uuid.NewString()
	ext := strings.ToLower(filepath.Ext(h.Filename))
	path := filepath.Join(s.storage, id+ext)
	out, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0640)
	if err != nil {
		write(w, 500, map[string]string{"error": "cannot store audio"})
		return
	}
	_, copyErr := io.Copy(out, f)
	closeErr := out.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(path)
		write(w, 500, map[string]string{"error": "cannot store audio"})
		return
	}
	_, err = s.db.Exec(r.Context(), `INSERT INTO recordings(id,original_filename,stored_path,mime_type,size_bytes) VALUES($1,$2,$3,$4,$5)`, id, filepath.Base(h.Filename), path, h.Header.Get("Content-Type"), h.Size)
	if err == nil {
		_, err = s.db.Exec(r.Context(), `INSERT INTO jobs(id,recording_id,status,stage,progress,model) VALUES($1,$1,'queued','waiting',0,'large-v3')`, id)
	}
	if err != nil {
		_ = os.Remove(path)
		write(w, 500, map[string]string{"error": "cannot create job"})
		return
	}
	write(w, 201, map[string]any{"id": id, "status": "queued"})
}
func validateAudio(h *multipart.FileHeader) error {
	allowed := map[string]bool{".mp3": true, ".wav": true, ".m4a": true, ".flac": true, ".ogg": true, ".webm": true}
	if !allowed[strings.ToLower(filepath.Ext(h.Filename))] {
		return errors.New("unsupported audio format")
	}
	return nil
}
func (s *server) listJobs(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query(r.Context(), `SELECT j.id,r.original_filename,j.status,j.stage,j.progress,j.created_at::text,j.error_message FROM jobs j JOIN recordings r ON r.id=j.recording_id ORDER BY j.created_at DESC LIMIT 100`)
	if err != nil {
		write(w, 500, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()
	out := []job{}
	for rows.Next() {
		var j job
		if rows.Scan(&j.ID, &j.Filename, &j.Status, &j.Stage, &j.Progress, &j.CreatedAt, &j.Error) == nil {
			out = append(out, j)
		}
	}
	write(w, 200, out)
}
func (s *server) getJob(w http.ResponseWriter, r *http.Request) {
	var j job
	err := s.db.QueryRow(r.Context(), `SELECT j.id,r.original_filename,j.status,j.stage,j.progress,j.created_at::text,j.error_message FROM jobs j JOIN recordings r ON r.id=j.recording_id WHERE j.id=$1`, chi.URLParam(r, "id")).Scan(&j.ID, &j.Filename, &j.Status, &j.Stage, &j.Progress, &j.CreatedAt, &j.Error)
	if err != nil {
		write(w, 404, map[string]string{"error": "job not found"})
		return
	}
	write(w, 200, j)
}
func (s *server) transcript(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var lang, text string
	var duration float64
	err := s.db.QueryRow(r.Context(), `SELECT language,duration_seconds,raw_text FROM transcripts WHERE job_id=$1`, id).Scan(&lang, &duration, &text)
	if err != nil {
		write(w, 404, map[string]string{"error": "transcript not ready"})
		return
	}
	rows, _ := s.db.Query(r.Context(), `SELECT id,speaker,start_time,end_time,text FROM segments WHERE transcript_id=$1 ORDER BY sequence`, id)
	defer rows.Close()
	ss := []segment{}
	for rows.Next() {
		var v segment
		_ = rows.Scan(&v.ID, &v.Speaker, &v.Start, &v.End, &v.Text)
		ss = append(ss, v)
	}
	write(w, 200, map[string]any{"jobId": id, "language": lang, "duration": duration, "text": text, "segments": ss})
}
func (s *server) exportText(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rows, err := s.db.Query(r.Context(), `SELECT speaker,start_time,end_time,text FROM segments WHERE transcript_id=$1 ORDER BY sequence`, id)
	if err != nil {
		http.Error(w, "not ready", 404)
		return
	}
	defer rows.Close()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="transcript.txt"`)
	for rows.Next() {
		var sp, t string
		var a, b float64
		_ = rows.Scan(&sp, &a, &b, &t)
		fmt.Fprintf(w, "[%s–%s] %s\n%s\n\n", clock(a), clock(b), sp, t)
	}
}
func clock(v float64) string { return fmt.Sprintf("%02d:%02d", int(v)/60, int(v)%60) }
func (s *server) retry(w http.ResponseWriter, r *http.Request) {
	tag, err := s.db.Exec(r.Context(), `UPDATE jobs SET status='queued',stage='waiting',progress=0,error_message=NULL,attempt_count=attempt_count+1 WHERE id=$1 AND status='failed'`, chi.URLParam(r, "id"))
	if err != nil || tag.RowsAffected() == 0 {
		write(w, 409, map[string]string{"error": "job is not retryable"})
		return
	}
	write(w, 202, map[string]string{"status": "queued"})
}
func (s *server) cancel(w http.ResponseWriter, r *http.Request) {
	tag, err := s.db.Exec(r.Context(), `UPDATE jobs SET status='cancelled',stage='cancelled',completed_at=now() WHERE id=$1 AND status IN ('queued','processing')`, chi.URLParam(r, "id"))
	if err != nil || tag.RowsAffected() == 0 {
		write(w, 409, map[string]string{"error": "job is not cancellable"})
		return
	}
	write(w, 200, map[string]string{"status": "cancelled"})
}
func (s *server) deleteJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var path string
	err := s.db.QueryRow(r.Context(), `SELECT r.stored_path FROM recordings r JOIN jobs j ON j.recording_id=r.id WHERE j.id=$1 AND j.status<>'processing'`, id).Scan(&path)
	if err != nil {
		write(w, 409, map[string]string{"error": "active or missing job"})
		return
	}
	_, _ = s.db.Exec(r.Context(), `DELETE FROM recordings WHERE id=$1`, id)
	_ = os.Remove(path)
	w.WriteHeader(204)
}
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", env("CORS_ORIGIN", "http://localhost:3000"))
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "ms", strconv.FormatInt(time.Since(start).Milliseconds(), 10))
	})
}
