package storage

import (
	"clip-block/internal/core/domain"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type JSONRepository struct {
	mu       sync.Mutex
	filePath string
}

func NewJSONRepository(filename string) *JSONRepository {
	// Garante que o diretório existe (ex: ~/.config/clipblock)
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "clipblock")
	_ = os.MkdirAll(path, 0755)

	return &JSONRepository{
		filePath: filepath.Join(path, filename),
	}
}

func (r *JSONRepository) load() ([]domain.Clip, error) {
	file, err := os.ReadFile(r.filePath)
	if os.IsNotExist(err) {
		return []domain.Clip{}, nil
	}
	if err != nil {
		return nil, err
	}

	var clips []domain.Clip
	if err := json.Unmarshal(file, &clips); err != nil {
		return []domain.Clip{}, nil
	}
	return clips, nil
}

func (r *JSONRepository) saveFile(clips []domain.Clip) error {
	data, err := json.MarshalIndent(clips, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.filePath, data, 0644)
}

func (r *JSONRepository) GetAll() ([]domain.Clip, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	clips, err := r.load()
	if err != nil {
		return nil, err
	}

	// Ordena: Pinned primeiro (opcional), depois por data decrescente
	sort.Slice(clips, func(i, j int) bool {
		return clips[i].CreatedAt.After(clips[j].CreatedAt)
	})

	return clips, nil
}

func (r *JSONRepository) Save(clip domain.Clip) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	clips, _ := r.load()
	// Adiciona no início
	clips = append([]domain.Clip{clip}, clips...)
	return r.saveFile(clips)
}

func (r *JSONRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	clips, _ := r.load()
	newClips := []domain.Clip{}
	for _, c := range clips {
		if c.ID != id {
			newClips = append(newClips, c)
		}
	}
	return r.saveFile(newClips)
}

func (r *JSONRepository) TogglePin(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	clips, _ := r.load()
	for i := range clips {
		if clips[i].ID == id {
			clips[i].IsPinned = !clips[i].IsPinned
			break
		}
	}
	return r.saveFile(clips)
}
