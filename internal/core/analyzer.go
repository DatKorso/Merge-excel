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

// FindBrandColumnInFirstRows ищет столбец "Бренд в одежде и обуви*" в строке 2
// Проверяет все столбцы до нахождения нужной ячейки
// Возвращает 0-based индекс столбца или -1 если не найден
func (a *BaseAnalyzer) FindBrandColumnInFirstRows(filePath, sheetName string, headerRow int) (int, error) {
	reader, err := excel.NewReader(filePath)
	if err != nil {
		return -1, fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer reader.Close()

	if !reader.SheetExists(sheetName) {
		return -1, fmt.Errorf("лист '%s' не найден", sheetName)
	}

	// Читаем строку 2 (для поиска названий атрибутов)
	row2, err := reader.GetHeaderRow(sheetName, 2)
	if err != nil {
		return -1, fmt.Errorf("не удалось прочитать строку 2: %w", err)
	}

	// Ищем ячейку "Бренд в одежде и обуви*" во всей строке
	for i, cell := range row2 {
		if cell == "Бренд в одежде и обуви*" {
			a.logger.Info("найден столбец бренда", "column_index", i, "column_letter", columnIndexToLetter(i), "sheet", sheetName)
			return i, nil
		}
	}

	a.logger.Warn("столбец 'Бренд в одежде и обуви*' не найден в строке 2", "sheet", sheetName)
	return -1, nil
}

// columnIndexToLetter преобразует 0-based индекс столбца в букву Excel (0 -> A, 25 -> Z, 26 -> AA и т.д.)
func columnIndexToLetter(index int) string {
	result := ""
	for index >= 0 {
		result = string(rune('A'+index%26)) + result
		index = index/26 - 1
	}
	return result
}
