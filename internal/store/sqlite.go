package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"hooklet/internal/model"
	"time"

	_ "modernc.org/sqlite"
)

 type SqliteStore struct {
	db *sql.DB
}

var _ Store = (*SqliteStore)(nil)

func NewSqliteStore(dbPath string)(*SqliteStore,error){
	db, err := sql.Open("sqlite", dbPath)
    
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %v", err)
	}

	s := &SqliteStore{db: db}

	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate db: %v", err)
	}
	return s, nil
}

func (s *SqliteStore) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS requests (
		id TEXT PRIMARY KEY,
		method TEXT NOT NULL,
		path TEXT NOT NULL,
		query_string TEXT NOT NULL,
		headers TEXT NOT NULL,
		body BLOB,
		content_type TEXT NOT NULL,
		content_length INTEGER NOT NULL,
		remote_addr TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_requests_created_at ON requests(created_at DESC);

	CREATE TABLE IF NOT EXISTS replay_attempts (
		id TEXT PRIMARY KEY,
		request_id TEXT NOT NULL,
		target_url TEXT NOT NULL,
		status_code INTEGER NOT NULL,
		response_body BLOB,
		latency_ms INTEGER NOT NULL,
		error TEXT,
		created_at DATETIME NOT NULL,
		FOREIGN KEY(request_id) REFERENCES requests(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_replays_request_id ON replay_attempts(request_id, created_at ASC);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *SqliteStore) Close() error{
	return s.db.Close()
}

func (s *SqliteStore) SaveRequest(ctx context.Context, req *model.WebhookRequest) error {
	
	headersJSON, err := json.Marshal(req.Headers)
	if err != nil {
		return fmt.Errorf("marshal header: %v", err)
	}
    
	query := `
	INSERT INTO requests (id, method, path, query_string, headers, body, content_type, content_length, remote_addr, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = s.db.ExecContext(
		ctx,
		query,
		req.ID,
		req.Method,
		req.Path,
		req.QueryString,
		string(headersJSON),
		req.Body,
		req.ContentType,
		req.ContentLength,
		req.RemoteAddr,
		req.CreatedAt.UTC().Format(time.RFC3339Nano),
	)

	if err != nil {
		return fmt.Errorf("insert request: %w", err)
	}

	return nil
}
func (s *SqliteStore) GetRequest(ctx context.Context, id string) (*model.WebhookRequest, error) {
	query := `SELECT id, method, path, query_string, headers, body, content_type, content_length, remote_addr, created_at
	FROM requests
	WHERE id = ?`

	row := s.db.QueryRowContext(ctx, query, id)
	var req model.WebhookRequest
	var headersJSON string
	var createdAtStr string
	err := row.Scan(
		&req.ID,
		&req.Method,
		&req.Path,
		&req.QueryString,
		&headersJSON,
		&req.Body,
		&req.ContentType,
		&req.ContentLength,
		&req.RemoteAddr,
		&createdAtStr,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan request: %w", err)
	}
	if err := json.Unmarshal([]byte(headersJSON), &req.Headers); err != nil {
		return nil, fmt.Errorf("unmarshal headers: %w", err)
	}
	t, err := time.Parse(time.RFC3339Nano, createdAtStr)
	if err != nil {
 		t, _ = time.Parse(time.RFC3339, createdAtStr)
	}
	req.CreatedAt = t

	replays, err := s.GetReplayAttempts(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get replay attempts: %w", err)
	}
	req.ReplayAttempts = replays

	return &req, nil
}

func (s *SqliteStore) ListRequests(ctx context.Context, limit int, offset int) ([]*model.WebhookRequest, error) {
	query := `
	SELECT id, method, path, query_string, headers, body, content_type, content_length, remote_addr, created_at
	FROM requests
	ORDER BY created_at DESC
	LIMIT ? OFFSET ?
	`
	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query list requests: %w", err)
	}
	defer rows.Close()

	var list []*model.WebhookRequest
	for rows.Next() {
		var req model.WebhookRequest
		var headersJSON string
		var createdAtStr string

		if err := rows.Scan(
			&req.ID,
			&req.Method,
			&req.Path,
			&req.QueryString,
			&headersJSON,
			&req.Body,
			&req.ContentType,
			&req.ContentLength,
			&req.RemoteAddr,
			&createdAtStr,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		_ = json.Unmarshal([]byte(headersJSON), &req.Headers)
		t, err := time.Parse(time.RFC3339Nano, createdAtStr)
		if err != nil {
			t, _ = time.Parse(time.RFC3339, createdAtStr)
		}
		req.CreatedAt = t

		list = append(list, &req)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return list, nil
}

func(s *SqliteStore) SaveReplayAttempt(ctx context.Context, attempt *model.ReplayAttempt)error{
	  query := `
	INSERT INTO replay_attempts (id, request_id, target_url, status_code, response_body, latency_ms, error, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(
		ctx,
		query,
		attempt.ID,
		attempt.RequestID,
		attempt.TargetURL,
		attempt.StatusCode,
		attempt.ResponseBody,
		attempt.LatencyMs,
		attempt.Error,
		attempt.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("insert replay attempt: %w", err)
	}
	return nil
}


func (s *SqliteStore) GetReplayAttempts(ctx context.Context, requestID string) ([]model.ReplayAttempt, error) {
	query := `
	SELECT id, request_id, target_url, status_code, response_body, latency_ms, error, created_at
	FROM replay_attempts
	WHERE request_id = ?
	ORDER BY created_at ASC
	`
	rows, err := s.db.QueryContext(ctx, query, requestID)
	if err != nil {
		return nil, fmt.Errorf("query replay attempts: %w", err)
	}
	defer rows.Close()

	var attempts []model.ReplayAttempt
	for rows.Next() {
		var att model.ReplayAttempt
		var createdAtStr string

		if err := rows.Scan(
			&att.ID,
			&att.RequestID,
			&att.TargetURL,
			&att.StatusCode,
			&att.ResponseBody,
			&att.LatencyMs,
			&att.Error,
			&createdAtStr,
		); err != nil {
			return nil, fmt.Errorf("scan replay attempt: %w", err)
		}

		t, err := time.Parse(time.RFC3339Nano, createdAtStr)
		if err != nil {
			t, _ = time.Parse(time.RFC3339, createdAtStr)
		}
		att.CreatedAt = t

		attempts = append(attempts, att)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("replay attempts iteration: %w", err)
	}

	return attempts, nil
}