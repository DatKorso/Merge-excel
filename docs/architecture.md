# Архитектура приложения

## Общая структура

```
┌─────────────────────────────────────────────────────┐
│         GUI Layer (Fyne)                            │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────┐ │
│  │Tab 1:        │  │Tab 2:        │  │Tab 3:     │ │
│  │Base File     │→ │Add Files     │→ │Merge &    │ │
│  │Analysis      │  │              │  │Save       │ │
│  └──────┬───────┘  └──────┬───────┘  └─────┬─────┘ │
└─────────┼──────────────────┼────────────────┼───────┘
          │                  │                │
          ▼                  ▼                ▼
┌─────────────────────────────────────────────────────┐
│       Business Logic Layer                          │
│  ┌──────────────┐  ┌──────────────┐                │
│  │BaseAnalyzer  │  │Merger        │                │
│  │(rules config)│  │(apply rules) │                │
│  └──────┬───────┘  └──────┬───────┘                │
└─────────┼──────────────────┼─────────────────────────┘
          │                  │
          ▼                  ▼
┌─────────────────────────────────────────────────────┐
│       Data Access Layer                             │
│  ┌──────────────┐  ┌──────────────┐                │
│  │ExcelReader   │  │Exporter      │                │
│  │(excelize)    │  │(save result) │                │
│  └──────────────┘  └──────────────┘                │
└─────────────────────────────────────────────────────┘
```

## Основные компоненты

### 1. GUI Layer (Слой представления) - Fyne AppTabs

#### Tab 1: BaseFileTab
**Назначение:** Выбор и анализ базового файла

**Компоненты:**
- `widget.Button` для выбора базового Excel файла
- `widget.Tree` для отображения листов с чекбоксами
- `widget.Entry` (number) для указания номера строки заголовков для каждого листа
- `widget.Table` для предпросмотра заголовков
- `widget.Button` "Сохранить профиль" и "Загрузить профиль"

**Функционал:**
```go
type BaseFileTab struct {
    file          *widget.Label
    sheetsTree    *widget.Tree
    previewTable  *widget.Table
    analyzer      *core.BaseAnalyzer
}

func (t *BaseFileTab) SelectBaseFile() error
func (t *BaseFileTab) AnalyzeFile(path string) error
func (t *BaseFileTab) PreviewHeaders(sheet string, headerRow int) error
func (t *BaseFileTab) SaveProfile(name string) error
func (t *BaseFileTab) LoadProfile(path string) error
```

#### Tab 2: FileListTab
**Назначение:** Добавление файлов для объединения

**Компоненты:**
- `widget.List` для отображения добавленных файлов
- `widget.Button` "Добавить файлы"
- `widget.Button` "Удалить выбранное"
- `widget.Button` "Очистить все"
- Поддержка Drag & Drop через `storage.Dropper`

**Функционал:**
```go
type FileListTab struct {
    fileList    *widget.List
    files       []string
}

func (t *FileListTab) AddFiles(paths []string)
func (t *FileListTab) RemoveSelected()
func (t *FileListTab) ClearAll()
func (t *FileListTab) GetFiles() []string
```

#### Tab 3: MergeTab
**Назначение:** Объединение и сохранение

**Компоненты:**
- `widget.Button` "Объединить"
- `widget.ProgressBar` для индикации прогресса
- `widget.Table` для предпросмотра результата
- `widget.Label` для статистики
- `widget.Button` "Сохранить как..."

**Функционал:**
```go
type MergeTab struct {
    progressBar  *widget.ProgressBar
    statusLabel  *widget.Label
    previewTable *widget.Table
    merger       *core.Merger
}

func (t *MergeTab) StartMerge(files []string, config *core.Config)
func (t *MergeTab) UpdateProgress(current, total int, message string)
func (t *MergeTab) ShowPreview(data map[string][][]string)
func (t *MergeTab) SaveResult(path string) error
```

### 2. Business Logic Layer

#### BaseAnalyzer
```go
package core

type SheetConfig struct {
    SheetName string   `json:"sheet_name"`
    HeaderRow int      `json:"header_row"` // 1-based
    Enabled   bool     `json:"enabled"`
    Headers   []string `json:"headers"`
}

type BaseAnalyzer struct {
    filePath     string
    sheetsConfig map[string]*SheetConfig
}

func NewBaseAnalyzer(filePath string) *BaseAnalyzer

// GetAllSheets возвращает список всех листов в файле
func (a *BaseAnalyzer) GetAllSheets() ([]string, error)

// SetSheetConfig настраивает конфигурацию для листа
func (a *BaseAnalyzer) SetSheetConfig(sheetName string, headerRow int, enabled bool) error

// GetHeaders получает заголовки из указанной строки
func (a *BaseAnalyzer) GetHeaders(sheetName string, headerRow int) ([]string, error)

// ExportConfig экспортирует конфигурацию для сохранения
func (a *BaseAnalyzer) ExportConfig() map[string]*SheetConfig

// ImportConfig импортирует конфигурацию
func (a *BaseAnalyzer) ImportConfig(config map[string]*SheetConfig)
```

#### Merger
```go
package core

type ProgressCallback func(current, total int, message string)

type Merger struct {
    baseConfig       map[string]*SheetConfig
    progressCallback ProgressCallback
}

func NewMerger(config map[string]*SheetConfig) *Merger

// MergeFiles объединяет все файлы согласно конфигурации
// Возвращает map где ключ = имя листа, значение = данные
func (m *Merger) MergeFiles(filePaths []string) (map[string][][]string, error)

// mergeSingleSheet объединяет один лист из всех файлов
func (m *Merger) mergeSingleSheet(sheetName string, config *SheetConfig, files []string) ([][]string, error)

// SetProgressCallback устанавливает callback для обновления прогресса
func (m *Merger) SetProgressCallback(callback ProgressCallback)
```

### 3. Data Access Layer

#### ExcelReader
```go
package excel

import "github.com/xuri/excelize/v2"

type Reader struct{}

func NewReader() *Reader

// GetSheetNames получает список всех листов в файле
func (r *Reader) GetSheetNames(filePath string) ([]string, error)

// ReadSheet читает лист из файла
// headerRow - номер строки с заголовками (1-based)
func (r *Reader) ReadSheet(filePath, sheetName string, headerRow int) ([][]string, error)

// ReadHeaders читает только заголовки из указанной строки
func (r *Reader) ReadHeaders(filePath, sheetName string, headerRow int) ([]string, error)

// ValidateFile проверяет, что файл является валидным Excel файлом
func (r *Reader) ValidateFile(filePath string) error
```

#### Exporter
```go
package excel

type Exporter struct{}

func NewExporter() *Exporter

// SaveToExcel сохраняет результат в Excel файл
// data: map[sheetName]rows
func (e *Exporter) SaveToExcel(data map[string][][]string, outputPath string) error

// GetFileInfo получает информацию о сохраненном файле
func (e *Exporter) GetFileInfo(outputPath string) (FileInfo, error)

type FileInfo struct {
    Size       int64
    SheetCount int
    RowCount   map[string]int
}
```

#### ConfigManager
```go
package config

import "encoding/json"

type Manager struct {
    configDir string
}

func NewManager(configDir string) *Manager

// SaveProfile сохраняет профиль конфигурации в JSON
func (m *Manager) SaveProfile(config map[string]*core.SheetConfig, profileName string) error

// LoadProfile загружает профиль конфигурации из JSON
func (m *Manager) LoadProfile(profilePath string) (map[string]*core.SheetConfig, error)

// ListProfiles получает список всех сохраненных профилей
func (m *Manager) ListProfiles() ([]string, error)
```

## Паттерны проектирования

### 1. Tabbed Interface Pattern (Fyne AppTabs)
Пошаговый интерфейс через вкладки для упрощения процесса:
- **Tab 1:** Анализ базового файла и настройка правил
- **Tab 2:** Добавление файлов для объединения
- **Tab 3:** Выполнение объединения и сохранение результата

### 2. Observer Pattern (через каналы Go)
Для обновления прогресса операций:
```go
type ProgressUpdate struct {
    Current int
    Total   int
    Message string
}

// В Merger
func (m *Merger) MergeFiles(files []string) (map[string][][]string, error) {
    progressChan := make(chan ProgressUpdate)
    
    go func() {
        // Обработка файлов
        for i, file := range files {
            progressChan <- ProgressUpdate{
                Current: i + 1,
                Total:   len(files),
                Message: fmt.Sprintf("Processing %s", file),
            }
        }
        close(progressChan)
    }()
    
    // В GUI слушаем канал и обновляем UI
    return result, nil
}
```

### 3. Configuration Pattern
Для управления правилами объединения:
```go
// JSON конфигурация
type Config struct {
    BaseFile string                    `json:"base_file"`
    Sheets   map[string]*SheetConfig   `json:"sheets"`
}

// Пример
{
    "base_file": "sales_2024_q1.xlsx",
    "sheets": {
        "Продажи": {
            "enabled": true,
            "header_row": 1,
            "headers": ["Дата", "Товар", "Количество", "Сумма"]
        },
        "Возвраты": {
            "enabled": true,
            "header_row": 2,
            "headers": ["Дата", "Товар", "Причина"]
        }
    }
}
```

### 4. Repository Pattern
Изоляция работы с данными:
```go
type ExcelRepository interface {
    GetSheetNames(filePath string) ([]string, error)
    ReadSheet(filePath, sheetName string, headerRow int) ([][]string, error)
    SaveToExcel(data map[string][][]string, outputPath string) error
}

// Реализация через excelize
type ExcelizeRepository struct {
    reader   *excel.Reader
    exporter *excel.Exporter
}
```

## Поток данных

### Типичный сценарий использования

```
┌─────────────────────────────────────────────────────────────┐
│ ШАГ 1: Выбор и настройка базового файла                    │
└─────────────────────────────────────────────────────────────┘
1. Пользователь выбирает базовый файл
   └─> BaseFileStep.select_base_file()
       └─> BaseAnalyzer(file_path)
           └─> ExcelReader.get_sheet_names()
               └─> Отображение в QTreeWidget

2. Пользователь настраивает каждый лист
   └─> BaseFileStep.on_sheet_selected()
       └─> BaseAnalyzer.set_sheet_config(sheet, header_row, enabled)
           └─> ExcelReader.read_headers(sheet, header_row)
               └─> Предпросмотр в QTableWidget

3. (Опционально) Сохранение профиля
   └─> BaseFileStep.save_profile()
       └─> ConfigManager.save_profile(config)

┌─────────────────────────────────────────────────────────────┐
│ ШАГ 2: Добавление файлов для объединения                   │
└─────────────────────────────────────────────────────────────┘
4. Пользователь добавляет файлы (drag & drop или кнопка)
   └─> FileListStep.add_files(paths)
       └─> Валидация каждого файла
           └─> ExcelReader.validate_file(path)

5. (Опционально) Изменение порядка файлов

┌─────────────────────────────────────────────────────────────┐
│ ШАГ 3: Объединение и сохранение                            │
└─────────────────────────────────────────────────────────────┘
6. Пользователь нажимает "Объединить"
   └─> MergeStep.start_merge()
       └─> MergeFacade.merge_all(files, progress_callback)
           └─> ExcelMerger.merge_files()
               ├─> Для каждого листа:
               │   └─> ExcelMerger._merge_single_sheet()
               │       ├─> Чтение базового файла
               │       ├─> Чтение остальных файлов
               │       └─> pd.concat([базовый, файл1, файл2, ...])
               └─> Обновление прогресса через callback

7. Отображение предпросмотра
   └─> MergeStep.show_preview(result)
       └─> Отображение первых 100 строк в QTableWidget

8. Пользователь нажимает "Сохранить как..."
   └─> MergeStep.save_result()
       └─> ExcelExporter.save_to_excel(data, output_path)
           └─> Уведомление об успехе
```

## Обработка ошибок

```go
package errors

import "fmt"

// Базовые ошибки
var (
    ErrFileRead       = fmt.Errorf("file read error")
    ErrSheetNotFound  = fmt.Errorf("sheet not found")
    ErrInvalidHeader  = fmt.Errorf("invalid header row")
    ErrConfigError    = fmt.Errorf("configuration error")
)

// SheetNotFoundError - лист не найден в файле
type SheetNotFoundError struct {
    FilePath  string
    SheetName string
}

func (e *SheetNotFoundError) Error() string {
    return fmt.Sprintf("sheet '%s' not found in file %s", e.SheetName, e.FilePath)
}

// InvalidHeaderRowError - неверный номер строки заголовков
type InvalidHeaderRowError struct {
    Row int
}

func (e *InvalidHeaderRowError) Error() string {
    return fmt.Sprintf("invalid header row: %d (must be >= 1)", e.Row)
}
```

### Стратегия обработки ошибок

1. **Отсутствие листа в файле:**
```go
func (m *Merger) mergeSingleSheet(sheetName string, config *SheetConfig, files []string) ([][]string, error) {
    for _, file := range files {
        sheets, err := m.reader.GetSheetNames(file)
        if err != nil {
            log.Printf("Warning: cannot read %s: %v", file, err)
            continue // Пропускаем файл
        }
        
        if !contains(sheets, sheetName) {
            log.Printf("Warning: sheet '%s' not found in %s", sheetName, file)
            continue // Пропускаем файл
        }
        // Обработка...
    }
}
```

2. **Поврежденный файл:**
```go
func (r *Reader) ValidateFile(filePath string) error {
    f, err := excelize.OpenFile(filePath)
    if err != nil {
        return fmt.Errorf("cannot open file: %w", err)
    }
    defer f.Close()
    return nil
}
```

3. **Использование defer для очистки ресурсов:**
```go
func (r *Reader) ReadSheet(filePath, sheetName string, headerRow int) ([][]string, error) {
    f, err := excelize.OpenFile(filePath)
    if err != nil {
        return nil, err
    }
    defer func() {
        if err := f.Close(); err != nil {
            log.Printf("Error closing file: %v", err)
        }
    }()
    
    // Чтение данных...
}
```

## Конфигурация

### Профиль объединения (merge_profile.json)
```json
{
  "profile_name": "Quarterly Sales Report",
  "created_at": "2024-11-11T10:30:00",
  "base_file": "sales_2024_q1.xlsx",
  "sheets": {
    "Продажи": {
      "enabled": true,
      "header_row": 1,
      "headers": ["Дата", "Регион", "Товар", "Количество", "Сумма"]
    },
    "Возвраты": {
      "enabled": true,
      "header_row": 2,
      "headers": ["Дата", "Номер заказа", "Причина", "Сумма"]
    },
    "Отчет": {
      "enabled": false,
      "header_row": 1,
      "headers": []
    }
  }
}
```

### Настройки приложения (app_config.json)
```json
{
  "app": {
    "theme": "light",
    "language": "ru",
    "last_output_directory": "/Users/user/Documents"
  },
  "preview": {
    "max_rows": 100,
    "show_index": false
  },
  "processing": {
    "skip_empty_rows": true,
    "show_warnings": true
  }
}
```

## Логирование

```go
package logger

import (
    "log/slog"
    "os"
)

var Log *slog.Logger

func Init(logFile string) error {
    f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        return err
    }
    
    Log = slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    
    return nil
}

// Использование
logger.Log.Info("merging files", 
    "count", len(files),
    "base_file", baseFile)
    
logger.Log.Warn("sheet not found",
    "sheet", sheetName,
    "file", filePath)
    
logger.Log.Error("failed to read file",
    "file", filePath,
    "error", err)
```

## Тестирование

### Unit Tests

#### excel/reader_test.go
```go
func TestGetSheetNames(t *testing.T) {
    // Проверка получения списка листов
}

func TestReadSheetWithHeaderRow(t *testing.T) {
    // Проверка чтения с указанной строкой заголовков
}

func TestReadHeaders(t *testing.T) {
    // Проверка чтения только заголовков
}
```

#### core/analyzer_test.go
```go
func TestAnalyzeBaseFile(t *testing.T) {
    // Проверка анализа базового файла
}

func TestSetSheetConfig(t *testing.T) {
    // Проверка настройки конфигурации листа
}

func TestExportImportConfig(t *testing.T) {
    // Проверка экспорта и импорта конфигурации
}
```

#### core/merger_test.go
```go
func TestMergeSingleSheet(t *testing.T) {
    // Проверка объединения одного листа из нескольких файлов
}

func TestMergeMultipleSheets(t *testing.T) {
    // Проверка объединения нескольких листов
}

func TestMergeWithDifferentHeaderRows(t *testing.T) {
    // Проверка объединения с разными строками заголовков
}
```

### Benchmarks

```go
func BenchmarkMergeFiles(b *testing.B) {
    // Бенчмарк производительности объединения
    for i := 0; i < b.N; i++ {
        merger.MergeFiles(testFiles)
    }
}

func BenchmarkReadLargeFile(b *testing.B) {
    // Бенчмарк чтения больших файлов
}
```

### Integration Tests
```go
func TestFullMergeCycle(t *testing.T) {
    // Полный цикл: анализ → добавление файлов → объединение → сохранение
}

func TestRealExcelFiles(t *testing.T) {
    // Тестирование с реальными Excel файлами
}

func TestErrorHandling(t *testing.T) {
    // Проверка обработки ошибок (отсутствующие листы, поврежденные файлы)
}
```

### Запуск тестов
```bash
# Все тесты
go test ./...

# С покрытием
go test -cover ./...

# Verbose
go test -v ./...

# Бенчмарки
go test -bench=. ./...
```
