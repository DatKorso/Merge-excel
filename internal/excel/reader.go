package excel

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xuri/excelize/v2"

	apperrors "github.com/DatKorso/Merge-excel/internal/errors"
)

// Reader предоставляет методы для чтения Excel файлов
type Reader struct {
	file *excelize.File
	path string
}

// NewReader создает новый Reader для указанного файла
func NewReader(path string) (*Reader, error) {
	// Проверяем существование файла
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, apperrors.NewFileNotFoundError(path)
	}

	// Проверяем расширение файла
	ext := filepath.Ext(path)
	if ext != ".xlsx" && ext != ".xlsm" {
		return nil, apperrors.NewInvalidFormatError(path)
	}

	// Открываем файл
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, apperrors.NewFileReadError(path, err)
	}

	return &Reader{
		file: f,
		path: path,
	}, nil
}

// Close закрывает файл и освобождает ресурсы
func (r *Reader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// GetSheetNames возвращает список всех листов в файле
func (r *Reader) GetSheetNames() []string {
	return r.file.GetSheetList()
}

// SheetExists проверяет существование листа
func (r *Reader) SheetExists(sheetName string) bool {
	for _, name := range r.GetSheetNames() {
		if name == sheetName {
			return true
		}
	}
	return false
}

// GetRows возвращает все строки указанного листа
func (r *Reader) GetRows(sheetName string) ([][]string, error) {
	if !r.SheetExists(sheetName) {
		return nil, apperrors.NewSheetNotFoundError(sheetName, r.path)
	}

	rows, err := r.file.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows from sheet '%s': %w", sheetName, err)
	}

	return rows, nil
}

// GetHeaderRow возвращает строку заголовков с указанного листа
// headerRowNum - номер строки заголовков (1-based index)
func (r *Reader) GetHeaderRow(sheetName string, headerRowNum int) ([]string, error) {
	if headerRowNum < 1 {
		return nil, apperrors.NewInvalidHeaderRowError(headerRowNum)
	}

	if !r.SheetExists(sheetName) {
		return nil, apperrors.NewSheetNotFoundError(sheetName, r.path)
	}

	rows, err := r.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	if len(rows) < headerRowNum {
		return nil, fmt.Errorf("лист '%s' содержит только %d строк, но указана строка заголовков %d",
			sheetName, len(rows), headerRowNum)
	}

	// Возвращаем строку заголовков (индекс = headerRowNum - 1)
	headers := rows[headerRowNum-1]

	// Фильтруем пустые заголовки
	var filteredHeaders []string
	for _, h := range headers {
		if h != "" {
			filteredHeaders = append(filteredHeaders, h)
		}
	}

	if len(filteredHeaders) == 0 {
		return nil, fmt.Errorf("строка заголовков %d на листе '%s' пуста", headerRowNum, sheetName)
	}

	return filteredHeaders, nil
}

// GetDataRows возвращает строки данных (начиная после строки заголовков)
// headerRowNum - номер строки заголовков (1-based index)
func (r *Reader) GetDataRows(sheetName string, headerRowNum int) ([][]string, error) {
	rows, err := r.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	if len(rows) <= headerRowNum {
		// Нет данных после заголовков
		return [][]string{}, nil
	}

	// Возвращаем строки начиная с headerRowNum (все что после заголовков)
	return rows[headerRowNum:], nil
}

// GetCellValue возвращает значение указанной ячейки
func (r *Reader) GetCellValue(sheetName, cell string) (string, error) {
	if !r.SheetExists(sheetName) {
		return "", apperrors.NewSheetNotFoundError(sheetName, r.path)
	}

	value, err := r.file.GetCellValue(sheetName, cell)
	if err != nil {
		return "", fmt.Errorf("failed to get cell value %s: %w", cell, err)
	}

	return value, nil
}

// GetRowCount возвращает количество строк на листе
func (r *Reader) GetRowCount(sheetName string) (int, error) {
	rows, err := r.GetRows(sheetName)
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

// ValidateFile проверяет базовую валидность файла
func (r *Reader) ValidateFile() error {
	sheets := r.GetSheetNames()
	if len(sheets) == 0 {
		return apperrors.NewEmptyFileError(r.path)
	}
	return nil
}

// GetFilePath возвращает путь к открытому файлу
func (r *Reader) GetFilePath() string {
	return r.path
}

// ValidateStructure проверяет совместимость структуры с базовым файлом
// baseHeaders - заголовки из базового файла для сравнения
func (r *Reader) ValidateStructure(sheetName string, headerRowNum int, baseHeaders []string) error {
	headers, err := r.GetHeaderRow(sheetName, headerRowNum)
	if err != nil {
		return err
	}

	// Проверяем количество заголовков
	if len(headers) != len(baseHeaders) {
		return fmt.Errorf("несовпадение количества столбцов на листе '%s': ожидается %d, получено %d",
			sheetName, len(baseHeaders), len(headers))
	}

	// Проверяем совпадение заголовков
	for i, header := range headers {
		if header != baseHeaders[i] {
			return fmt.Errorf("несовпадение заголовка столбца %d на листе '%s': ожидается '%s', получено '%s'",
				i+1, sheetName, baseHeaders[i], header)
		}
	}

	return nil
}
