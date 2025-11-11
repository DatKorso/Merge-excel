package core

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"github.com/DatKorso/Merge-excel/internal/excel"
)

// ProgressCallback функция обратного вызова для обновления прогресса
type ProgressCallback func(current, total int, message string)

// ProgressUpdate информация об обновлении прогресса
type ProgressUpdate struct {
	Current int    // Текущий шаг
	Total   int    // Всего шагов
	Message string // Сообщение о текущей операции
}

// Merger выполняет объединение данных из нескольких Excel файлов
type Merger struct {
	reader           *excel.Reader
	progressCallback ProgressCallback
	logger           *slog.Logger
	mu               sync.Mutex
}

// NewMerger создает новый объединитель файлов
func NewMerger(reader *excel.Reader, logger *slog.Logger) *Merger {
	if logger == nil {
		logger = slog.Default()
	}

	return &Merger{
		reader: reader,
		logger: logger,
	}
}

// SetProgressCallback устанавливает функцию обратного вызова для прогресса
func (m *Merger) SetProgressCallback(callback ProgressCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.progressCallback = callback
}

// notifyProgress уведомляет о прогрессе выполнения
func (m *Merger) notifyProgress(current, total int, message string) {
	m.mu.Lock()
	callback := m.progressCallback
	m.mu.Unlock()

	if callback != nil {
		callback(current, total, message)
	}
}

// MergeResult результат объединения файлов
type MergeResult struct {
	WorkbookData    interface{}            // Данные workbook для сохранения
	ProcessedFiles  int                    // Общее количество обработанных файлов
	ProcessedSheets int                    // Количество обработанных листов
	TotalRows       int                    // Общее количество объединенных строк
	SheetStats      map[string]*SheetStat  // Статистика по листам
	Duration        time.Duration          // Время выполнения
	Warnings        []string               // Предупреждения при обработке
}

// SheetStat статистика по листу
type SheetStat struct {
	RowsMerged int
	FilesCount int
}

// MergeFiles объединяет несколько Excel файлов согласно конфигурации
func (m *Merger) MergeFiles(filePaths []string, sheetConfigs map[string]*SheetConfig) (*MergeResult, error) {
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("список файлов для объединения пуст")
	}

	if len(sheetConfigs) == 0 {
		return nil, fmt.Errorf("нет листов для обработки")
	}

	m.logger.Info("начало объединения файлов",
		"files_count", len(filePaths),
		"sheets_count", len(sheetConfigs),
	)

	result := &MergeResult{
		SheetStats: make(map[string]*SheetStat),
		Warnings:   []string{},
	}

	// Вычисляем общее количество операций для прогресса
	totalOperations := len(sheetConfigs) * len(filePaths)
	currentOperation := 0

	// Создаем новую книгу для результата (используем первый файл как шаблон)
	// Пока оставим WorkbookData как placeholder - его заполнит writer
	// В реальной реализации здесь будет excelize.File

	// Обрабатываем каждый лист
	for sheetName, sheetConfig := range sheetConfigs {
		if !sheetConfig.Enabled {
			continue
		}

		m.logger.Info("обработка листа", "sheet", sheetName)

		rowsMerged, warnings, err := m.mergeSheet(sheetName, sheetConfig, filePaths, &currentOperation, totalOperations)
		if err != nil {
			return nil, fmt.Errorf("ошибка при обработке листа '%s': %w", sheetName, err)
		}

		result.SheetStats[sheetName] = &SheetStat{
			RowsMerged: rowsMerged,
			FilesCount: len(filePaths),
		}
		result.TotalRows += rowsMerged
		result.Warnings = append(result.Warnings, warnings...)
		result.ProcessedSheets++
	}

	result.ProcessedFiles = len(filePaths)

	m.logger.Info("объединение завершено",
		"processed_files", result.ProcessedFiles,
		"total_rows", result.TotalRows,
		"processed_sheets", result.ProcessedSheets,
		"warnings_count", len(result.Warnings),
	)

	return result, nil
}

// mergeSheet объединяет один лист из всех файлов
func (m *Merger) mergeSheet(
	sheetName string,
	config *SheetConfig,
	filePaths []string,
	currentOp *int,
	totalOps int,
) (int, []string, error) {
	var warnings []string
	rowsMerged := 0

	for i, filePath := range filePaths {
		*currentOp++
		m.notifyProgress(*currentOp, totalOps,
			fmt.Sprintf("Обработка %s, лист %s (%d/%d)",
				filepath.Base(filePath), sheetName, i+1, len(filePaths)))

		// Открываем файл
		reader, err := excel.NewReader(filePath)
		if err != nil {
			warning := fmt.Sprintf("не удалось открыть файл %s: %v", filepath.Base(filePath), err)
			warnings = append(warnings, warning)
			m.logger.Warn(warning, "file", filePath, "error", err)
			continue
		}

		// Проверяем наличие листа
		if !reader.SheetExists(sheetName) {
			warning := fmt.Sprintf("лист '%s' не найден в файле %s", sheetName, filepath.Base(filePath))
			warnings = append(warnings, warning)
			m.logger.Warn(warning, "file", filePath, "sheet", sheetName)
			reader.Close()
			continue
		}

		// Получаем строки данных (без заголовков)
		dataRows, err := reader.GetDataRows(sheetName, config.HeaderRow)
		if err != nil {
			warning := fmt.Sprintf("не удалось прочитать данные из %s: %v",
				filepath.Base(filePath), err)
			warnings = append(warnings, warning)
			m.logger.Warn(warning, "file", filePath, "error", err)
			reader.Close()
			continue
		}

		// Фильтруем пустые строки
		dataRows = filterEmptyRows(dataRows)
		rowsMerged += len(dataRows)

		m.logger.Info("файл обработан",
			"file", filepath.Base(filePath),
			"sheet", sheetName,
			"rows_added", len(dataRows),
		)

		reader.Close()
	}

	return rowsMerged, warnings, nil
}

// filterEmptyRows фильтрует полностью пустые строки
func filterEmptyRows(rows [][]string) [][]string {
	filtered := make([][]string, 0, len(rows))

	for _, row := range rows {
		isEmpty := true
		for _, cell := range row {
			if cell != "" {
				isEmpty = false
				break
			}
		}

		if !isEmpty {
			filtered = append(filtered, row)
		}
	}

	return filtered
}
