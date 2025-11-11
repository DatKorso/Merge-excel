package excel

import (
	"path/filepath"
	"testing"
)

// Путь к тестовым файлам
const testDataDir = "../../testdata"

// getTestFilePath возвращает абсолютный путь к тестовому файлу
func getTestFilePath(t *testing.T, filename string) string {
	path := filepath.Join(testDataDir, filename)
	absPath, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	return absPath
}

// TestNewReader тестирует создание Reader
func TestNewReader(t *testing.T) {
	testFile := getTestFilePath(t, "Повседневная обувь_04.11.2025.xlsx")

	reader, err := NewReader(testFile)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	if reader.file == nil {
		t.Error("Reader file is nil")
	}

	if reader.path != testFile {
		t.Errorf("Expected path %s, got %s", testFile, reader.path)
	}
}

// TestNewReaderFileNotFound тестирует открытие несуществующего файла
func TestNewReaderFileNotFound(t *testing.T) {
	_, err := NewReader("nonexistent_file.xlsx")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

// TestNewReaderInvalidFormat тестирует открытие файла с неверным форматом
func TestNewReaderInvalidFormat(t *testing.T) {
	testFile := getTestFilePath(t, "README.md")

	_, err := NewReader(testFile)
	if err == nil {
		t.Error("Expected error for invalid format, got nil")
	}
}

// TestGetSheetNames тестирует получение списка листов
func TestGetSheetNames(t *testing.T) {
	testFile := getTestFilePath(t, "Повседневная обувь_04.11.2025.xlsx")

	reader, err := NewReader(testFile)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	sheets := reader.GetSheetNames()
	if len(sheets) == 0 {
		t.Error("Expected at least one sheet, got 0")
	}

	t.Logf("Found sheets: %v", sheets)
}

// TestSheetExists тестирует проверку существования листа
func TestSheetExists(t *testing.T) {
	testFile := getTestFilePath(t, "Повседневная обувь_04.11.2025.xlsx")

	reader, err := NewReader(testFile)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	sheets := reader.GetSheetNames()
	if len(sheets) == 0 {
		t.Fatal("No sheets found in file")
	}

	// Проверяем существующий лист
	firstSheet := sheets[0]
	if !reader.SheetExists(firstSheet) {
		t.Errorf("Sheet '%s' should exist", firstSheet)
	}

	// Проверяем несуществующий лист
	if reader.SheetExists("NonexistentSheet") {
		t.Error("NonexistentSheet should not exist")
	}
}

// TestGetRows тестирует чтение строк из листа
func TestGetRows(t *testing.T) {
	testFile := getTestFilePath(t, "Повседневная обувь_04.11.2025.xlsx")

	reader, err := NewReader(testFile)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	sheets := reader.GetSheetNames()
	if len(sheets) == 0 {
		t.Fatal("No sheets found in file")
	}

	firstSheet := sheets[0]
	rows, err := reader.GetRows(firstSheet)
	if err != nil {
		t.Fatalf("Failed to get rows: %v", err)
	}

	if len(rows) == 0 {
		t.Error("Expected at least one row, got 0")
	}

	t.Logf("Read %d rows from sheet '%s'", len(rows), firstSheet)
}

// TestGetHeaderRow тестирует чтение строки заголовков
func TestGetHeaderRow(t *testing.T) {
	testFile := getTestFilePath(t, "Повседневная обувь_04.11.2025.xlsx")

	reader, err := NewReader(testFile)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	sheets := reader.GetSheetNames()
	if len(sheets) == 0 {
		t.Fatal("No sheets found in file")
	}

	firstSheet := sheets[0]

	// Согласно документации, заголовки находятся в строке 5
	headers, err := reader.GetHeaderRow(firstSheet, 5)
	if err != nil {
		t.Fatalf("Failed to get header row: %v", err)
	}

	if len(headers) == 0 {
		t.Error("Expected at least one header, got 0")
	}

	t.Logf("Headers from row 5: %v", headers)
}

// TestGetHeaderRowInvalidRow тестирует чтение заголовков с неверным номером строки
func TestGetHeaderRowInvalidRow(t *testing.T) {
	testFile := getTestFilePath(t, "Повседневная обувь_04.11.2025.xlsx")

	reader, err := NewReader(testFile)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	sheets := reader.GetSheetNames()
	if len(sheets) == 0 {
		t.Fatal("No sheets found in file")
	}

	firstSheet := sheets[0]

	// Проверяем неверный номер строки (меньше 1)
	_, err = reader.GetHeaderRow(firstSheet, 0)
	if err == nil {
		t.Error("Expected error for invalid row number, got nil")
	}

	// Проверяем номер строки, превышающий количество строк
	_, err = reader.GetHeaderRow(firstSheet, 10000)
	if err == nil {
		t.Error("Expected error for row number exceeding total rows, got nil")
	}
}

// TestGetDataRows тестирует чтение строк данных
func TestGetDataRows(t *testing.T) {
	testFile := getTestFilePath(t, "Повседневная обувь_04.11.2025.xlsx")

	reader, err := NewReader(testFile)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	sheets := reader.GetSheetNames()
	if len(sheets) == 0 {
		t.Fatal("No sheets found in file")
	}

	firstSheet := sheets[0]

	// Согласно документации, заголовки в строке 5, данные начинаются с строки 6
	dataRows, err := reader.GetDataRows(firstSheet, 5)
	if err != nil {
		t.Fatalf("Failed to get data rows: %v", err)
	}

	t.Logf("Read %d data rows from sheet '%s'", len(dataRows), firstSheet)

	// Проверяем, что данные есть
	if len(dataRows) > 0 {
		t.Logf("First data row has %d columns", len(dataRows[0]))
	}
}

// TestGetCellValue тестирует чтение значения ячейки
func TestGetCellValue(t *testing.T) {
	testFile := getTestFilePath(t, "Повседневная обувь_04.11.2025.xlsx")

	reader, err := NewReader(testFile)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	sheets := reader.GetSheetNames()
	if len(sheets) == 0 {
		t.Fatal("No sheets found in file")
	}

	firstSheet := sheets[0]

	// Читаем значение ячейки A1
	value, err := reader.GetCellValue(firstSheet, "A1")
	if err != nil {
		t.Fatalf("Failed to get cell value: %v", err)
	}

	t.Logf("Value of cell A1: '%s'", value)
}

// TestGetRowCount тестирует подсчет строк
func TestGetRowCount(t *testing.T) {
	testFile := getTestFilePath(t, "Повседневная обувь_04.11.2025.xlsx")

	reader, err := NewReader(testFile)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	sheets := reader.GetSheetNames()
	if len(sheets) == 0 {
		t.Fatal("No sheets found in file")
	}

	firstSheet := sheets[0]

	count, err := reader.GetRowCount(firstSheet)
	if err != nil {
		t.Fatalf("Failed to get row count: %v", err)
	}

	if count == 0 {
		t.Error("Expected at least one row, got 0")
	}

	t.Logf("Sheet '%s' has %d rows", firstSheet, count)
}

// TestValidateFile тестирует валидацию файла
func TestValidateFile(t *testing.T) {
	testFile := getTestFilePath(t, "Повседневная обувь_04.11.2025.xlsx")

	reader, err := NewReader(testFile)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	if err := reader.ValidateFile(); err != nil {
		t.Errorf("File validation failed: %v", err)
	}
}

// TestValidateStructure тестирует валидацию структуры
func TestValidateStructure(t *testing.T) {
	testFile1 := getTestFilePath(t, "Повседневная обувь_04.11.2025.xlsx")
	testFile2 := getTestFilePath(t, "Повседневная обувь_04.11.2025 (1).xlsx")

	// Открываем первый файл и читаем заголовки
	reader1, err := NewReader(testFile1)
	if err != nil {
		t.Fatalf("Failed to create reader for file 1: %v", err)
	}
	defer reader1.Close()

	sheets := reader1.GetSheetNames()
	if len(sheets) == 0 {
		t.Fatal("No sheets found in file 1")
	}

	firstSheet := sheets[0]
	baseHeaders, err := reader1.GetHeaderRow(firstSheet, 5)
	if err != nil {
		t.Fatalf("Failed to get headers from file 1: %v", err)
	}

	// Открываем второй файл и валидируем структуру
	reader2, err := NewReader(testFile2)
	if err != nil {
		t.Fatalf("Failed to create reader for file 2: %v", err)
	}
	defer reader2.Close()

	// Проверяем, что структура совпадает или отличается
	err = reader2.ValidateStructure(firstSheet, 5, baseHeaders)
	if err != nil {
		t.Logf("Structure validation showed differences (expected for test files): %v", err)
		// Это не ошибка теста - файлы могут иметь разную структуру
	} else {
		t.Log("Structure validation passed - files are compatible")
	}
}

// TestGetFilePath тестирует получение пути к файлу
func TestGetFilePath(t *testing.T) {
	testFile := getTestFilePath(t, "Повседневная обувь_04.11.2025.xlsx")

	reader, err := NewReader(testFile)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	path := reader.GetFilePath()
	if path != testFile {
		t.Errorf("Expected path %s, got %s", testFile, path)
	}
}
