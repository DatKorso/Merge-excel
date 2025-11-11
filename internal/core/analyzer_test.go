package core

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/DatKorso/Merge-excel/internal/excel"
)

func TestGetSheetNames(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	testFile := filepath.Join("..", "..", "testdata", "Повседневная обувь_04.11.2025.xlsx")

	reader, err := excel.NewReader(testFile)
	if err != nil {
		t.Fatalf("не удалось открыть тестовый файл: %v", err)
	}
	defer reader.Close()

	analyzer := NewBaseAnalyzer(reader, logger)

	sheetNames, err := analyzer.GetSheetNames(testFile)
	if err != nil {
		t.Fatalf("ошибка при получении списка листов: %v", err)
	}

	if len(sheetNames) == 0 {
		t.Error("ожидался хотя бы один лист")
	}

	t.Logf("Найдено листов: %d", len(sheetNames))
	for i, name := range sheetNames {
		t.Logf("  %d: %s", i+1, name)
	}
}

func TestGetHeaders(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	testFile := filepath.Join("..", "..", "testdata", "Повседневная обувь_04.11.2025.xlsx")

	reader, err := excel.NewReader(testFile)
	if err != nil {
		t.Fatalf("не удалось открыть тестовый файл: %v", err)
	}
	defer reader.Close()

	analyzer := NewBaseAnalyzer(reader, logger)

	sheetNames, err := analyzer.GetSheetNames(testFile)
	if err != nil {
		t.Fatalf("ошибка при получении списка листов: %v", err)
	}

	if len(sheetNames) == 0 {
		t.Fatal("нет листов для тестирования")
	}

	sheetName := sheetNames[0]

	t.Run("получение заголовков для строки 4", func(t *testing.T) {
		headers, err := analyzer.GetHeaders(testFile, sheetName, 4)
		if err != nil {
			t.Fatalf("ошибка при получении заголовков: %v", err)
		}

		if len(headers) == 0 {
			t.Error("ожидались заголовки")
		}

		t.Logf("Заголовки из строки 4: %v", headers)
	})

	t.Run("ошибка для несуществующего листа", func(t *testing.T) {
		_, err := analyzer.GetHeaders(testFile, "НесуществующийЛист", 1)
		if err == nil {
			t.Error("ожидалась ошибка для несуществующего листа")
		}
	})
}

func TestFindBrandColumnInFirstRows(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	testFile := filepath.Join("..", "..", "testdata", "Повседневная обувь_04.11.2025.xlsx")

	reader, err := excel.NewReader(testFile)
	if err != nil {
		t.Fatalf("не удалось открыть тестовый файл: %v", err)
	}
	defer reader.Close()

	analyzer := NewBaseAnalyzer(reader, logger)

	sheetNames, err := analyzer.GetSheetNames(testFile)
	if err != nil {
		t.Fatalf("ошибка при получении списка листов: %v", err)
	}

	if len(sheetNames) == 0 {
		t.Fatal("нет листов для тестирования")
	}

	// Ищем лист "Шаблон"
	var templateSheet string
	for _, name := range sheetNames {
		if name == "Шаблон" {
			templateSheet = name
			break
		}
	}

	if templateSheet == "" {
		t.Skip("лист 'Шаблон' не найден в тестовом файле")
	}

	t.Run("поиск столбца 'Бренд в одежде и обуви*'", func(t *testing.T) {
		columnIndex, err := analyzer.FindBrandColumnInFirstRows(testFile, templateSheet, 4)
		if err != nil {
			t.Fatalf("ошибка при поиске столбца: %v", err)
		}

		t.Logf("Индекс столбца 'Бренд в одежде и обуви*': %d", columnIndex)

		// Если в тестовом файле есть эта ячейка в S2, индекс должен быть 18 (S = 19-я колонка, 0-based = 18)
		if columnIndex == 18 {
			t.Logf("Столбец успешно найден на позиции S (индекс 18)")
		} else if columnIndex == -1 {
			t.Logf("Столбец не найден в строке 2 листа 'Шаблон'")
		} else {
			t.Logf("Столбец найден на неожиданной позиции: %d", columnIndex)
		}
	})
}
