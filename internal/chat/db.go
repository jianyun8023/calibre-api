package chat

import (
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DB 数据库操作
type DB struct {
	conn *sql.DB
}

// NewDB 创建数据库连接
func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 运行迁移
	if err := runMigrations(conn); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	return db.conn.Close()
}

// runMigrations 运行数据库迁移
func runMigrations(conn *sql.DB) error {
	// 读取迁移文件
	migrationFiles, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, file := range migrationFiles {
		if file.IsDir() {
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + file.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		if _, err := conn.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file.Name(), err)
		}
	}

	return nil
}

// CreateConversation 创建新对话
func (db *DB) CreateConversation(title string) (*Conversation, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := db.conn.Exec(
		"INSERT INTO conversations (id, title, created_at, updated_at) VALUES (?, ?, ?, ?)",
		id, title, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return &Conversation{
		ID:        id,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// GetConversation 获取对话
func (db *DB) GetConversation(id string) (*Conversation, error) {
	var conv Conversation
	err := db.conn.QueryRow(
		"SELECT id, title, created_at, updated_at FROM conversations WHERE id = ?",
		id,
	).Scan(&conv.ID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("conversation not found")
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return &conv, nil
}

// ListConversations 列出所有对话
func (db *DB) ListConversations(limit int) ([]*Conversation, error) {
	rows, err := db.conn.Query(
		"SELECT id, title, created_at, updated_at FROM conversations ORDER BY updated_at DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer rows.Close()

	var conversations []*Conversation
	for rows.Next() {
		var conv Conversation
		err := rows.Scan(&conv.ID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, &conv)
	}

	return conversations, nil
}

// DeleteConversation 删除对话
func (db *DB) DeleteConversation(id string) error {
	_, err := db.conn.Exec("DELETE FROM conversations WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}
	return nil
}

// UpdateConversationTitle 更新对话标题
func (db *DB) UpdateConversationTitle(id, title string) error {
	_, err := db.conn.Exec("UPDATE conversations SET title = ?, updated_at = ? WHERE id = ?", title, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update conversation title: %w", err)
	}
	return nil
}

// SaveMessage 保存消息
func (db *DB) SaveMessage(msg *Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	msg.CreatedAt = time.Now()

	_, err := db.conn.Exec(
		`INSERT INTO messages (id, conversation_id, role, content, thinking, metadata, created_at) 
         VALUES (?, ?, ?, ?, ?, ?, ?)`,
		msg.ID, msg.ConversationID, msg.Role, msg.Content, msg.Thinking, msg.Metadata, msg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	// 更新对话时间
	_, err = db.conn.Exec("UPDATE conversations SET updated_at = ? WHERE id = ?", time.Now(), msg.ConversationID)
	if err != nil {
		return fmt.Errorf("failed to update conversation timestamp: %w", err)
	}

	return nil
}

// GetConversationMessages 获取对话的所有消息
func (db *DB) GetConversationMessages(conversationID string) ([]*Message, error) {
	rows, err := db.conn.Query(
		"SELECT id, conversation_id, role, content, thinking, metadata, created_at FROM messages WHERE conversation_id = ? ORDER BY created_at ASC",
		conversationID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	messages := make([]*Message, 0)
	for rows.Next() {
		var msg Message
		var thinking, metadata sql.NullString
		err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &thinking, &metadata, &msg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		if thinking.Valid {
			msg.Thinking = thinking.String
		}
		if metadata.Valid {
			msg.Metadata = metadata.String
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

// DeleteMessage 删除消息
func (db *DB) DeleteMessage(id string) error {
	_, err := db.conn.Exec("DELETE FROM messages WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}
