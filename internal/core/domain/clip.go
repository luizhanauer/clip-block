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
	GetAll() ([]Clip, error)
	Delete(id string) error
	TogglePin(id string) error
}
