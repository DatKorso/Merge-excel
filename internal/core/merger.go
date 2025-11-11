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
	WorkbookData    *excel.Writer          // Объединенная книга Excel для сохранения
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
// baseFilePath - путь к базовому файлу (его данные тоже будут включены)
// filePaths - список дополнительных файлов для объединения
func (m *Merger) MergeFiles(baseFilePath string, filePaths []string, sheetConfigs map[string]*SheetConfig) (*MergeResult, error) {
	if baseFilePath == "" {
		return nil, fmt.Errorf("путь к базовому файлу не указан")
	}

	if len(sheetConfigs) == 0 {
		return nil, fmt.Errorf("нет листов для обработки")
	}

	m.logger.Info("начало объединения файлов",
		"base_file", baseFilePath,
		"additional_files_count", len(filePaths),
		"sheets_count", len(sheetConfigs),
	)

	result := &MergeResult{
		SheetStats: make(map[string]*SheetStat),
		Warnings:   []string{},
	}

	// Создаем новый Writer для результата
	writer := excel.NewWriter()
	result.WorkbookData = writer

	// Вычисляем общее количество операций для прогресса
	// +1 для базового файла
	totalFiles := 1 + len(filePaths)
	totalOperations := len(sheetConfigs) * totalFiles
	currentOperation := 0

	// Обрабатываем каждый лист
	for sheetName, sheetConfig := range sheetConfigs {
		if !sheetConfig.Enabled {
			continue
		}

		m.logger.Info("обработка листа", "sheet", sheetName)

		rowsMerged, warnings, err := m.mergeSheetWithWriter(writer, sheetName, sheetConfig, baseFilePath, filePaths, &currentOperation, totalOperations)
		if err != nil {
			writer.Close()
			return nil, fmt.Errorf("ошибка при обработке листа '%s': %w", sheetName, err)
		}

		result.SheetStats[sheetName] = &SheetStat{
			RowsMerged: rowsMerged,
			FilesCount: totalFiles,
		}
		result.TotalRows += rowsMerged
		result.Warnings = append(result.Warnings, warnings...)
		result.ProcessedSheets++
	}

	result.ProcessedFiles = totalFiles

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

// mergeSheetWithWriter объединяет один лист из всех файлов и записывает в Writer
func (m *Merger) mergeSheetWithWriter(
	writer *excel.Writer,
	sheetName string,
	config *SheetConfig,
	baseFilePath string,
	filePaths []string,
	currentOp *int,
	totalOps int,
) (int, []string, error) {
	var warnings []string
	rowsMerged := 0

	// Создаем лист в результирующей книге
	if err := writer.CreateSheet(sheetName); err != nil {
		return 0, warnings, fmt.Errorf("не удалось создать лист '%s': %w", sheetName, err)
	}

	// Открываем базовый файл для копирования заголовков и строк до них
	baseReader, err := excel.NewReader(baseFilePath)
	if err != nil {
		return 0, warnings, fmt.Errorf("не удалось открыть базовый файл: %w", err)
	}
	defer baseReader.Close()

	// Проверяем наличие листа в базовом файле
	if !baseReader.SheetExists(sheetName) {
		return 0, warnings, fmt.Errorf("лист '%s' не найден в базовом файле", sheetName)
	}

	// Получаем все строки из базового файла
	baseRows, err := baseReader.GetRows(sheetName)
	if err != nil {
		return 0, warnings, fmt.Errorf("не удалось прочитать базовый файл: %w", err)
	}

	// Копируем строки до заголовков включительно (от 1 до headerRow)
	if config.HeaderRow > 0 && len(baseRows) >= config.HeaderRow {
		headerRows := baseRows[:config.HeaderRow]
		if err := writer.WriteRows(sheetName, 1, headerRows); err != nil {
			return 0, warnings, fmt.Errorf("не удалось записать заголовки: %w", err)
		}
	}

	// Начальная строка для данных (следующая после заголовков)
	currentRow := config.HeaderRow + 1

	// Объединяем все файлы (включая базовый)
	allFiles := append([]string{baseFilePath}, filePaths...)

	// Обрабатываем каждый файл
	for i, filePath := range allFiles {
		*currentOp++
		m.notifyProgress(*currentOp, totalOps,
			fmt.Sprintf("Обработка %s, лист %s (%d/%d)",
				filepath.Base(filePath), sheetName, i+1, len(allFiles)))

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

		// Записываем данные в результирующий файл
		if len(dataRows) > 0 {
			if err := writer.WriteRows(sheetName, currentRow, dataRows); err != nil {
				reader.Close()
				return 0, warnings, fmt.Errorf("не удалось записать данные: %w", err)
			}
			currentRow += len(dataRows)
			rowsMerged += len(dataRows)
		}

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
