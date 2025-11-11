package core

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"

	"github.com/korso/merge-excel/internal/excel"
	apperrors "github.com/korso/merge-excel/internal/errors"
)

// ProgressCallback функция обратного вызова для обновления прогресса
type ProgressCallback func(current, total int, message string)

// Merger выполняет объединение данных из нескольких Excel файлов
type Merger struct {
	profile          *Profile
	progressCallback ProgressCallback
	logger           *slog.Logger
	mu               sync.Mutex
}

// NewMerger создает новый объединитель файлов
func NewMerger(profile *Profile, logger *slog.Logger) *Merger {
	if logger == nil {
		logger = slog.Default()
	}

	return &Merger{
		profile: profile,
		logger:  logger,
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
	SheetData       map[string][][]string // Данные по листам: sheet_name -> rows
	TotalFiles      int                   // Общее количество обработанных файлов
	TotalRows       int                   // Общее количество объединенных строк
	ProcessedSheets int                   // Количество обработанных листов
	Warnings        []string              // Предупреждения при обработке
}

// MergeFiles объединяет несколько Excel файлов согласно профилю
func (m *Merger) MergeFiles(filePaths []string) (*MergeResult, error) {
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("список файлов для объединения пуст")
	}

	// Валидируем профиль
	if err := m.profile.Validate(); err != nil {
		return nil, fmt.Errorf("профиль невалиден: %w", err)
	}

	m.logger.Info("начало объединения файлов",
		"files_count", len(filePaths),
		"profile", m.profile.ProfileName,
	)

	result := &MergeResult{
		SheetData: make(map[string][][]string),
		Warnings:  []string{},
	}

	// Получаем список включенных листов
	enabledSheets := m.profile.GetEnabledSheets()
	if len(enabledSheets) == 0 {
		return nil, fmt.Errorf("нет включенных листов для обработки")
	}

	// Вычисляем общее количество операций для прогресса
	totalOperations := len(enabledSheets) * len(filePaths)
	currentOperation := 0

	// Обрабатываем каждый лист
	for _, sheetConfig := range enabledSheets {
		m.logger.Info("обработка листа", "sheet", sheetConfig.SheetName)

		sheetData, warnings, err := m.mergeSheet(sheetConfig, filePaths, &currentOperation, totalOperations)
		if err != nil {
			return nil, fmt.Errorf("ошибка при обработке листа '%s': %w", sheetConfig.SheetName, err)
		}

		result.SheetData[sheetConfig.SheetName] = sheetData
		result.Warnings = append(result.Warnings, warnings...)
		result.ProcessedSheets++
	}

	// Подсчитываем общую статистику
	result.TotalFiles = len(filePaths)
	for _, rows := range result.SheetData {
		result.TotalRows += len(rows)
	}

	m.logger.Info("объединение завершено",
		"total_files", result.TotalFiles,
		"total_rows", result.TotalRows,
		"processed_sheets", result.ProcessedSheets,
		"warnings_count", len(result.Warnings),
	)

	return result, nil
}

// mergeSheet объединяет один лист из всех файлов
func (m *Merger) mergeSheet(
	config SheetConfig,
	filePaths []string,
	currentOp *int,
	totalOps int,
) ([][]string, []string, error) {
	var allRows [][]string
	var warnings []string
	var headers []string
	headersSet := false

	for i, filePath := range filePaths {
		*currentOp++
		m.notifyProgress(*currentOp, totalOps,
			fmt.Sprintf("Обработка %s, лист %s (%d/%d)",
				filepath.Base(filePath), config.SheetName, i+1, len(filePaths)))

		// Открываем файл
		reader, err := excel.NewReader(filePath)
		if err != nil {
			warning := fmt.Sprintf("не удалось открыть файл %s: %v", filepath.Base(filePath), err)
			warnings = append(warnings, warning)
			m.logger.Warn(warning, "file", filePath, "error", err)
			continue
		}

		// Проверяем наличие листа
		if !reader.SheetExists(config.SheetName) {
			warning := fmt.Sprintf("лист '%s' не найден в файле %s", config.SheetName, filepath.Base(filePath))
			warnings = append(warnings, warning)
			m.logger.Warn(warning, "file", filePath, "sheet", config.SheetName)
			reader.Close()
			continue
		}

		// Валидируем структуру (только если заголовки уже установлены)
		if headersSet && len(config.Headers) > 0 {
			if err := reader.ValidateStructure(config.SheetName, config.HeaderRow, config.Headers); err != nil {
				warning := fmt.Sprintf("структура файла %s не совпадает с базовым: %v",
					filepath.Base(filePath), err)
				warnings = append(warnings, warning)
				m.logger.Warn(warning, "file", filePath, "error", err)
				reader.Close()
				continue
			}
		}

		// Получаем заголовки из первого файла
		if !headersSet {
			h, err := reader.GetHeaderRow(config.SheetName, config.HeaderRow)
			if err != nil {
				warning := fmt.Sprintf("не удалось прочитать заголовки из %s: %v",
					filepath.Base(filePath), err)
				warnings = append(warnings, warning)
				m.logger.Warn(warning, "file", filePath, "error", err)
				reader.Close()
				continue
			}
			headers = h
			headersSet = true

			// Добавляем заголовки как первую строку результата
			allRows = append(allRows, headers)
		}

		// Получаем строки данных (без заголовков)
		dataRows, err := reader.GetDataRows(config.SheetName, config.HeaderRow)
		if err != nil {
			warning := fmt.Sprintf("не удалось прочитать данные из %s: %v",
				filepath.Base(filePath), err)
			warnings = append(warnings, warning)
			m.logger.Warn(warning, "file", filePath, "error", err)
			reader.Close()
			continue
		}

		// Фильтруем пустые строки если включена настройка
		if m.profile.Settings.SkipEmptyRows {
			dataRows = filterEmptyRows(dataRows)
		}

		// Добавляем данные
		allRows = append(allRows, dataRows...)

		m.logger.Info("файл обработан",
			"file", filepath.Base(filePath),
			"sheet", config.SheetName,
			"rows_added", len(dataRows),
		)

		reader.Close()
	}

	// Проверяем, что получили хотя бы заголовки
	if len(allRows) == 0 {
		return nil, warnings, &apperrors.AppError{
			Code:    apperrors.ErrCodeMergeError,
			Message: fmt.Sprintf("не удалось получить данные для листа '%s'", config.SheetName),
		}
	}

	return allRows, warnings, nil
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

// ValidateFiles проверяет все файлы перед объединением
func (m *Merger) ValidateFiles(filePaths []string) error {
	if len(filePaths) == 0 {
		return fmt.Errorf("список файлов пуст")
	}

	for i, filePath := range filePaths {
		m.logger.Info("валидация файла", "file", filepath.Base(filePath), "index", i+1)

		reader, err := excel.NewReader(filePath)
		if err != nil {
			return fmt.Errorf("файл %s недоступен: %w", filepath.Base(filePath), err)
		}

		if err := reader.ValidateFile(); err != nil {
			reader.Close()
			return fmt.Errorf("файл %s невалиден: %w", filepath.Base(filePath), err)
		}

		reader.Close()
	}

	m.logger.Info("все файлы валидны", "count", len(filePaths))
	return nil
}

// GetStats возвращает статистику по профилю
func (m *Merger) GetStats() map[string]interface{} {
	enabledSheets := m.profile.GetEnabledSheets()

	return map[string]interface{}{
		"profile_name":         m.profile.ProfileName,
		"base_file":            m.profile.BaseFileName,
		"total_sheets":         len(m.profile.Sheets),
		"enabled_sheets":       len(enabledSheets),
		"skip_empty_rows":      m.profile.Settings.SkipEmptyRows,
		"show_warnings":        m.profile.Settings.ShowWarnings,
		"enabled_sheet_names":  getSheetNames(enabledSheets),
	}
}

// getSheetNames извлекает имена листов из конфигураций
func getSheetNames(configs []SheetConfig) []string {
	names := make([]string, len(configs))
	for i, config := range configs {
		names[i] = config.SheetName
	}
	return names
}
