package storage

import (
	"clip-block/internal/core/domain"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
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

func (r *JSONRepository) GetAll(page, pageSize int, pinned *bool) ([]domain.Clip, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	allClips, err := r.load()
	if err != nil {
		return nil, 0, err
	}

	// 1. Filtra se necessário
	var filteredClips []domain.Clip
	if pinned == nil {
		// Retorna todos os clips se nenhum filtro for especificado.
		filteredClips = allClips
	} else {
		for _, c := range allClips {
			if c.IsPinned == *pinned {
				filteredClips = append(filteredClips, c)
			}
		}
	}

	// 2. Ordena por data (mais recente primeiro)
	sort.Slice(filteredClips, func(i, j int) bool {
		return filteredClips[i].CreatedAt.After(filteredClips[j].CreatedAt)
	})

	totalItems := len(filteredClips)

	// 3. Pagina o resultado
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = totalItems // ou um valor padrão
	}

	start := (page - 1) * pageSize
	if start >= totalItems {
		return []domain.Clip{}, totalItems, nil // Página fora do range
	}

	end := start + pageSize
	if end > totalItems {
		end = totalItems
	}

	return filteredClips[start:end], totalItems, nil
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

func (r *JSONRepository) DeleteOlderThan(before time.Time) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	clips, err := r.load()
	if err != nil {
		return 0, err
	}

	var clipsToKeep []domain.Clip
	originalCount := len(clips)

	for _, c := range clips {
		// Mantém o clip se ele for fixado (pinned) OU se não for mais antigo que a data de corte
		if c.IsPinned || !c.CreatedAt.Before(before) {
			clipsToKeep = append(clipsToKeep, c)
		}
	}

	deletedCount := originalCount - len(clipsToKeep)
	if deletedCount > 0 {
		return deletedCount, r.saveFile(clipsToKeep)
	}

	return 0, nil
}

func (r *JSONRepository) DeleteAllUnpinned() (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	clips, err := r.load()
	if err != nil {
		return 0, err
	}

	var clipsToKeep []domain.Clip
	originalCount := len(clips)

	for _, c := range clips {
		if c.IsPinned {
			clipsToKeep = append(clipsToKeep, c)
		}
	}

	deletedCount := originalCount - len(clipsToKeep)
	if deletedCount > 0 {
		return deletedCount, r.saveFile(clipsToKeep)
	}

	return 0, nil
}

func (r *JSONRepository) DeleteUnpinnedInDateRange(start, end time.Time) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	clips, err := r.load()
	if err != nil {
		return 0, err
	}

	var clipsToKeep []domain.Clip
	originalCount := len(clips)

	for _, c := range clips {
		isWithinRange := !c.CreatedAt.Before(start) && c.CreatedAt.Before(end)
		// Mantém o clip se:
		// 1. Estiver fixado
		// 2. OU se NÃO estiver no intervalo de datas para apagar
		if c.IsPinned || !isWithinRange {
			clipsToKeep = append(clipsToKeep, c)
		}
	}

	deletedCount := originalCount - len(clipsToKeep)
	if deletedCount > 0 {
		return deletedCount, r.saveFile(clipsToKeep)
	}

	return 0, nil
}
