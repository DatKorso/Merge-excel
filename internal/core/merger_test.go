package core

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
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
	analyzer := NewBaseAnalyzer(nil, logger)
	defer analyzer.Close()

	if err := analyzer.AnalyzeFile(testFile1); err != nil {
		t.Fatalf("не удалось проанализировать файл: %v", err)
	}

	// Получаем конфигурацию листов
	configs := analyzer.GetSheetConfigs()
	if len(configs) == 0 {
		t.Fatal("нет листов для тестирования")
	}

	// Устанавливаем правильный номер строки заголовков (строка 5 для тестовых файлов)
	// Но обрабатываем только листы с достаточным количеством строк
	sheetConfigs := make(map[string]*SheetConfig)
	for i := range configs {
		// Пытаемся установить строку 5, если не получается - пропускаем этот лист
		configs[i].HeaderRow = 5
		configs[i].Enabled = true
		if err := analyzer.UpdateHeadersForSheet(&configs[i]); err != nil {
			t.Logf("Пропускаем лист '%s': %v", configs[i].SheetName, err)
			continue
		}
		sheetConfigs[configs[i].SheetName] = &configs[i]
	}
	
	if len(sheetConfigs) == 0 {
		t.Skip("нет листов с достаточным количеством строк для тестирования")
	}

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
