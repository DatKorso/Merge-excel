package core

import (
	"fmt"
	"log/slog"

	"github.com/korso/merge-excel/internal/excel"
)

// BaseAnalyzer анализирует базовый файл и создает конфигурацию для объединения
type BaseAnalyzer struct {
	filePath string
	reader   *excel.Reader
	logger   *slog.Logger
}

// NewBaseAnalyzer создает новый анализатор базового файла
func NewBaseAnalyzer(filePath string, logger *slog.Logger) (*BaseAnalyzer, error) {
	reader, err := excel.NewReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть базовый файл: %w", err)
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &BaseAnalyzer{
		filePath: filePath,
		reader:   reader,
		logger:   logger,
	}, nil
}

// Close закрывает базовый файл и освобождает ресурсы
func (a *BaseAnalyzer) Close() error {
	if a.reader != nil {
		return a.reader.Close()
	}
	return nil
}

// AnalyzeSheets анализирует все листы в базовом файле и создает начальную конфигурацию
// По умолчанию: enabled=true, header_row=1 для каждого листа
func (a *BaseAnalyzer) AnalyzeSheets() ([]SheetConfig, error) {
	// Валидация базового файла
	if err := a.reader.ValidateFile(); err != nil {
		return nil, fmt.Errorf("базовый файл невалиден: %w", err)
	}

	sheetNames := a.reader.GetSheetNames()
	if len(sheetNames) == 0 {
		return nil, fmt.Errorf("базовый файл не содержит листов")
	}

	configs := make([]SheetConfig, 0, len(sheetNames))

	for _, sheetName := range sheetNames {
		// Создаем начальную конфигурацию с настройками по умолчанию
		config := SheetConfig{
			SheetName: sheetName,
			Enabled:   true,
			HeaderRow: 1, // По умолчанию первая строка
			Headers:   []string{},
		}

		// Пытаемся прочитать заголовки из первой строки
		headers, err := a.reader.GetHeaderRow(sheetName, 1)
		if err != nil {
			a.logger.Warn("не удалось прочитать заголовки из строки 1",
				"sheet", sheetName,
				"error", err,
			)
			// Продолжаем с пустыми заголовками - пользователь может настроить позже
		} else {
			config.Headers = headers
		}

		configs = append(configs, config)
		
		a.logger.Info("проанализирован лист",
			"sheet", sheetName,
			"headers_count", len(config.Headers),
		)
	}

	return configs, nil
}

// GetSheetNames возвращает список всех листов в базовом файле
func (a *BaseAnalyzer) GetSheetNames() []string {
	return a.reader.GetSheetNames()
}

// UpdateHeadersForSheet обновляет заголовки для конфигурации листа
// Используется когда пользователь изменяет номер строки заголовков
func (a *BaseAnalyzer) UpdateHeadersForSheet(config *SheetConfig) error {
	if config.HeaderRow < 1 {
		return fmt.Errorf("номер строки заголовков должен быть больше 0")
	}

	headers, err := a.reader.GetHeaderRow(config.SheetName, config.HeaderRow)
	if err != nil {
		return fmt.Errorf("не удалось прочитать заголовки из строки %d: %w", 
			config.HeaderRow, err)
	}

	config.Headers = headers
	
	a.logger.Info("обновлены заголовки для листа",
		"sheet", config.SheetName,
		"header_row", config.HeaderRow,
		"headers_count", len(headers),
	)

	return nil
}

// GetSheetPreview возвращает предпросмотр данных листа
// headerRow - номер строки заголовков (1-based)
// previewRows - количество строк данных для предпросмотра
func (a *BaseAnalyzer) GetSheetPreview(sheetName string, headerRow, previewRows int) (*SheetPreview, error) {
	if !a.reader.SheetExists(sheetName) {
		return nil, fmt.Errorf("лист '%s' не найден в базовом файле", sheetName)
	}

	// Получаем заголовки
	headers, err := a.reader.GetHeaderRow(sheetName, headerRow)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать заголовки: %w", err)
	}

	// Получаем строки данных
	dataRows, err := a.reader.GetDataRows(sheetName, headerRow)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать данные: %w", err)
	}

	// Ограничиваем количество строк для предпросмотра
	if len(dataRows) > previewRows {
		dataRows = dataRows[:previewRows]
	}

	// Получаем общее количество строк
	totalRows, err := a.reader.GetRowCount(sheetName)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить количество строк: %w", err)
	}

	preview := &SheetPreview{
		SheetName:  sheetName,
		HeaderRow:  headerRow,
		Headers:    headers,
		DataRows:   dataRows,
		TotalRows:  totalRows,
		DataRowsCount: totalRows - headerRow, // Строки данных = всего строк - строки до данных
	}

	return preview, nil
}

// ValidateSheetConfig проверяет корректность конфигурации листа
func (a *BaseAnalyzer) ValidateSheetConfig(config *SheetConfig) error {
	if config.SheetName == "" {
		return fmt.Errorf("имя листа не может быть пустым")
	}

	if !a.reader.SheetExists(config.SheetName) {
		return fmt.Errorf("лист '%s' не существует в базовом файле", config.SheetName)
	}

	if config.HeaderRow < 1 {
		return fmt.Errorf("номер строки заголовков должен быть больше 0")
	}

	// Проверяем, что строка заголовков существует
	rowCount, err := a.reader.GetRowCount(config.SheetName)
	if err != nil {
		return fmt.Errorf("не удалось получить количество строк: %w", err)
	}

	if rowCount < config.HeaderRow {
		return fmt.Errorf("лист содержит только %d строк, но указана строка заголовков %d",
			rowCount, config.HeaderRow)
	}

	return nil
}

// GetFilePath возвращает путь к базовому файлу
func (a *BaseAnalyzer) GetFilePath() string {
	return a.filePath
}

// SheetPreview структура для предпросмотра данных листа
type SheetPreview struct {
	SheetName     string     // Имя листа
	HeaderRow     int        // Номер строки заголовков
	Headers       []string   // Заголовки столбцов
	DataRows      [][]string // Строки данных для предпросмотра
	TotalRows     int        // Общее количество строк в листе
	DataRowsCount int        // Количество строк данных (без заголовков)
}
