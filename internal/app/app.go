package app

import (
	"clip-block/internal/core/domain"
	"context"
	"fmt"
	"math"
	"time"

	"github.com/atotto/clipboard"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx         context.Context
	repo        domain.ClipRepository
	lastContent string
	isVisible   bool
}

// PaginatedClips é uma estrutura para retornar dados paginados para o frontend.
type PaginatedClips struct {
	Clips      []domain.Clip `json:"clips"`
	TotalItems int           `json:"total_items"`
	TotalPages int           `json:"total_pages"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
}

func NewApp(repo domain.ClipRepository) *App {
	return &App{
		repo:      repo,
		isVisible: true,
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// 1. Inicia o monitor de clipboard
	go a.startClipboardWatcher()

	// 2. DETECTOR DE FOCO
	// Esconde a janela se clicar fora dela
	runtime.EventsOn(ctx, "wails:window:lostfocus", func(optionalData ...interface{}) {
		fmt.Println(">> Perdeu foco: Escondendo...")
		runtime.WindowHide(a.ctx)
		a.isVisible = false
	})

	// Rastreia quando ela volta a ter foco
	runtime.EventsOn(ctx, "wails:window:focused", func(optionalData ...interface{}) {
		a.isVisible = true
	})
}

func (a *App) OnSecondInstanceLaunch(data options.SecondInstanceData) {
	fmt.Println(">> Atalho F9 detectado!")

	if a.isVisible {
		// Se já está na tela -> ESCONDE
		fmt.Println(">> Toggle: Escondendo")
		runtime.WindowHide(a.ctx)
		a.isVisible = false
	} else {
		// Se está escondido -> MOSTRA
		fmt.Println(">> Toggle: Mostrando")
		runtime.WindowUnminimise(a.ctx)
		runtime.WindowShow(a.ctx)
		runtime.WindowSetAlwaysOnTop(a.ctx, true)
		runtime.WindowSetAlwaysOnTop(a.ctx, false)
		a.isVisible = true
	}
}

func (a *App) startClipboardWatcher() {
	ticker := time.NewTicker(500 * time.Millisecond)
	for range ticker.C {
		content, err := clipboard.ReadAll()
		if err != nil || content == "" || content == a.lastContent {
			continue
		}

		a.lastContent = content
		clip := domain.NewClip(content)
		_ = a.repo.Save(clip)
		runtime.EventsEmit(a.ctx, "clip-added", clip)
	}
}

// --- Exports para Frontend ---
// GetClips retorna uma lista paginada de clips.
func (a *App) GetClips(page int, pageSize int, pinned *bool) (*PaginatedClips, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50 // Um valor padrão para o tamanho da página
	}

	clips, total, err := a.repo.GetAll(page, pageSize, pinned)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar clips: %w", err)
	}

	// Atualiza o último conteúdo apenas se estivermos na primeira página e houver clips
	if page == 1 && len(clips) > 0 && pinned == nil {
		a.lastContent = clips[0].Content
	}

	totalPages := 0
	if total > 0 && pageSize > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}

	return &PaginatedClips{
		Clips:      clips,
		TotalItems: total,
		TotalPages: totalPages,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (a *App) DeleteClip(id string) { a.repo.Delete(id) }
func (a *App) TogglePin(id string)  { a.repo.TogglePin(id) }

// Função para atualizar conteúdo (usada pelo Frontend se necessário)
func (a *App) UpdateClipContent(id string, newContent string) {
	// TODO: Melhorar esta função para não precisar buscar todos os clips.
	// Uma busca por ID no repositório seria mais eficiente.
	clips, _, err := a.repo.GetAll(1, 9999, nil)
	if err != nil {
		return
	}

	for i, clip := range clips {
		if clip.ID == id {
			clips[i].Content = newContent
			a.repo.Delete(id)     // Remove antigo
			a.repo.Save(clips[i]) // Salva editado
			break
		}
	}
}

func (a *App) WriteToClipboard(content string) {
	a.lastContent = content
	clipboard.WriteAll(content)
	// runtime.WindowHide(a.ctx) // Opcional: esconder ao copiar
}

func (a *App) AddClip(content string) {
	if content == "" {
		return
	}
	clip := domain.NewClip(content)
	if err := a.repo.Save(clip); err != nil {
		return
	}
	runtime.EventsEmit(a.ctx, "clip-added", clip)
}

// CleanClipsOlderThan apaga os clips mais antigos que o número de dias informado, exceto os fixados.
func (a *App) CleanClipsOlderThan(days int) (int, error) {
	if days <= 0 {
		return 0, fmt.Errorf("o número de dias deve ser positivo")
	}

	cutoffDate := time.Now().AddDate(0, 0, -days)
	deletedCount, err := a.repo.DeleteOlderThan(cutoffDate)
	if err != nil {
		return 0, fmt.Errorf("erro ao limpar clips antigos: %w", err)
	}

	// Emite um evento para o frontend saber que a lista foi alterada
	if deletedCount > 0 {
		runtime.EventsEmit(a.ctx, "clips-cleaned", deletedCount)
	}

	return deletedCount, nil
}

// CleanAllUnpinned apaga todos os clips que não estão fixados.
func (a *App) CleanAllUnpinned() (int, error) {
	deletedCount, err := a.repo.DeleteAllUnpinned()
	if err != nil {
		return 0, fmt.Errorf("erro ao limpar clips não fixados: %w", err)
	}

	if deletedCount > 0 {
		runtime.EventsEmit(a.ctx, "clips-cleaned", deletedCount)
	}

	return deletedCount, nil
}

// CleanTodayClips apaga todos os clips de hoje que não estão fixados.
func (a *App) CleanTodayClips() (int, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	deletedCount, err := a.repo.DeleteUnpinnedInDateRange(startOfDay, endOfDay)
	if err != nil {
		return 0, fmt.Errorf("erro ao limpar clips de hoje: %w", err)
	}

	if deletedCount > 0 {
		runtime.EventsEmit(a.ctx, "clips-cleaned", deletedCount)
	}

	return deletedCount, nil
}
