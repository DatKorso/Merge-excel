package core

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/DatKorso/Merge-excel/internal/excel"
)

func TestNewMerger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	merger := NewMerger(nil, logger)

	if merger == nil {
		t.Fatal("merger не должен быть nil")
	}

	if merger.logger != logger {
		t.Error("логгер не установлен корректно")
	}
}

func TestMergeFiles(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	testFile1 := filepath.Join("..", "..", "testdata", "Повседневная обувь_04.11.2025.xlsx")
	testFile2 := filepath.Join("..", "..", "testdata", "Повседневная обувь_04.11.2025 (1).xlsx")

	// Проверяем существование тестовых файлов
	if _, err := os.Stat(testFile1); os.IsNotExist(err) {
		t.Skip("тестовый файл не найден:", testFile1)
	}
	if _, err := os.Stat(testFile2); os.IsNotExist(err) {
		t.Skip("тестовый файл не найден:", testFile2)
	}

	// Создаем анализатор для получения конфигурации
	reader, err := excel.NewReader(testFile1)
	if err != nil {
		t.Fatalf("не удалось открыть тестовый файл: %v", err)
	}
	defer reader.Close()

	analyzer := NewBaseAnalyzer(reader, logger)

	// Получаем список листов
	sheetNames, err := analyzer.GetSheetNames(testFile1)
	if err != nil {
		t.Fatalf("не удалось получить список листов: %v", err)
	}

	if len(sheetNames) == 0 {
		t.Fatal("нет листов для тестирования")
	}

	// Создаем конфигурацию для первого листа
	sheetConfigs := make(map[string]*SheetConfig)
	sheetName := sheetNames[0]

	config := &SheetConfig{
		SheetName: sheetName,
		Enabled:   true,
		HeaderRow: 4, // Строка 4 для тестовых файлов
		Headers:   []string{},
	}

	// Получаем заголовки
	headers, err := analyzer.GetHeaders(testFile1, sheetName, 4)
	if err != nil {
		t.Fatalf("не удалось получить заголовки: %v", err)
	}

	config.Headers = headers
	sheetConfigs[sheetName] = config

	// Создаем merger
	merger := NewMerger(nil, logger)

	// Тестируем прогресс callback
	var progressUpdates int
	merger.SetProgressCallback(func(current, total int, message string) {
		progressUpdates++
		t.Logf("Прогресс: %d/%d - %s", current, total, message)
	})

	// Выполняем объединение (базовый файл + дополнительные файлы)
	files := []string{testFile2}
	result, err := merger.MergeFiles(testFile1, files, sheetConfigs)
	if err != nil {
		t.Fatalf("ошибка при объединении файлов: %v", err)
	}

	// Проверяем результат
	if result == nil {
		t.Fatal("результат не должен быть nil")
	}

	if result.ProcessedFiles != 2 {
		t.Errorf("ожидалось 2 файла, получено %d", result.ProcessedFiles)
	}

	if result.ProcessedSheets == 0 {
		t.Error("ожидался хотя бы один обработанный лист")
	}

	if len(result.SheetStats) == 0 {
		t.Error("ожидались данные листов")
	}

	if progressUpdates == 0 {
		t.Error("ожидались обновления прогресса")
	}

	// Выводим статистику
	t.Logf("Всего файлов: %d", result.ProcessedFiles)
	t.Logf("Обработано листов: %d", result.ProcessedSheets)
	t.Logf("Всего строк: %d", result.TotalRows)
	t.Logf("Предупреждений: %d", len(result.Warnings))

	for sheetName, stats := range result.SheetStats {
		t.Logf("Лист '%s': %d строк", sheetName, stats.RowsMerged)
	}

	for _, warning := range result.Warnings {
		t.Logf("Предупреждение: %s", warning)
	}

	// Очистка
	if result.WorkbookData != nil {
		result.WorkbookData.Close()
	}
}

func TestMergeFilesWithErrors(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("пустой базовый файл", func(t *testing.T) {
		merger := NewMerger(nil, logger)
		sheetConfigs := map[string]*SheetConfig{
			"Sheet1": {
				SheetName: "Sheet1",
				Enabled:   true,
				HeaderRow: 1,
			},
		}

		_, err := merger.MergeFiles("", []string{}, sheetConfigs)
		if err == nil {
			t.Error("ожидалась ошибка для пустого базового файла")
		}
	})

	t.Run("нет листов для обработки", func(t *testing.T) {
		merger := NewMerger(nil, logger)

		_, err := merger.MergeFiles("test.xlsx", []string{"file1.xlsx"}, map[string]*SheetConfig{})
		if err == nil {
			t.Error("ожидалась ошибка когда нет листов для обработки")
		}
	})
}

func TestFilterEmptyRows(t *testing.T) {
	tests := []struct {
		name     string
		input    [][]string
		expected int
	}{
		{
			name: "нет пустых строк",
			input: [][]string{
				{"A", "B", "C"},
				{"1", "2", "3"},
			},
			expected: 2,
		},
		{
			name: "одна пустая строка",
			input: [][]string{
				{"A", "B", "C"},
				{"", "", ""},
				{"1", "2", "3"},
			},
			expected: 2,
		},
		{
			name: "частично пустые строки",
			input: [][]string{
				{"A", "", "C"},
				{"", "", ""},
				{"1", "2", ""},
			},
			expected: 2,
		},
		{
			name:     "все строки пустые",
			input:    [][]string{{"", ""}, {"", ""}},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterEmptyRows(tt.input)
			if len(result) != tt.expected {
				t.Errorf("ожидалось %d строк, получено %d", tt.expected, len(result))
			}
		})
	}
}

func TestFilterRowsByColumnValue(t *testing.T) {
	tests := []struct {
		name         string
		input        [][]string
		columnIndex  int
		filterValues []string
		expected     int
	}{
		{
			name: "оставляем только Shuzzi",
			input: [][]string{
				{"A", "Shuzzi", "C"},
				{"B", "Other", "D"},
				{"E", "Shuzzi", "F"},
			},
			columnIndex:  1,
			filterValues: []string{"Shuzzi"},
			expected:     2, // Только строки с "Shuzzi"
		},
		{
			name: "фильтрация с разным регистром",
			input: [][]string{
				{"A", "shuzzi", "C"},
				{"B", "SHUZZI", "D"},
				{"E", "Other", "F"},
			},
			columnIndex:  1,
			filterValues: []string{"Shuzzi"},
			expected:     2, // Только строки с "shuzzi" и "SHUZZI"
		},
		{
			name: "фильтрация с пробелами",
			input: [][]string{
				{"A", " Shuzzi ", "C"},
				{"B", "Shuzzi", "D"},
				{"E", "Other", "F"},
			},
			columnIndex:  1,
			filterValues: []string{"Shuzzi"},
			expected:     2, // Только строки с "Shuzzi"
		},
		{
			name: "оставляем несколько значений",
			input: [][]string{
				{"A", "Value1", "C"},
				{"B", "Value2", "D"},
				{"E", "Value3", "F"},
			},
			columnIndex:  1,
			filterValues: []string{"Value1", "Value3"},
			expected:     2, // Строки с "Value1" и "Value3"
		},
		{
			name: "нет совпадений",
			input: [][]string{
				{"A", "Keep1", "C"},
				{"B", "Keep2", "D"},
			},
			columnIndex:  1,
			filterValues: []string{"NotFound"},
			expected:     0, // Ни одна строка не подходит
		},
		{
			name: "отрицательный индекс колонки",
			input: [][]string{
				{"A", "B", "C"},
				{"D", "E", "F"},
			},
			columnIndex:  -1,
			filterValues: []string{"B"},
			expected:     2, // Фильтрация не применяется
		},
		{
			name: "пустой список значений для фильтрации",
			input: [][]string{
				{"A", "B", "C"},
				{"D", "E", "F"},
			},
			columnIndex:  1,
			filterValues: []string{},
			expected:     2, // Фильтрация не применяется
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterRowsByColumnValue(tt.input, tt.columnIndex, tt.filterValues)
			if len(result) != tt.expected {
				t.Errorf("ожидалось %d строк, получено %d", tt.expected, len(result))
			}
		})
	}
}

func TestExtractArticlesFromRows(t *testing.T) {
	tests := []struct {
		name       string
		headerRow  []string
		dataRows   [][]string
		expected   map[string]bool
	}{
		{
			name:      "извлечение артикулов",
			headerRow: []string{"Название", "Артикул*", "Цена"},
			dataRows: [][]string{
				{"Товар 1", "ART-001", "1000"},
				{"Товар 2", "ART-002", "2000"},
				{"Товар 3", "ART-001", "1500"}, // Дубликат
			},
			expected: map[string]bool{
				"ART-001": true,
				"ART-002": true,
			},
		},
		{
			name:      "артикул с пробелами",
			headerRow: []string{"Название", "Артикул*", "Цена"},
			dataRows: [][]string{
				{"Товар 1", " ART-001 ", "1000"},
				{"Товар 2", "ART-002", "2000"},
			},
			expected: map[string]bool{
				"ART-001": true,
				"ART-002": true,
			},
		},
		{
			name:      "столбец артикул не найден",
			headerRow: []string{"Название", "Код", "Цена"},
			dataRows: [][]string{
				{"Товар 1", "CODE-001", "1000"},
			},
			expected: map[string]bool{},
		},
		{
			name:      "пустые артикулы игнорируются",
			headerRow: []string{"Название", "Артикул*", "Цена"},
			dataRows: [][]string{
				{"Товар 1", "ART-001", "1000"},
				{"Товар 2", "", "2000"},
				{"Товар 3", "  ", "1500"},
			},
			expected: map[string]bool{
				"ART-001": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractArticlesFromRows(tt.headerRow, tt.dataRows)
			
			if len(result) != len(tt.expected) {
				t.Errorf("ожидалось %d артикулов, получено %d", len(tt.expected), len(result))
			}
			
			for article := range tt.expected {
				if !result[article] {
					t.Errorf("артикул '%s' не найден в результате", article)
				}
			}
		})
	}
}

func TestFilterRowsByArticles(t *testing.T) {
	tests := []struct {
		name      string
		headerRow []string
		dataRows  [][]string
		articles  map[string]bool
		expected  int
	}{
		{
			name:      "фильтрация по артикулам",
			headerRow: []string{"Название", "Артикул*", "Цена"},
			dataRows: [][]string{
				{"Товар 1", "ART-001", "1000"},
				{"Товар 2", "ART-002", "2000"},
				{"Товар 3", "ART-003", "1500"},
			},
			articles: map[string]bool{
				"ART-001": true,
				"ART-003": true,
			},
			expected: 2,
		},
		{
			name:      "фильтрация с пробелами в артикулах",
			headerRow: []string{"Название", "Артикул*", "Цена"},
			dataRows: [][]string{
				{"Товар 1", " ART-001 ", "1000"},
				{"Товар 2", "ART-002", "2000"},
			},
			articles: map[string]bool{
				"ART-001": true,
			},
			expected: 1,
		},
		{
			name:      "пустой список артикулов",
			headerRow: []string{"Название", "Артикул*", "Цена"},
			dataRows: [][]string{
				{"Товар 1", "ART-001", "1000"},
				{"Товар 2", "ART-002", "2000"},
			},
			articles: map[string]bool{},
			expected: 0,
		},
		{
			name:      "столбец артикул не найден",
			headerRow: []string{"Название", "Код", "Цена"},
			dataRows: [][]string{
				{"Товар 1", "CODE-001", "1000"},
			},
			articles: map[string]bool{
				"CODE-001": true,
			},
			expected: 0,
		},
		{
			name:      "нет совпадений",
			headerRow: []string{"Название", "Артикул*", "Цена"},
			dataRows: [][]string{
				{"Товар 1", "ART-001", "1000"},
				{"Товар 2", "ART-002", "2000"},
			},
			articles: map[string]bool{
				"ART-999": true,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterRowsByArticles(tt.headerRow, tt.dataRows, tt.articles)
			if len(result) != tt.expected {
				t.Errorf("ожидалось %d строк, получено %d", tt.expected, len(result))
			}
		})
	}
}
