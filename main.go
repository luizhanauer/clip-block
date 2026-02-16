package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"

	"clip-block/internal/app"
	"clip-block/internal/infra/storage"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 1. Inicializa Infra
	repo := storage.NewJSONRepository("clips.json")

	// 2. Inicializa App Core
	myApp := app.NewApp(repo)

	// 3. Roda Wails
	err := wails.Run(&options.App{
		Title:  "ClipBlock",
		Width:  800, // Ajustado para ser mais 'slim' como um sidebar
		Height: 700,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		StartHidden:      true, // Mude para true se quiser iniciar minimizado no tray
		BackgroundColour: &options.RGBA{R: 15, G: 15, B: 18, A: 1},
		OnStartup:        myApp.Startup,
		Bind: []interface{}{
			myApp,
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "clip-block-uuid-v1",
			// [CORREÇÃO 3] Linha descomentada para ativar a lógica do F9
			OnSecondInstanceLaunch: myApp.OnSecondInstanceLaunch,
		},
		Linux: &linux.Options{
			WindowIsTranslucent: false,
			WebviewGpuPolicy:    linux.WebviewGpuPolicyAlways,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
