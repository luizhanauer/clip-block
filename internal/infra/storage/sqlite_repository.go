package storage

import (
	"clip-block/internal/core/domain"
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository() (*SQLiteRepository, error) {
	dbPath := GetDatabasePath()
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	repo := &SQLiteRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, err
	}

	return repo, nil
}

func GetDatabasePath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".local", "share", "clip-block")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "clip-block.db")
}

func (r *SQLiteRepository) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS clips (
		id TEXT PRIMARY KEY,
		content TEXT NOT NULL,
		is_pinned BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteRepository) GetAll(page, pageSize int, pinned *bool) ([]domain.Clip, int, error) {
	var totalItems int
	countQuery := "SELECT COUNT(*) FROM clips"
	var countArgs []interface{}

	whereClause := ""
	if pinned != nil {
		whereClause = " WHERE is_pinned = ?"
		countArgs = append(countArgs, *pinned)
	}

	err := r.db.QueryRow(countQuery+whereClause, countArgs...).Scan(&totalItems)
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT id, content, is_pinned, created_at FROM clips" + whereClause + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args := append(countArgs, pageSize, (page-1)*pageSize)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var clips []domain.Clip
	for rows.Next() {
		var c domain.Clip
		if err := rows.Scan(&c.ID, &c.Content, &c.IsPinned, &c.CreatedAt); err != nil {
			return nil, 0, err
		}
		clips = append(clips, c)
	}

	if clips == nil {
		clips = []domain.Clip{}
	}

	return clips, totalItems, nil
}

func (r *SQLiteRepository) Save(clip domain.Clip) error {
	query := "INSERT INTO clips (id, content, is_pinned, created_at) VALUES (?, ?, ?, ?)"
	_, err := r.db.Exec(query, clip.ID, clip.Content, clip.IsPinned, clip.CreatedAt)
	return err
}

func (r *SQLiteRepository) Delete(id string) error {
	query := "DELETE FROM clips WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

func (r *SQLiteRepository) TogglePin(id string) error {
	query := "UPDATE clips SET is_pinned = NOT is_pinned WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

func (r *SQLiteRepository) DeleteOlderThan(before time.Time) (int, error) {
	query := "DELETE FROM clips WHERE created_at < ? AND is_pinned = 0"
	res, err := r.db.Exec(query, before)
	if err != nil {
		return 0, err
	}
	rows, err := res.RowsAffected()
	return int(rows), err
}

func (r *SQLiteRepository) DeleteAllUnpinned() (int, error) {
	query := "DELETE FROM clips WHERE is_pinned = 0"
	res, err := r.db.Exec(query)
	if err != nil {
		return 0, err
	}
	rows, err := res.RowsAffected()
	return int(rows), err
}

func (r *SQLiteRepository) DeleteUnpinnedInDateRange(start, end time.Time) (int, error) {
	query := "DELETE FROM clips WHERE is_pinned = 0 AND created_at >= ? AND created_at < ?"
	res, err := r.db.Exec(query, start, end)
	if err != nil {
		return 0, err
	}
	rows, err := res.RowsAffected()
	return int(rows), err
}
