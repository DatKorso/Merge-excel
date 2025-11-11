package core

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestNewMerger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	profile := NewProfile("test_profile")
	profile.BaseFileName = "test.xlsx"

	merger := NewMerger(profile, logger)

	if merger == nil {
		t.Fatal("merger не должен быть nil")
	}

	if merger.profile != profile {
		t.Error("профиль не установлен корректно")
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
	analyzer, err := NewBaseAnalyzer(testFile1, logger)
	if err != nil {
		t.Fatalf("не удалось создать анализатор: %v", err)
	}
	defer analyzer.Close()

	// Анализируем листы
	configs, err := analyzer.AnalyzeSheets()
	if err != nil {
		t.Fatalf("ошибка при анализе листов: %v", err)
	}

	if len(configs) == 0 {
		t.Fatal("нет листов для тестирования")
	}

	// Устанавливаем правильный номер строки заголовков (строка 5 для тестовых файлов)
	// Но обрабатываем только листы с достаточным количеством строк
	validConfigs := []SheetConfig{}
	for i := range configs {
		// Пытаемся установить строку 5, если не получается - пропускаем этот лист
		configs[i].HeaderRow = 5
		if err := analyzer.UpdateHeadersForSheet(&configs[i]); err != nil {
			t.Logf("Пропускаем лист '%s': %v", configs[i].SheetName, err)
			continue
		}
		validConfigs = append(validConfigs, configs[i])
	}
	
	if len(validConfigs) == 0 {
		t.Skip("нет листов с достаточным количеством строк для тестирования")
	}

	// Создаем профиль
	profile := NewProfile("test_merge")
	profile.BaseFileName = testFile1
	profile.Sheets = validConfigs

	// Создаем merger
	merger := NewMerger(profile, logger)

	// Тестируем прогресс callback
	var progressUpdates int
	merger.SetProgressCallback(func(current, total int, message string) {
		progressUpdates++
		t.Logf("Прогресс: %d/%d - %s", current, total, message)
	})

	// Выполняем объединение
	files := []string{testFile1, testFile2}
	result, err := merger.MergeFiles(files)
	if err != nil {
		t.Fatalf("ошибка при объединении файлов: %v", err)
	}

	// Проверяем результат
	if result == nil {
		t.Fatal("результат не должен быть nil")
	}

	if result.TotalFiles != 2 {
		t.Errorf("ожидалось 2 файла, получено %d", result.TotalFiles)
	}

	if result.ProcessedSheets == 0 {
		t.Error("ожидался хотя бы один обработанный лист")
	}

	if len(result.SheetData) == 0 {
		t.Error("ожидались данные листов")
	}

	if progressUpdates == 0 {
		t.Error("ожидались обновления прогресса")
	}

	// Выводим статистику
	t.Logf("Всего файлов: %d", result.TotalFiles)
	t.Logf("Обработано листов: %d", result.ProcessedSheets)
	t.Logf("Всего строк: %d", result.TotalRows)
	t.Logf("Предупреждений: %d", len(result.Warnings))

	for sheetName, rows := range result.SheetData {
		t.Logf("Лист '%s': %d строк", sheetName, len(rows))
	}

	for _, warning := range result.Warnings {
		t.Logf("Предупреждение: %s", warning)
	}
}

func TestMergeFilesWithErrors(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("пустой список файлов", func(t *testing.T) {
		profile := NewProfile("test")
		profile.BaseFileName = "test.xlsx"
		merger := NewMerger(profile, logger)

		_, err := merger.MergeFiles([]string{})
		if err == nil {
			t.Error("ожидалась ошибка для пустого списка файлов")
		}
	})

	t.Run("нет включенных листов", func(t *testing.T) {
		profile := NewProfile("test")
		profile.BaseFileName = "test.xlsx"
		profile.Sheets = []SheetConfig{
			{
				SheetName: "Sheet1",
				Enabled:   false,
				HeaderRow: 1,
			},
		}

		merger := NewMerger(profile, logger)

		_, err := merger.MergeFiles([]string{"file1.xlsx"})
		if err == nil {
			t.Error("ожидалась ошибка когда нет включенных листов")
		}
	})
}

func TestValidateFiles(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	testFile := filepath.Join("..", "..", "testdata", "Повседневная обувь_04.11.2025.xlsx")

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("тестовый файл не найден:", testFile)
	}

	profile := NewProfile("test")
	merger := NewMerger(profile, logger)

	t.Run("валидные файлы", func(t *testing.T) {
		err := merger.ValidateFiles([]string{testFile})
		if err != nil {
			t.Errorf("не ожидалась ошибка для валидных файлов: %v", err)
		}
	})

	t.Run("несуществующий файл", func(t *testing.T) {
		err := merger.ValidateFiles([]string{"несуществующий.xlsx"})
		if err == nil {
			t.Error("ожидалась ошибка для несуществующего файла")
		}
	})

	t.Run("пустой список", func(t *testing.T) {
		err := merger.ValidateFiles([]string{})
		if err == nil {
			t.Error("ожидалась ошибка для пустого списка")
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

func TestGetStats(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	profile := NewProfile("test_profile")
	profile.BaseFileName = "base.xlsx"
	profile.Sheets = []SheetConfig{
		{SheetName: "Sheet1", Enabled: true, HeaderRow: 1},
		{SheetName: "Sheet2", Enabled: true, HeaderRow: 1},
		{SheetName: "Sheet3", Enabled: false, HeaderRow: 1},
	}

	merger := NewMerger(profile, logger)
	stats := merger.GetStats()

	if stats["profile_name"] != "test_profile" {
		t.Errorf("неверное имя профиля в статистике")
	}

	if stats["total_sheets"] != 3 {
		t.Errorf("ожидалось 3 листа, получено %v", stats["total_sheets"])
	}

	if stats["enabled_sheets"] != 2 {
		t.Errorf("ожидалось 2 включенных листа, получено %v", stats["enabled_sheets"])
	}

	t.Logf("Статистика: %+v", stats)
}
