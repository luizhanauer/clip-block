package app

import (
	"clip-block/internal/core/domain"
	"context"
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	"github.com/wailsapp/wails/v2/pkg/options" // <--- [CORREÇÃO 1] Import necessário
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx         context.Context
	repo        domain.ClipRepository
	lastContent string
	isVisible   bool
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

// [CORREÇÃO 2] Assinatura atualizada para aceitar 'options.SecondInstanceData'
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
func (a *App) GetClips() []domain.Clip {
	clips, _ := a.repo.GetAll()
	if len(clips) > 0 {
		a.lastContent = clips[0].Content
	}
	return clips
}

func (a *App) DeleteClip(id string) { a.repo.Delete(id) }
func (a *App) TogglePin(id string)  { a.repo.TogglePin(id) }

// Função para atualizar conteúdo (usada pelo Frontend se necessário)
func (a *App) UpdateClipContent(id string, newContent string) {
	clips, err := a.repo.GetAll()
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
