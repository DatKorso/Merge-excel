package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// Config конфигурация логгера
type Config struct {
	Level      slog.Level
	LogFile    string
	MaxSize    int64 // максимальный размер файла в байтах
	MaxBackups int   // максимальное количество старых лог-файлов
	Console    bool  // выводить ли в консоль
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	return &Config{
		Level:      slog.LevelInfo,
		LogFile:    filepath.Join(homeDir, ".excel-merger", "logs", "excel-merger.log"),
		MaxSize:    10 * 1024 * 1024, // 10 MB
		MaxBackups: 5,
		Console:    true,
	}
}

// InitLogger инициализирует логгер
func InitLogger(cfg *Config) (*slog.Logger, error) {
	// Создаем директорию для логов
	logDir := filepath.Dir(cfg.LogFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Проверяем размер файла и ротируем при необходимости
	if info, err := os.Stat(cfg.LogFile); err == nil && info.Size() > cfg.MaxSize {
		if err := rotateLogFile(cfg); err != nil {
			return nil, fmt.Errorf("failed to rotate log file: %w", err)
		}
	}

	// Открываем файл для записи
	file, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Настраиваем вывод
	var writer io.Writer = file
	if cfg.Console {
		writer = io.MultiWriter(file, os.Stdout)
	}

	// Создаем хендлер
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: true,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger, nil
}

// rotateLogFile выполняет ротацию лог-файлов
func rotateLogFile(cfg *Config) error {
	// Удаляем самый старый файл, если достигнут лимит
	if cfg.MaxBackups > 0 {
		oldestBackup := fmt.Sprintf("%s.%d", cfg.LogFile, cfg.MaxBackups)
		os.Remove(oldestBackup) // Игнорируем ошибку, если файл не существует
	}

	// Сдвигаем все файлы
	for i := cfg.MaxBackups - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", cfg.LogFile, i)
		newPath := fmt.Sprintf("%s.%d", cfg.LogFile, i+1)
		os.Rename(oldPath, newPath) // Игнорируем ошибку, если файл не существует
	}

	// Переименовываем текущий файл
	backupPath := fmt.Sprintf("%s.1", cfg.LogFile)
	return os.Rename(cfg.LogFile, backupPath)
}
