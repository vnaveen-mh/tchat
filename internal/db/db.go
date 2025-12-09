package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/firebase/genkit/go/ai"
	_ "modernc.org/sqlite"
)

// ConversationTurn represents a single user input and model response
type ConversationTurn struct {
	SessionId    string
	MsgId        int64
	UserInput    string
	ModelOutput  string
	DurationMs   int64
	TTFCMs       int64
	Chunks       int
	InputLength  int
	OutputLength int
	Timestamp    time.Time
}

// Session represents a chat session record.
type Session struct {
	SessionId string
	Title     string
	ModelName string
	CreatedAt time.Time
}

// Store handles database operations
type Store struct {
	db *sql.DB
}

// New creates a new Store instance and initializes the database
func New(dbPath string) (*Store, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(ON)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	s := &Store{db: db}

	// Initialize schema
	if err := s.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return s, nil
}

// initSchema creates the necessary tables
func (s *Store) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS chat_sessions (
		session_id TEXT PRIMARY KEY,
		title TEXT,
		model_name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

CREATE TABLE IF NOT EXISTS chat_messages (
		msg_id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		user_input TEXT NOT NULL,
		llm_response TEXT NOT NULL,
		duration_ms INTEGER NOT NULL,
		ttfc_ms INTEGER,
		chunks INTEGER,
		input_length INTEGER,
		output_length INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(session_id) REFERENCES chat_sessions(session_id) ON DELETE CASCADE
	);	


	CREATE TABLE IF NOT EXISTS chat_history (
		id 			INTEGER PRIMARY KEY AUTOINCREMENT,
		role 		TEXT NOT NULL,
		content 	TEXT NOT NULL,
		created_at 	DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := s.db.Exec(schema); err != nil {
		return err
	}

	return nil
}

// CreateSession inserts a new chat session.
func (s *Store) CreateSession(session Session) error {
	query := `
		INSERT INTO chat_sessions (session_id, title, model_name)
		VALUES (?, ?, ?)
	`
	if _, err := s.db.Exec(query, session.SessionId, session.Title, session.ModelName); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// GetSessionByID retrieves a chat session by ID.
func (s *Store) GetSessionByID(sessionID string) (*Session, error) {
	query := `
		SELECT session_id, title, model_name, created_at
		FROM chat_sessions
		WHERE session_id = ?
	`

	var sess Session
	var createdAt sql.NullTime

	if err := s.db.QueryRow(query, sessionID).Scan(&sess.SessionId, &sess.Title, &sess.ModelName, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if createdAt.Valid {
		sess.CreatedAt = createdAt.Time
	}

	return &sess, nil
}

// ListSessions returns recent sessions ordered by creation time descending.
func (s *Store) ListSessions(limit int) ([]Session, error) {
	if limit <= 0 {
		return []Session{}, nil
	}

	query := `
		SELECT session_id, title, model_name, created_at
		FROM chat_sessions
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var sess Session
		var createdAt sql.NullTime
		if err := rows.Scan(&sess.SessionId, &sess.Title, &sess.ModelName, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		if createdAt.Valid {
			sess.CreatedAt = createdAt.Time
		}
		sessions = append(sessions, sess)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate sessions: %w", err)
	}

	return sessions, nil
}

// SaveTurn saves a conversation turn to the database
func (s *Store) SaveTurn(turn ConversationTurn) (int64, error) {
	query := `
		INSERT INTO chat_messages (
			session_id, user_input, llm_response, duration_ms, ttfc_ms, chunks, input_length, output_length
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(query,
		turn.SessionId,
		turn.UserInput,
		turn.ModelOutput,
		turn.DurationMs,
		turn.TTFCMs,
		turn.Chunks,
		turn.InputLength,
		turn.OutputLength,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert conversation: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// GetByID retrieves a single conversation by Message ID (or Turn ID)
func (s *Store) GetByMsgID(id int64) (*ConversationTurn, error) {
	query := `
		SELECT msg_id, session_id, user_input, llm_response,
		       duration_ms, ttfc_ms, chunks, input_length, output_length,
			   created_at
		FROM chat_messages
		WHERE msg_id = ?
	`

	var turn ConversationTurn
	var ttfcMs, chunks sql.NullInt64
	var createdAt sql.NullTime

	err := s.db.QueryRow(query, id).Scan(
		&turn.MsgId,
		&turn.SessionId,
		&turn.UserInput,
		&turn.ModelOutput,
		&turn.DurationMs,
		&ttfcMs,
		&chunks,
		&turn.InputLength,
		&turn.OutputLength,
		&createdAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("chat message not found")
		}
		return nil, fmt.Errorf("failed to get chat message: %w", err)
	}

	if ttfcMs.Valid {
		turn.TTFCMs = ttfcMs.Int64
	}
	if chunks.Valid {
		turn.Chunks = int(chunks.Int64)
	}

	if createdAt.Valid {
		turn.Timestamp = createdAt.Time
	}

	return &turn, nil
}

// GetRecentMessages retrieves the most recent N messages ordered by creation time descending.
func (s *Store) GetRecentMessages(limit int) ([]ConversationTurn, error) {
	if limit <= 0 {
		return []ConversationTurn{}, nil
	}

	query := `
		SELECT msg_id, session_id, user_input, llm_response,
		       duration_ms, ttfc_ms, chunks, input_length, output_length,
		       created_at
		FROM chat_messages
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent messages: %w", err)
	}
	defer rows.Close()

	var msgs []ConversationTurn
	for rows.Next() {
		var turn ConversationTurn
		var ttfcMs, chunks sql.NullInt64
		var createdAt sql.NullTime

		if err := rows.Scan(
			&turn.MsgId,
			&turn.SessionId,
			&turn.UserInput,
			&turn.ModelOutput,
			&turn.DurationMs,
			&ttfcMs,
			&chunks,
			&turn.InputLength,
			&turn.OutputLength,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan recent message: %w", err)
		}

		if ttfcMs.Valid {
			turn.TTFCMs = ttfcMs.Int64
		}
		if chunks.Valid {
			turn.Chunks = int(chunks.Int64)
		}
		if createdAt.Valid {
			turn.Timestamp = createdAt.Time
		}

		msgs = append(msgs, turn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate recent messages: %w", err)
	}

	return msgs, nil
}

// GetByDateRange retrieves conversations within a date range
func (s *Store) GetByDateRange(start, end time.Time) ([]ConversationTurn, error) {
	query := `
		SELECT msg_id, session_id, user_input, llm_response,
		       duration_ms, ttfc_ms, chunks, input_length, output_length,
		       created_at
		FROM chat_messages
		WHERE created_at BETWEEN ? AND ?
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages by date range: %w", err)
	}
	defer rows.Close()

	var msgs []ConversationTurn
	for rows.Next() {
		var turn ConversationTurn
		var ttfcMs, chunks sql.NullInt64
		var createdAt sql.NullTime

		if err := rows.Scan(
			&turn.MsgId,
			&turn.SessionId,
			&turn.UserInput,
			&turn.ModelOutput,
			&turn.DurationMs,
			&ttfcMs,
			&chunks,
			&turn.InputLength,
			&turn.OutputLength,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if ttfcMs.Valid {
			turn.TTFCMs = ttfcMs.Int64
		}
		if chunks.Valid {
			turn.Chunks = int(chunks.Int64)
		}
		if createdAt.Valid {
			turn.Timestamp = createdAt.Time
		}

		msgs = append(msgs, turn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate messages: %w", err)
	}

	return msgs, nil
}

// GetMessagesBySession returns messages for a session, ordered by creation time ascending.
func (s *Store) GetMessagesBySession(sessionID string, limit, offset int) ([]ConversationTurn, error) {
	if limit <= 0 {
		return []ConversationTurn{}, nil
	}

	query := `
		SELECT msg_id, session_id, user_input, llm_response,
		       duration_ms, ttfc_ms, chunks, input_length, output_length,
		       created_at
		FROM chat_messages
		WHERE session_id = ?
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, sessionID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages by session: %w", err)
	}
	defer rows.Close()

	var msgs []ConversationTurn
	for rows.Next() {
		var turn ConversationTurn
		var ttfcMs, chunks sql.NullInt64
		var createdAt sql.NullTime

		if err := rows.Scan(
			&turn.MsgId,
			&turn.SessionId,
			&turn.UserInput,
			&turn.ModelOutput,
			&turn.DurationMs,
			&ttfcMs,
			&chunks,
			&turn.InputLength,
			&turn.OutputLength,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if ttfcMs.Valid {
			turn.TTFCMs = ttfcMs.Int64
		}
		if chunks.Valid {
			turn.Chunks = int(chunks.Int64)
		}
		if createdAt.Valid {
			turn.Timestamp = createdAt.Time
		}

		msgs = append(msgs, turn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate messages by session: %w", err)
	}

	return msgs, nil
}

// GetStats returns statistics about stored conversations
func (s *Store) GetStats() (map[string]interface{}, error) {
	query := `
		SELECT
			(SELECT COUNT(*) FROM chat_sessions) as total_sessions,
			(SELECT COUNT(DISTINCT model_name) FROM chat_sessions) as unique_models,
			(SELECT COUNT(*) FROM chat_messages) as total_messages,
			(SELECT AVG(duration_ms) FROM chat_messages) as avg_duration_ms,
			(SELECT MIN(duration_ms) FROM chat_messages) as min_duration_ms,
			(SELECT MAX(duration_ms) FROM chat_messages) as max_duration_ms,
			(SELECT AVG(input_length) FROM chat_messages) as avg_input_length,
			(SELECT AVG(output_length) FROM chat_messages) as avg_output_length
	`

	var stats map[string]interface{} = make(map[string]interface{})
	var totalSessions, uniqueModels, totalMessages int
	var avgDuration, minDuration, maxDuration sql.NullFloat64
	var avgInputLen, avgOutputLen sql.NullFloat64

	err := s.db.QueryRow(query).Scan(
		&totalSessions,
		&uniqueModels,
		&totalMessages,
		&avgDuration,
		&minDuration,
		&maxDuration,
		&avgInputLen,
		&avgOutputLen,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query stats: %w", err)
	}

	stats["total_conversations"] = totalSessions
	stats["total_messages"] = totalMessages
	stats["unique_models"] = uniqueModels
	if avgDuration.Valid {
		stats["avg_duration_ms"] = avgDuration.Float64
	}
	if minDuration.Valid {
		stats["min_duration_ms"] = minDuration.Float64
	}
	if maxDuration.Valid {
		stats["max_duration_ms"] = maxDuration.Float64
	}
	if avgInputLen.Valid {
		stats["avg_input_length"] = avgInputLen.Float64
	}
	if avgOutputLen.Valid {
		stats["avg_output_length"] = avgOutputLen.Float64
	}

	return stats, nil
}

// SaveHistory saves history messages in chat_history table
func (s *Store) SaveHistory(ctx context.Context, messages []*ai.Message) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// wipe existing history snapshot
	if _, err := tx.ExecContext(ctx, `DELETE FROM chat_history`); err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO chat_history (role, content)
        VALUES (?, ?)
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, m := range messages {
		b, err := json.Marshal(m.Content)
		if err != nil {
			return err
		}
		if _, err := stmt.ExecContext(ctx, m.Role, string(b)); err != nil {
			return err
		}

	}
	return tx.Commit()
}

// Load loads the conversation history and system prompt from the history table
func (s *Store) LoadHistory(ctx context.Context) ([]*ai.Message, error) {
	rows, err := s.db.QueryContext(ctx, `
        SELECT role, content
        FROM chat_history
        ORDER BY id ASC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	msgs := []*ai.Message{}

	for rows.Next() {
		var role, content string
		if err := rows.Scan(&role, &content); err != nil {
			return nil, err
		}

		var data = []*ai.Part{}
		if err := json.Unmarshal([]byte(content), &data); err != nil {
			return nil, err
		}
		msg := ai.NewMessage(ai.Role(role), nil, data...)
		msgs = append(msgs, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return msgs, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}
