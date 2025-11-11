package core

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestNewBaseAnalyzer(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Путь к тестовому файлу
	testFile := filepath.Join("..", "..", "testdata", "Повседневная обувь_04.11.2025.xlsx")

	t.Run("успешное создание анализатора", func(t *testing.T) {
		analyzer, err := NewBaseAnalyzer(testFile, logger)
		if err != nil {
			t.Fatalf("не удалось создать анализатор: %v", err)
		}
		defer analyzer.Close()

		if analyzer.filePath != testFile {
			t.Errorf("ожидалось filePath = %s, получено %s", testFile, analyzer.filePath)
		}

		if analyzer.reader == nil {
			t.Error("reader не должен быть nil")
		}
	})

	t.Run("ошибка при несуществующем файле", func(t *testing.T) {
		_, err := NewBaseAnalyzer("несуществующий_файл.xlsx", logger)
		if err == nil {
			t.Error("ожидалась ошибка для несуществующего файла")
		}
	})
}

func TestAnalyzeSheets(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	testFile := filepath.Join("..", "..", "testdata", "Повседневная обувь_04.11.2025.xlsx")

	analyzer, err := NewBaseAnalyzer(testFile, logger)
	if err != nil {
		t.Fatalf("не удалось создать анализатор: %v", err)
	}
	defer analyzer.Close()

	configs, err := analyzer.AnalyzeSheets()
	if err != nil {
		t.Fatalf("ошибка при анализе листов: %v", err)
	}

	if len(configs) == 0 {
		t.Error("ожидался хотя бы один лист")
	}

	// Проверяем первый лист
	for _, config := range configs {
		if config.SheetName == "" {
			t.Error("имя листа не должно быть пустым")
		}

		if !config.Enabled {
			t.Error("по умолчанию лист должен быть включен")
		}

		if config.HeaderRow != 1 {
			t.Errorf("по умолчанию HeaderRow должен быть 1, получено %d", config.HeaderRow)
		}

		t.Logf("Лист: %s, Заголовков: %d", config.SheetName, len(config.Headers))
	}
}

func TestGetSheetNames(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	testFile := filepath.Join("..", "..", "testdata", "Повседневная обувь_04.11.2025.xlsx")

	analyzer, err := NewBaseAnalyzer(testFile, logger)
	if err != nil {
		t.Fatalf("не удалось создать анализатор: %v", err)
	}
	defer analyzer.Close()

	sheetNames := analyzer.GetSheetNames()
	if len(sheetNames) == 0 {
		t.Error("ожидался хотя бы один лист")
	}

	t.Logf("Найдено листов: %d", len(sheetNames))
	for i, name := range sheetNames {
		t.Logf("  %d: %s", i+1, name)
	}
}

func TestUpdateHeadersForSheet(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	testFile := filepath.Join("..", "..", "testdata", "Повседневная обувь_04.11.2025.xlsx")

	analyzer, err := NewBaseAnalyzer(testFile, logger)
	if err != nil {
		t.Fatalf("не удалось создать анализатор: %v", err)
	}
	defer analyzer.Close()

	// Получаем первый лист
	configs, err := analyzer.AnalyzeSheets()
	if err != nil {
		t.Fatalf("ошибка при анализе листов: %v", err)
	}

	if len(configs) == 0 {
		t.Fatal("нет листов для тестирования")
	}

	config := &configs[0]

	t.Run("обновление заголовков для строки 5", func(t *testing.T) {
		config.HeaderRow = 5
		err := analyzer.UpdateHeadersForSheet(config)
		if err != nil {
			t.Fatalf("ошибка при обновлении заголовков: %v", err)
		}

		if len(config.Headers) == 0 {
			t.Error("ожидались заголовки после обновления")
		}

		t.Logf("Заголовки из строки 5: %v", config.Headers)
	})

	t.Run("ошибка для недопустимой строки", func(t *testing.T) {
		config.HeaderRow = 0
		err := analyzer.UpdateHeadersForSheet(config)
		if err == nil {
			t.Error("ожидалась ошибка для HeaderRow = 0")
		}
	})
}

func TestGetSheetPreview(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	testFile := filepath.Join("..", "..", "testdata", "Повседневная обувь_04.11.2025.xlsx")

	analyzer, err := NewBaseAnalyzer(testFile, logger)
	if err != nil {
		t.Fatalf("не удалось создать анализатор: %v", err)
	}
	defer analyzer.Close()

	sheetNames := analyzer.GetSheetNames()
	if len(sheetNames) == 0 {
		t.Fatal("нет листов для тестирования")
	}

	sheetName := sheetNames[0]

	t.Run("предпросмотр с заголовками в строке 5", func(t *testing.T) {
		preview, err := analyzer.GetSheetPreview(sheetName, 5, 10)
		if err != nil {
			t.Fatalf("ошибка при получении предпросмотра: %v", err)
		}

		if preview.SheetName != sheetName {
			t.Errorf("ожидалось имя листа %s, получено %s", sheetName, preview.SheetName)
		}

		if preview.HeaderRow != 5 {
			t.Errorf("ожидалось HeaderRow = 5, получено %d", preview.HeaderRow)
		}

		if len(preview.Headers) == 0 {
			t.Error("ожидались заголовки")
		}

		t.Logf("Лист: %s", preview.SheetName)
		t.Logf("Заголовки: %v", preview.Headers)
		t.Logf("Всего строк: %d", preview.TotalRows)
		t.Logf("Строк данных: %d", preview.DataRowsCount)
		t.Logf("Строк в предпросмотре: %d", len(preview.DataRows))
	})

	t.Run("ошибка для несуществующего листа", func(t *testing.T) {
		_, err := analyzer.GetSheetPreview("НесуществующийЛист", 1, 10)
		if err == nil {
			t.Error("ожидалась ошибка для несуществующего листа")
		}
	})
}

func TestValidateSheetConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	testFile := filepath.Join("..", "..", "testdata", "Повседневная обувь_04.11.2025.xlsx")

	analyzer, err := NewBaseAnalyzer(testFile, logger)
	if err != nil {
		t.Fatalf("не удалось создать анализатор: %v", err)
	}
	defer analyzer.Close()

	sheetNames := analyzer.GetSheetNames()
	if len(sheetNames) == 0 {
		t.Fatal("нет листов для тестирования")
	}

	t.Run("валидная конфигурация", func(t *testing.T) {
		config := &SheetConfig{
			SheetName: sheetNames[0],
			Enabled:   true,
			HeaderRow: 5,
			Headers:   []string{},
		}

		err := analyzer.ValidateSheetConfig(config)
		if err != nil {
			t.Errorf("ожидалась валидная конфигурация: %v", err)
		}
	})

	t.Run("пустое имя листа", func(t *testing.T) {
		config := &SheetConfig{
			SheetName: "",
			Enabled:   true,
			HeaderRow: 1,
		}

		err := analyzer.ValidateSheetConfig(config)
		if err == nil {
			t.Error("ожидалась ошибка для пустого имени листа")
		}
	})

	t.Run("несуществующий лист", func(t *testing.T) {
		config := &SheetConfig{
			SheetName: "НесуществующийЛист",
			Enabled:   true,
			HeaderRow: 1,
		}

		err := analyzer.ValidateSheetConfig(config)
		if err == nil {
			t.Error("ожидалась ошибка для несуществующего листа")
		}
	})

	t.Run("недопустимый номер строки заголовков", func(t *testing.T) {
		config := &SheetConfig{
			SheetName: sheetNames[0],
			Enabled:   true,
			HeaderRow: 0,
		}

		err := analyzer.ValidateSheetConfig(config)
		if err == nil {
			t.Error("ожидалась ошибка для HeaderRow = 0")
		}
	})
}
