package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/DatKorso/Merge-excel/internal/config"
	"github.com/DatKorso/Merge-excel/internal/gui"
	"github.com/DatKorso/Merge-excel/internal/logger"
	"github.com/DatKorso/Merge-excel/internal/updater"
)

const (
	appVersion = "0.1.0"
	appID      = "com.github.excel-merger"
	
	// GitHub repository info
	githubOwner = "DatKorso"
	githubRepo  = "Merge-excel"
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
	
	// Запускаем проверку обновлений в фоновой горутине
	go checkForUpdates(appLogger, application)
	
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

// checkForUpdates проверяет наличие обновлений в фоновом режиме
func checkForUpdates(appLogger *slog.Logger, application *gui.App) {
	// Небольшая задержка, чтобы окно успело загрузиться
	time.Sleep(2 * time.Second)
	
	appLogger.Info("Запуск проверки обновлений")
	
	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	// Создаем checker для обновлений
	updateChecker := updater.NewUpdateChecker(appVersion, githubOwner, githubRepo, appLogger)
	
	// Проверяем обновления
	releaseInfo, err := updateChecker.CheckForUpdates(ctx)
	if err != nil {
		appLogger.Warn("Не удалось проверить обновления", "error", err)
		return
	}
	
	// Если обновление доступно, показываем диалог
	if releaseInfo != nil && releaseInfo.IsNewer {
		appLogger.Info("Найдено обновление, показываю диалог",
			"new_version", releaseInfo.Version,
		)
		
		// Показываем диалог в UI потоке
		window := application.GetWindow()
		if window != nil {
			updater.ShowUpdateDialog(window, releaseInfo)
		}
	}
}
