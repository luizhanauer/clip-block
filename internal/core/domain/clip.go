package domain

import (
	"time"

	"github.com/google/uuid"
)

// Clip representa um item copiado
type Clip struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	IsPinned  bool      `json:"is_pinned"`
}

// NewClip cria uma nova instância com ID e Timestamp automáticos
func NewClip(content string) Clip {
	return Clip{
		ID:        uuid.New().String(),
		Content:   content,
		CreatedAt: time.Now(),
		IsPinned:  false,
	}
}

// ClipRepository define o contrato para persistência
type ClipRepository interface {
	Save(clip Clip) error
	// GetAll retorna uma fatia paginada de clips e o total de itens.
	// O filtro 'pinned' pode ser nil (todos), true (fixados) ou false (não fixados).
	GetAll(page, pageSize int, pinned *bool) (clips []Clip, total int, err error)
	Delete(id string) error
	TogglePin(id string) error
	// DeleteOlderThan apaga clips (não fixados) criados antes da data especificada.
	// Retorna o número de clips apagados.
	DeleteOlderThan(before time.Time) (int, error)
	// DeleteAllUnpinned apaga todos os clips que não estão fixados.
	DeleteAllUnpinned() (int, error)
	// DeleteUnpinnedInDateRange apaga clips não fixados dentro de um intervalo de datas.
	DeleteUnpinnedInDateRange(start, end time.Time) (int, error)
}
