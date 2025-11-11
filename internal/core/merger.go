package core

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
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
	templateArticles map[string]bool // Уникальные артикулы из листа "Шаблон" для Ozon пресета
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

	// Инициализируем карту для артикулов
	m.templateArticles = make(map[string]bool)

	// Вычисляем общее количество операций для прогресса
	// +1 для базового файла
	totalFiles := 1 + len(filePaths)
	totalOperations := len(sheetConfigs) * totalFiles
	currentOperation := 0

	// Сначала обрабатываем лист "Шаблон", если он есть (для Ozon пресета)
	templateConfig, hasTemplate := sheetConfigs["Шаблон"]
	if hasTemplate && templateConfig.Enabled {
		m.logger.Info("обработка листа", "sheet", "Шаблон")

		rowsMerged, warnings, err := m.mergeSheetWithWriter(writer, "Шаблон", templateConfig, baseFilePath, filePaths, &currentOperation, totalOperations)
		if err != nil {
			writer.Close()
			return nil, fmt.Errorf("ошибка при обработке листа '%s': %w", "Шаблон", err)
		}

		result.SheetStats["Шаблон"] = &SheetStat{
			RowsMerged: rowsMerged,
			FilesCount: totalFiles,
		}
		result.TotalRows += rowsMerged
		result.Warnings = append(result.Warnings, warnings...)
		result.ProcessedSheets++

		m.logger.Info("лист 'Шаблон' обработан, извлечено артикулов", "count", len(m.templateArticles))
	}

	// Обрабатываем остальные листы
	for sheetName, sheetConfig := range sheetConfigs {
		// Пропускаем уже обработанный лист "Шаблон"
		if sheetName == "Шаблон" {
			continue
		}

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

		// Применяем фильтрацию по значению столбца, если настроена
		if config.FilterColumn >= 0 && len(config.FilterValues) > 0 {
			beforeFilter := len(dataRows)
			
			// DEBUG: Собираем уникальные значения в столбце для логирования
			uniqueValues := make(map[string]int)
			for _, row := range dataRows {
				if config.FilterColumn < len(row) {
					val := row[config.FilterColumn]
					uniqueValues[val]++
				}
			}
			
			dataRows = filterRowsByColumnValue(dataRows, config.FilterColumn, config.FilterValues)
			afterFilter := len(dataRows)
			excludedCount := beforeFilter - afterFilter
			
			m.logger.Info("применена фильтрация по столбцу",
				"file", filepath.Base(filePath),
				"sheet", sheetName,
				"before_filter", beforeFilter,
				"after_filter", afterFilter,
				"excluded_count", excludedCount,
				"kept_values", config.FilterValues,
				"column_index", config.FilterColumn,
				"unique_brands_before_filter", uniqueValues,
			)
		}

		// Для листа "Шаблон" извлекаем артикулы после фильтрации (для Ozon пресета)
		if sheetName == "Шаблон" && len(dataRows) > 0 {
			// Получаем заголовки
			var headerRow []string
			if config.HeaderRow > 0 && len(baseRows) >= config.HeaderRow {
				headerRow = baseRows[config.HeaderRow-1]
			}
			
			// Извлекаем артикулы из обработанных строк
			articles := extractArticlesFromRows(headerRow, dataRows)
			
			// Добавляем артикулы в общую карту
			for article := range articles {
				m.templateArticles[article] = true
			}
			
			m.logger.Info("извлечены артикулы из листа Шаблон",
				"file", filepath.Base(filePath),
				"articles_count", len(articles),
				"total_articles", len(m.templateArticles),
			)
		}

		// Применяем фильтрацию по артикулам из листа "Шаблон", если настроена
		if config.UseTemplateArticles && len(m.templateArticles) > 0 && len(dataRows) > 0 {
			beforeFilter := len(dataRows)
			
			// Получаем заголовки
			var headerRow []string
			if config.HeaderRow > 0 && len(baseRows) >= config.HeaderRow {
				headerRow = baseRows[config.HeaderRow-1]
			}
			
			dataRows = filterRowsByArticles(headerRow, dataRows, m.templateArticles)
			afterFilter := len(dataRows)
			excludedCount := beforeFilter - afterFilter
			
			m.logger.Info("применена фильтрация по артикулам из листа Шаблон",
				"file", filepath.Base(filePath),
				"sheet", sheetName,
				"before_filter", beforeFilter,
				"after_filter", afterFilter,
				"excluded_count", excludedCount,
				"template_articles_count", len(m.templateArticles),
			)
		}

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

// filterRowsByColumnValue фильтрует строки, оставляя только те, где значение в указанном столбце совпадает с одним из заданных значений
func filterRowsByColumnValue(rows [][]string, columnIndex int, filterValues []string) [][]string {
	if columnIndex < 0 || len(filterValues) == 0 {
		return rows
	}

	// Нормализуем значения для фильтрации: trim + lowercase
	normalizedFilterValues := make([]string, len(filterValues))
	for i, val := range filterValues {
		normalizedFilterValues[i] = strings.ToLower(strings.TrimSpace(val))
	}

	filtered := make([][]string, 0, len(rows))

	for _, row := range rows {
		// Проверяем, что столбец существует в строке
		if columnIndex >= len(row) {
			// Если столбца нет, исключаем строку
			continue
		}

		// Нормализуем значение ячейки: trim + lowercase
		cellValue := strings.ToLower(strings.TrimSpace(row[columnIndex]))
		shouldKeep := false

		// Проверяем, совпадает ли значение ячейки с одним из нужных значений
		for _, filterValue := range normalizedFilterValues {
			if cellValue == filterValue {
				shouldKeep = true
				break
			}
		}

		// Оставляем только строки с нужными значениями
		if shouldKeep {
			filtered = append(filtered, row)
		}
	}

	return filtered
}

// extractArticlesFromRows извлекает уникальные артикулы из строк данных
// headerRow - строка заголовков (обычно строка 2)
// dataRows - строки данных
// Возвращает map с уникальными артикулами для быстрого поиска
func extractArticlesFromRows(headerRow []string, dataRows [][]string) map[string]bool {
	articles := make(map[string]bool)

	// Ищем столбец "Артикул*" в заголовках
	articleColIndex := -1
	for i, header := range headerRow {
		if strings.Contains(strings.ToLower(header), "артикул") {
			articleColIndex = i
			break
		}
	}

	// Если столбец не найден, возвращаем пустой map
	if articleColIndex == -1 {
		return articles
	}

	// Извлекаем уникальные артикулы
	for _, row := range dataRows {
		if articleColIndex < len(row) {
			article := strings.TrimSpace(row[articleColIndex])
			if article != "" {
				articles[article] = true
			}
		}
	}

	return articles
}

// filterRowsByArticles фильтрует строки по списку артикулов
// headerRow - строка заголовков
// dataRows - строки данных для фильтрации
// articles - map с разрешенными артикулами
// Возвращает только строки, артикулы которых есть в articles
func filterRowsByArticles(headerRow []string, dataRows [][]string, articles map[string]bool) [][]string {
	if len(articles) == 0 {
		// Если артикулов нет, возвращаем пустой массив
		return [][]string{}
	}

	// Ищем столбец "Артикул*" в заголовках
	articleColIndex := -1
	for i, header := range headerRow {
		if strings.Contains(strings.ToLower(header), "артикул") {
			articleColIndex = i
			break
		}
	}

	// Если столбец не найден, возвращаем пустой массив
	if articleColIndex == -1 {
		return [][]string{}
	}

	// Фильтруем строки
	filtered := make([][]string, 0, len(dataRows))
	for _, row := range dataRows {
		if articleColIndex < len(row) {
			article := strings.TrimSpace(row[articleColIndex])
			if articles[article] {
				filtered = append(filtered, row)
			}
		}
	}

	return filtered
}

