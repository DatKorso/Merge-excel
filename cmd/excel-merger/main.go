package main

import (
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"

	"github.com/korso/merge-excel/internal/logger"
)

const (
	appVersion = "0.1.0-alpha"
	appID      = "com.github.excel-merger"
)

func main() {
	// Инициализация директорий приложения
	if err := initAppDirectories(); err != nil {
		log.Fatalf("Ошибка при инициализации директорий: %v", err)
	}

	// Инициализация логгера
	logCfg := logger.DefaultConfig()
	appLogger, err := logger.InitLogger(logCfg)
	if err != nil {
		log.Fatalf("Ошибка при инициализации логгера: %v", err)
	}

	appLogger.Info("Excel Merger запущен", 
		"version", appVersion,
		"log_file", logCfg.LogFile,
	)

	// Создание Fyne приложения
	myApp := app.NewWithID(appID)
	myWindow := myApp.NewWindow("Excel Merger v" + appVersion)

	// Временное содержимое (будет заменено на GUI из internal/gui)
	content := widget.NewLabel("Excel Merger запущен!\nГлавное окно будет реализовано на следующем этапе.")

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(800, 600))
	
	// Логирование закрытия приложения
	myWindow.SetOnClosed(func() {
		appLogger.Info("Excel Merger завершен")
	})

	appLogger.Info("GUI инициализирован, отображаю окно")
	myWindow.ShowAndRun()
}

// initAppDirectories создает необходимые директории при первом запуске
func initAppDirectories() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	appDir := filepath.Join(homeDir, ".excel-merger")

	dirs := []string{
		appDir,
		filepath.Join(appDir, "configs", "profiles"),
		filepath.Join(appDir, "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}
