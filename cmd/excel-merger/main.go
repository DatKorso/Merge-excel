package main

import (
"log"
"os"
"path/filepath"

"github.com/DatKorso/Merge-excel/internal/config"
"github.com/DatKorso/Merge-excel/internal/gui"
"github.com/DatKorso/Merge-excel/internal/logger"
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

	// Инициализация config manager
	configManager, err := config.NewManager(appLogger)
	if err != nil {
		log.Fatalf("Ошибка при инициализации config manager: %v", err)
	}

	// Создание и запуск GUI приложения
	application := gui.NewApp(appLogger, configManager)
	
	appLogger.Info("GUI инициализирован, запускаю приложение")
	application.Run()
	
	appLogger.Info("Excel Merger завершен")
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
