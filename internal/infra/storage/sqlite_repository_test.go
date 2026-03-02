package storage

import (
	"clip-block/internal/core/domain"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB cria um banco de dados temporário para testes e retorna o repositório e o caminho do arquivo
func setupTestDB(t *testing.T) (*SQLiteRepository, string) {
	// Cria um arquivo temporário para o banco de dados
	tmpFile, err := os.CreateTemp("", "clip-block-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	dbPath := tmpFile.Name()
	tmpFile.Close()

	// Abre a conexão com o banco de dados
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		os.Remove(dbPath)
		t.Fatalf("Failed to open database: %v", err)
	}

	repo := &SQLiteRepository{db: db}

	// Executa a migração para criar as tabelas
	if err := repo.migrate(); err != nil {
		db.Close()
		os.Remove(dbPath)
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return repo, dbPath
}

func TestSQLiteRepository_SaveAndGet(t *testing.T) {
	repo, dbPath := setupTestDB(t)
	defer os.Remove(dbPath)
	defer repo.db.Close()

	clip := domain.Clip{
		ID:        "1",
		Content:   "test content",
		IsPinned:  false,
		CreatedAt: time.Now().UTC(),
	}

	if err := repo.Save(clip); err != nil {
		t.Errorf("Save() error = %v", err)
	}

	// Test GetAll
	clips, total, err := repo.GetAll(1, 10, nil)
	if err != nil {
		t.Errorf("GetAll() error = %v", err)
	}

	if total != 1 {
		t.Errorf("Expected total 1, got %d", total)
	}
	if len(clips) != 1 {
		t.Errorf("Expected 1 clip, got %d", len(clips))
	}
	if len(clips) > 0 {
		if clips[0].ID != clip.ID {
			t.Errorf("Expected ID %s, got %s", clip.ID, clips[0].ID)
		}
		if clips[0].Content != clip.Content {
			t.Errorf("Expected Content %s, got %s", clip.Content, clips[0].Content)
		}
	}
}

func TestSQLiteRepository_TogglePin(t *testing.T) {
	repo, dbPath := setupTestDB(t)
	defer os.Remove(dbPath)
	defer repo.db.Close()

	clip := domain.Clip{
		ID:        "1",
		Content:   "test",
		IsPinned:  false,
		CreatedAt: time.Now(),
	}
	repo.Save(clip)

	// Pin
	if err := repo.TogglePin("1"); err != nil {
		t.Errorf("TogglePin() error = %v", err)
	}

	clips, _, _ := repo.GetAll(1, 10, nil)
	if len(clips) > 0 && !clips[0].IsPinned {
		t.Error("Expected clip to be pinned")
	}

	// Unpin
	if err := repo.TogglePin("1"); err != nil {
		t.Errorf("TogglePin() error = %v", err)
	}

	clips, _, _ = repo.GetAll(1, 10, nil)
	if len(clips) > 0 && clips[0].IsPinned {
		t.Error("Expected clip to be unpinned")
	}
}

func TestSQLiteRepository_Delete(t *testing.T) {
	repo, dbPath := setupTestDB(t)
	defer os.Remove(dbPath)
	defer repo.db.Close()

	repo.Save(domain.Clip{ID: "1", Content: "c1", CreatedAt: time.Now()})

	if err := repo.Delete("1"); err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	_, total, _ := repo.GetAll(1, 10, nil)
	if total != 0 {
		t.Errorf("Expected 0 clips, got %d", total)
	}
}

func TestSQLiteRepository_DeleteOlderThan(t *testing.T) {
	repo, dbPath := setupTestDB(t)
	defer os.Remove(dbPath)
	defer repo.db.Close()

	now := time.Now()
	old := now.Add(-24 * time.Hour)

	// 1. Old unpinned (should be deleted)
	repo.Save(domain.Clip{ID: "1", Content: "old", CreatedAt: old, IsPinned: false})
	// 2. New unpinned (should stay)
	repo.Save(domain.Clip{ID: "2", Content: "new", CreatedAt: now, IsPinned: false})
	// 3. Old pinned (should stay)
	repo.Save(domain.Clip{ID: "3", Content: "pinned old", CreatedAt: old, IsPinned: true})

	count, err := repo.DeleteOlderThan(now.Add(-1 * time.Hour))
	if err != nil {
		t.Errorf("DeleteOlderThan() error = %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 deleted clip, got %d", count)
	}

	clips, _, _ := repo.GetAll(1, 10, nil)
	if len(clips) != 2 {
		t.Errorf("Expected 2 clips remaining, got %d", len(clips))
	}
}

func TestSQLiteRepository_DeleteAllUnpinned(t *testing.T) {
	repo, dbPath := setupTestDB(t)
	defer os.Remove(dbPath)
	defer repo.db.Close()

	repo.Save(domain.Clip{ID: "1", Content: "unpinned", IsPinned: false})
	repo.Save(domain.Clip{ID: "2", Content: "pinned", IsPinned: true})

	count, err := repo.DeleteAllUnpinned()
	if err != nil {
		t.Errorf("DeleteAllUnpinned() error = %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 deleted clip, got %d", count)
	}

	clips, _, _ := repo.GetAll(1, 10, nil)
	if len(clips) != 1 {
		t.Errorf("Expected 1 clip remaining, got %d", len(clips))
	}
	if len(clips) > 0 && !clips[0].IsPinned {
		t.Error("Remaining clip should be pinned")
	}
}
