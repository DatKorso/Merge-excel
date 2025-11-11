package core

import (
	"fmt"
	"log/slog"

	"github.com/DatKorso/Merge-excel/internal/excel"
)

// BaseAnalyzer анализирует базовый файл и создает конфигурацию для объединения
type BaseAnalyzer struct {
	reader *excel.Reader
	logger *slog.Logger
}

// NewBaseAnalyzer создает новый анализатор базового файла
func NewBaseAnalyzer(reader *excel.Reader, logger *slog.Logger) *BaseAnalyzer {
	if logger == nil {
		logger = slog.Default()
	}

	return &BaseAnalyzer{
		reader: reader,
		logger: logger,
	}
}

// GetSheetNames возвращает список всех листов в базовом файле
func (a *BaseAnalyzer) GetSheetNames(filePath string) ([]string, error) {
	reader, err := excel.NewReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer reader.Close()

	sheetNames := reader.GetSheetNames()
	if len(sheetNames) == 0 {
		return nil, fmt.Errorf("файл не содержит листов")
	}

	return sheetNames, nil
}

// GetHeaders возвращает заголовки для указанного листа
func (a *BaseAnalyzer) GetHeaders(filePath, sheetName string, headerRow int) ([]string, error) {
	reader, err := excel.NewReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer reader.Close()

	if !reader.SheetExists(sheetName) {
		return nil, fmt.Errorf("лист '%s' не найден", sheetName)
	}

	headers, err := reader.GetHeaderRow(sheetName, headerRow)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать заголовки: %w", err)
	}

	return headers, nil
}
