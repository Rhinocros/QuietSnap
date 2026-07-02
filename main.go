package main

import (
	"embed"
	"log"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"autoscreen/pkg/autostart"
	"autoscreen/pkg/config"
	"autoscreen/pkg/scheduler"
)

//go:embed frontend/dist
var assets embed.FS

//go:embed build/resources/appicon.png
var appIcon []byte

type AppService struct {
	app         *application.App
	selectorWin *application.WebviewWindow
}

func (a *AppService) GetConfig() *config.Config {
	return config.LoadConfig()
}

func (a *AppService) SaveConfig(cfg *config.Config) error {
	if cfg.AutoStart {
		autostart.Enable()
	} else {
		autostart.Disable()
	}
	return config.SaveConfig(cfg)
}

func (a *AppService) StartScheduler() error {
	cfg := config.LoadConfig()
	return scheduler.Start(cfg)
}

func (a *AppService) StopScheduler() {
	scheduler.Stop()
}

func (a *AppService) IsSchedulerRunning() bool {
	return scheduler.IsRunning()
}

func (a *AppService) SelectDirectory() (string, error) {
	return a.app.Dialog.OpenFile().
		CanChooseDirectories(true).
		CanChooseFiles(false).
		SetTitle("选择保存目录").
		PromptForSingleSelection()
}

func (a *AppService) MinimizeWindow() {
	windows := a.app.Window.GetAll()
	if len(windows) > 0 {
		windows[0].Hide()
	}
}

func (a *AppService) SetWindowTitle(title string) {
	windows := a.app.Window.GetAll()
	if len(windows) > 0 {
		windows[0].SetTitle(title)
	}
}

func (a *AppService) OpenRegionSelector() {
	if a.selectorWin != nil {
		a.selectorWin.Show()
		a.selectorWin.Focus()
		return
	}

	a.selectorWin = a.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Select Region",
		Width:            800,
		Height:           600,
		Frameless:        true,
		AlwaysOnTop:      true,
		BackgroundType:   application.BackgroundTypeTransparent,
		BackgroundColour: application.NewRGBA(0, 0, 0, 0),
		URL:              "/?mode=selector",
	})

	a.selectorWin.Fullscreen()
	a.selectorWin.Show()
	a.selectorWin.Focus()
}

type Region struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

func (a *AppService) SubmitRegion(r Region) {
	a.app.Event.Emit("region-selected", r)
	if a.selectorWin != nil {
		a.selectorWin.Close()
		a.selectorWin = nil
	}
}

func main() {
	appService := &AppService{}

	app := application.New(application.Options{
		Name:        "QuietSnap",
		Description: "定时后台自动截图软件",
		Icon:        appIcon,
		Services: []application.Service{
			application.NewService(appService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})
	appService.app = app

	tray := app.SystemTray.New()
	tray.SetTooltip("QuietSnap")
	tray.SetIcon(appIcon)

	myMenu := app.NewMenu()
	myMenu.Add("显示界面").OnClick(func(ctx *application.Context) {
		windows := app.Window.GetAll()
		if len(windows) > 0 {
			windows[0].Show()
			windows[0].Focus()
		} else {
			// If window was destroyed, recreate it
			createWindow(app)
		}
	})
	myMenu.AddSeparator()
	myMenu.Add("退出程序").OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	tray.SetMenu(myMenu)

	// Determine if we should start hidden (e.g. from autostart)
	startHidden := false
	for _, arg := range os.Args {
		if arg == "--autostart" {
			startHidden = true
		}
	}

	cfg := config.LoadConfig()
	if startHidden && !cfg.ShowUIOnAutoStart {
		// Just run in tray, don't show window
	} else {
		createWindow(app)
	}

	// Auto resume scheduler if it was running? Usually we want it to auto-start if configured, but let's just start it.
	if cfg.AutoStart {
		scheduler.Start(cfg)
	}

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func createWindow(app *application.App) *application.WebviewWindow {
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "QuietSnap 自动截屏控制面板",
		Width:  800,
		Height: 760,
		Hidden: false,
	})
	
	window.RegisterHook(events.Common.WindowClosing, func(evt *application.WindowEvent) {
		evt.Cancel() // Cancel the default close behavior
		app.Event.Emit("show-close-dialog")
	})

	return window
}

func (a *AppService) HideWindow() {
	windows := a.app.Window.GetAll()
	if len(windows) > 0 {
		windows[0].Hide()
	}
}

func (a *AppService) QuitApp() {
	a.app.Quit()
}
