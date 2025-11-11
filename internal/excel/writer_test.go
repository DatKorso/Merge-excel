package excel

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewWriter тестирует создание нового Writer
func TestNewWriter(t *testing.T) {
	writer := NewWriter()
	if writer == nil {
		t.Fatal("NewWriter returned nil")
	}
	defer writer.Close()

	if writer.file == nil {
		t.Error("Writer file is nil")
	}
}

// TestCreateSheet тестирует создание нового листа
func TestCreateSheet(t *testing.T) {
	writer := NewWriter()
	defer writer.Close()

	sheetName := "TestSheet"
	err := writer.CreateSheet(sheetName)
	if err != nil {
		t.Fatalf("Failed to create sheet: %v", err)
	}

	// Проверяем, что лист существует
	if !writer.SheetExists(sheetName) {
		t.Errorf("Sheet '%s' was not created", sheetName)
	}
}

// TestDeleteSheet тестирует удаление листа
func TestDeleteSheet(t *testing.T) {
	writer := NewWriter()
	defer writer.Close()

	sheetName := "SheetToDelete"
	err := writer.CreateSheet(sheetName)
	if err != nil {
		t.Fatalf("Failed to create sheet: %v", err)
	}

	// Удаляем лист
	err = writer.DeleteSheet(sheetName)
	if err != nil {
		t.Fatalf("Failed to delete sheet: %v", err)
	}

	// Проверяем, что лист удален
	if writer.SheetExists(sheetName) {
		t.Errorf("Sheet '%s' was not deleted", sheetName)
	}
}

// TestWriteHeaderRow тестирует запись строки заголовков
func TestWriteHeaderRow(t *testing.T) {
	writer := NewWriter()
	defer writer.Close()

	sheetName := "TestSheet"
	err := writer.CreateSheet(sheetName)
	if err != nil {
		t.Fatalf("Failed to create sheet: %v", err)
	}

	headers := []string{"Имя", "Возраст", "Город"}
	err = writer.WriteHeaderRow(sheetName, 1, headers)
	if err != nil {
		t.Fatalf("Failed to write header row: %v", err)
	}

	// Проверяем, что заголовки записаны
	rows, err := writer.file.GetRows(sheetName)
	if err != nil {
		t.Fatalf("Failed to get rows: %v", err)
	}

	if len(rows) == 0 {
		t.Fatal("No rows found after writing headers")
	}

	if len(rows[0]) != len(headers) {
		t.Errorf("Expected %d headers, got %d", len(headers), len(rows[0]))
	}

	for i, expected := range headers {
		if rows[0][i] != expected {
			t.Errorf("Header %d: expected '%s', got '%s'", i, expected, rows[0][i])
		}
	}
}

// TestWriteRow тестирует запись одной строки
func TestWriteRow(t *testing.T) {
	writer := NewWriter()
	defer writer.Close()

	sheetName := "TestSheet"
	err := writer.CreateSheet(sheetName)
	if err != nil {
		t.Fatalf("Failed to create sheet: %v", err)
	}

	data := []string{"Иван", "30", "Москва"}
	err = writer.WriteRow(sheetName, 1, data)
	if err != nil {
		t.Fatalf("Failed to write row: %v", err)
	}

	// Проверяем, что данные записаны
	rows, err := writer.file.GetRows(sheetName)
	if err != nil {
		t.Fatalf("Failed to get rows: %v", err)
	}

	if len(rows) == 0 {
		t.Fatal("No rows found after writing data")
	}

	if len(rows[0]) != len(data) {
		t.Errorf("Expected %d values, got %d", len(data), len(rows[0]))
	}

	for i, expected := range data {
		if rows[0][i] != expected {
			t.Errorf("Value %d: expected '%s', got '%s'", i, expected, rows[0][i])
		}
	}
}

// TestWriteRows тестирует запись множества строк
func TestWriteRows(t *testing.T) {
	writer := NewWriter()
	defer writer.Close()

	sheetName := "TestSheet"
	err := writer.CreateSheet(sheetName)
	if err != nil {
		t.Fatalf("Failed to create sheet: %v", err)
	}

	data := [][]string{
		{"Иван", "30", "Москва"},
		{"Мария", "25", "Санкт-Петербург"},
		{"Петр", "35", "Казань"},
	}

	err = writer.WriteRows(sheetName, 1, data)
	if err != nil {
		t.Fatalf("Failed to write rows: %v", err)
	}

	// Проверяем, что данные записаны
	rows, err := writer.file.GetRows(sheetName)
	if err != nil {
		t.Fatalf("Failed to get rows: %v", err)
	}

	if len(rows) != len(data) {
		t.Errorf("Expected %d rows, got %d", len(data), len(rows))
	}
}

// TestSetCellValue тестирует установку значения ячейки
func TestSetCellValue(t *testing.T) {
	writer := NewWriter()
	defer writer.Close()

	sheetName := "TestSheet"
	err := writer.CreateSheet(sheetName)
	if err != nil {
		t.Fatalf("Failed to create sheet: %v", err)
	}

	err = writer.SetCellValue(sheetName, "A1", "Test Value")
	if err != nil {
		t.Fatalf("Failed to set cell value: %v", err)
	}

	// Проверяем значение
	value, err := writer.file.GetCellValue(sheetName, "A1")
	if err != nil {
		t.Fatalf("Failed to get cell value: %v", err)
	}

	if value != "Test Value" {
		t.Errorf("Expected 'Test Value', got '%s'", value)
	}
}

// TestSave тестирует сохранение файла
func TestSave(t *testing.T) {
	writer := NewWriter()
	defer writer.Close()

	sheetName := "TestSheet"
	err := writer.CreateSheet(sheetName)
	if err != nil {
		t.Fatalf("Failed to create sheet: %v", err)
	}

	headers := []string{"Имя", "Возраст"}
	err = writer.WriteHeaderRow(sheetName, 1, headers)
	if err != nil {
		t.Fatalf("Failed to write headers: %v", err)
	}

	data := [][]string{
		{"Иван", "30"},
		{"Мария", "25"},
	}
	err = writer.WriteRows(sheetName, 2, data)
	if err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	// Создаем временный файл для сохранения
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test_output.xlsx")

	err = writer.Save(outputPath)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Проверяем, что файл создан
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output file was not created: %s", outputPath)
	}

	// Проверяем, что файл можно прочитать
	reader, err := NewReader(outputPath)
	if err != nil {
		t.Fatalf("Failed to open saved file: %v", err)
	}
	defer reader.Close()

	// Проверяем содержимое
	rows, err := reader.GetRows(sheetName)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if len(rows) != 3 { // 1 header + 2 data rows
		t.Errorf("Expected 3 rows, got %d", len(rows))
	}
}

// TestWriteSheetWithData тестирует создание листа с данными
func TestWriteSheetWithData(t *testing.T) {
	writer := NewWriter()
	defer writer.Close()

	sheetName := "TestSheet"
	headers := []string{"Колонка 1", "Колонка 2", "Колонка 3"}
	data := [][]string{
		{"A", "B", "C"},
		{"D", "E", "F"},
		{"G", "H", "I"},
	}

	err := writer.WriteSheetWithData(sheetName, 1, headers, data)
	if err != nil {
		t.Fatalf("Failed to write sheet with data: %v", err)
	}

	// Проверяем результат
	rows, err := writer.file.GetRows(sheetName)
	if err != nil {
		t.Fatalf("Failed to get rows: %v", err)
	}

	// Ожидаем 4 строки (1 header + 3 data)
	if len(rows) != 4 {
		t.Errorf("Expected 4 rows, got %d", len(rows))
	}

	// Проверяем заголовки
	for i, expected := range headers {
		if rows[0][i] != expected {
			t.Errorf("Header %d: expected '%s', got '%s'", i, expected, rows[0][i])
		}
	}

	// Проверяем данные
	for rowIdx, expectedRow := range data {
		for colIdx, expected := range expectedRow {
			actual := rows[rowIdx+1][colIdx]
			if actual != expected {
				t.Errorf("Row %d, Col %d: expected '%s', got '%s'",
					rowIdx+1, colIdx, expected, actual)
			}
		}
	}
}

// TestMergeSheetData тестирует добавление данных к существующему листу
func TestMergeSheetData(t *testing.T) {
	writer := NewWriter()
	defer writer.Close()

	sheetName := "TestSheet"

	// Создаем лист с начальными данными
	headers := []string{"Имя", "Возраст"}
	initialData := [][]string{
		{"Иван", "30"},
		{"Мария", "25"},
	}

	err := writer.WriteSheetWithData(sheetName, 1, headers, initialData)
	if err != nil {
		t.Fatalf("Failed to write initial data: %v", err)
	}

	// Добавляем новые данные
	newData := [][]string{
		{"Петр", "35"},
		{"Анна", "28"},
	}

	err = writer.MergeSheetData(sheetName, 1, newData)
	if err != nil {
		t.Fatalf("Failed to merge data: %v", err)
	}

	// Проверяем результат
	rows, err := writer.file.GetRows(sheetName)
	if err != nil {
		t.Fatalf("Failed to get rows: %v", err)
	}

	// Ожидаем 5 строк (1 header + 2 initial + 2 new)
	expectedRowCount := 1 + len(initialData) + len(newData)
	if len(rows) != expectedRowCount {
		t.Errorf("Expected %d rows, got %d", expectedRowCount, len(rows))
	}
}

// TestWriterGetSheetNames тестирует получение списка листов
func TestWriterGetSheetNames(t *testing.T) {
	writer := NewWriter()
	defer writer.Close()

	// Создаем несколько листов
	sheets := []string{"Sheet1", "Sheet2", "Sheet3"}
	for _, sheetName := range sheets {
		err := writer.CreateSheet(sheetName)
		if err != nil {
			t.Fatalf("Failed to create sheet '%s': %v", sheetName, err)
		}
	}

	// Получаем список листов
	sheetNames := writer.GetSheetNames()

	// Проверяем, что все листы присутствуют
	for _, expectedSheet := range sheets {
		found := false
		for _, actualSheet := range sheetNames {
			if actualSheet == expectedSheet {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Sheet '%s' not found in sheet names", expectedSheet)
		}
	}
}

// TestNewWriterFromFile тестирует создание Writer из существующего файла
func TestNewWriterFromFile(t *testing.T) {
	testFile := getTestFilePath(t, "Повседневная обувь_04.11.2025.xlsx")

	writer, err := NewWriterFromFile(testFile)
	if err != nil {
		t.Fatalf("Failed to create writer from file: %v", err)
	}
	defer writer.Close()

	if writer.file == nil {
		t.Error("Writer file is nil")
	}

	// Проверяем, что можем получить список листов
	sheets := writer.GetSheetNames()
	if len(sheets) == 0 {
		t.Error("Expected at least one sheet, got 0")
	}

	t.Logf("Loaded file with %d sheets", len(sheets))
}
