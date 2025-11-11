package excel

import (
	"fmt"

	"github.com/xuri/excelize/v2"

	apperrors "github.com/DatKorso/Merge-excel/internal/errors"
)

// Writer предоставляет методы для записи Excel файлов
type Writer struct {
	file *excelize.File
}

// NewWriter создает новый Writer
func NewWriter() *Writer {
	return &Writer{
		file: excelize.NewFile(),
	}
}

// NewWriterFromFile создает Writer на основе существующего файла
func NewWriterFromFile(path string) (*Writer, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, apperrors.NewFileReadError(path, err)
	}

	return &Writer{
		file: f,
	}, nil
}

// Close закрывает файл
func (w *Writer) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// CreateSheet создает новый лист с указанным именем
func (w *Writer) CreateSheet(sheetName string) error {
	index, err := w.file.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("failed to create sheet '%s': %w", sheetName, err)
	}

	// Устанавливаем активным первый лист (если это первый созданный лист)
	if index == 0 {
		w.file.SetActiveSheet(index)
	}

	return nil
}

// DeleteSheet удаляет лист
func (w *Writer) DeleteSheet(sheetName string) error {
	if err := w.file.DeleteSheet(sheetName); err != nil {
		return fmt.Errorf("failed to delete sheet '%s': %w", sheetName, err)
	}
	return nil
}

// WriteHeaderRow записывает строку заголовков
func (w *Writer) WriteHeaderRow(sheetName string, rowNum int, headers []string) error {
	for colIdx, header := range headers {
		cell, err := excelize.CoordinatesToCellName(colIdx+1, rowNum)
		if err != nil {
			return fmt.Errorf("failed to get cell name: %w", err)
		}

		if err := w.file.SetCellValue(sheetName, cell, header); err != nil {
			return fmt.Errorf("failed to write header to cell %s: %w", cell, err)
		}
	}

	return nil
}

// WriteRow записывает одну строку данных
func (w *Writer) WriteRow(sheetName string, rowNum int, data []string) error {
	for colIdx, value := range data {
		cell, err := excelize.CoordinatesToCellName(colIdx+1, rowNum)
		if err != nil {
			return fmt.Errorf("failed to get cell name: %w", err)
		}

		if err := w.file.SetCellValue(sheetName, cell, value); err != nil {
			return fmt.Errorf("failed to write value to cell %s: %w", cell, err)
		}
	}

	return nil
}

// WriteRows записывает множество строк данных
func (w *Writer) WriteRows(sheetName string, startRow int, rows [][]string) error {
	for i, row := range rows {
		if err := w.WriteRow(sheetName, startRow+i, row); err != nil {
			return err
		}
	}
	return nil
}

// SetCellValue устанавливает значение ячейки
func (w *Writer) SetCellValue(sheetName, cell string, value interface{}) error {
	if err := w.file.SetCellValue(sheetName, cell, value); err != nil {
		return fmt.Errorf("failed to set cell value %s: %w", cell, err)
	}
	return nil
}

// SetColumnWidth устанавливает ширину столбца
func (w *Writer) SetColumnWidth(sheetName, startCol, endCol string, width float64) error {
	if err := w.file.SetColWidth(sheetName, startCol, endCol, width); err != nil {
		return fmt.Errorf("failed to set column width: %w", err)
	}
	return nil
}

// AutoFilter добавляет автофильтр на указанный диапазон
func (w *Writer) AutoFilter(sheetName, rangeRef string) error {
	if err := w.file.AutoFilter(sheetName, rangeRef, []excelize.AutoFilterOptions{}); err != nil {
		return fmt.Errorf("failed to add auto filter: %w", err)
	}
	return nil
}

// SetActiveSheet устанавливает активный лист
func (w *Writer) SetActiveSheet(sheetName string) error {
	index, err := w.file.GetSheetIndex(sheetName)
	if err != nil {
		return fmt.Errorf("failed to get sheet index for '%s': %w", sheetName, err)
	}

	w.file.SetActiveSheet(index)
	return nil
}

// Save сохраняет файл по указанному пути
func (w *Writer) Save(path string) error {
	if err := w.file.SaveAs(path); err != nil {
		return apperrors.NewSaveError(path, err)
	}
	return nil
}

// GetFile возвращает внутренний объект excelize.File для продвинутых операций
func (w *Writer) GetFile() *excelize.File {
	return w.file
}

// CopySheet копирует лист из другого файла
func (w *Writer) CopySheet(sourceFile *excelize.File, sourceSheet, targetSheet string) error {
	// Создаем новый лист
	if err := w.CreateSheet(targetSheet); err != nil {
		return err
	}

	// Получаем все строки из исходного листа
	rows, err := sourceFile.GetRows(sourceSheet)
	if err != nil {
		return fmt.Errorf("failed to read source sheet '%s': %w", sourceSheet, err)
	}

	// Записываем строки в новый лист
	if err := w.WriteRows(targetSheet, 1, rows); err != nil {
		return err
	}

	return nil
}

// MergeSheetData объединяет данные в существующий лист
// Добавляет строки данных после существующих строк
func (w *Writer) MergeSheetData(sheetName string, headerRow int, newRows [][]string) error {
	// Получаем текущие строки
	existingRows, err := w.file.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to get existing rows: %w", err)
	}

	// Определяем начальную строку для новых данных
	startRow := len(existingRows) + 1

	// Записываем новые строки
	if err := w.WriteRows(sheetName, startRow, newRows); err != nil {
		return err
	}

	return nil
}

// GetSheetNames возвращает список всех листов
func (w *Writer) GetSheetNames() []string {
	return w.file.GetSheetList()
}

// SheetExists проверяет существование листа
func (w *Writer) SheetExists(sheetName string) bool {
	for _, name := range w.GetSheetNames() {
		if name == sheetName {
			return true
		}
	}
	return false
}

// SetSheetRow записывает целую строку за один раз (более эффективно)
func (w *Writer) SetSheetRow(sheetName, cell string, values []interface{}) error {
	if err := w.file.SetSheetRow(sheetName, cell, &values); err != nil {
		return fmt.Errorf("failed to set sheet row: %w", err)
	}
	return nil
}

// WriteSheetWithData создает лист и записывает в него данные
// Первая строка - заголовки, остальные - данные
func (w *Writer) WriteSheetWithData(sheetName string, headerRow int, headers []string, data [][]string) error {
	// Создаем лист, если его нет
	if !w.SheetExists(sheetName) {
		if err := w.CreateSheet(sheetName); err != nil {
			return err
		}
	}

	// Записываем заголовки
	if err := w.WriteHeaderRow(sheetName, headerRow, headers); err != nil {
		return err
	}

	// Записываем данные
	if len(data) > 0 {
		if err := w.WriteRows(sheetName, headerRow+1, data); err != nil {
			return err
		}
	}

	return nil
}
