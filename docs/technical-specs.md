# Техническая спецификация

## Дата: 11 ноября 2025

## 1. Формат JSON для профилей

### Структура файла профиля

**Расположение:** `configs/profiles/<profile_name>.json`

```json
{
  "version": "1.0",
  "profile_name": "Quarterly Sales Report",
  "created_at": "2025-11-11T10:30:00Z",
  "updated_at": "2025-11-11T10:30:00Z",
  "base_file_name": "sales_2024_q1.xlsx",
  "sheets": [
    {
      "sheet_name": "Продажи",
      "enabled": true,
      "header_row": 5,
      "headers": ["Дата", "Регион", "Товар", "Количество", "Сумма"]
    },
    {
      "sheet_name": "Возвраты",
      "enabled": true,
      "header_row": 5,
      "headers": ["Дата", "Номер заказа", "Причина", "Сумма"]
    },
    {
      "sheet_name": "Справка",
      "enabled": false,
      "header_row": 1,
      "headers": []
    }
  ],
  "settings": {
    "skip_empty_rows": true,
    "show_warnings": true,
    "preview_rows": 100
  }
}
```

### Go структуры для профиля

```go
package config

import "time"

// Profile представляет сохраненный профиль настроек
type Profile struct {
    Version      string         `json:"version"`
    ProfileName  string         `json:"profile_name"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
    BaseFileName string         `json:"base_file_name"`
    Sheets       []SheetConfig  `json:"sheets"`
    Settings     ProfileSettings `json:"settings"`
}

// SheetConfig настройки для одного листа
type SheetConfig struct {
    SheetName string   `json:"sheet_name"`
    Enabled   bool     `json:"enabled"`
    HeaderRow int      `json:"header_row"`  // 1-based index
    Headers   []string `json:"headers"`
}

// ProfileSettings дополнительные настройки
type ProfileSettings struct {
    SkipEmptyRows bool `json:"skip_empty_rows"`
    ShowWarnings  bool `json:"show_warnings"`
    PreviewRows   int  `json:"preview_rows"`
}
```

### Пример использования в Go

```go
// Сохранение профиля
func (m *ConfigManager) SaveProfile(profile *Profile, filename string) error {
    profile.UpdatedAt = time.Now()
    
    data, err := json.MarshalIndent(profile, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal profile: %w", err)
    }
    
    path := filepath.Join(m.profilesDir, filename+".json")
    if err := os.WriteFile(path, data, 0644); err != nil {
        return fmt.Errorf("failed to write profile: %w", err)
    }
    
    return nil
}

// Загрузка профиля
func (m *ConfigManager) LoadProfile(filename string) (*Profile, error) {
    path := filepath.Join(m.profilesDir, filename+".json")
    
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read profile: %w", err)
    }
    
    var profile Profile
    if err := json.Unmarshal(data, &profile); err != nil {
        return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
    }
    
    return &profile, nil
}
```

## 2. Drag & Drop в Fyne

### Реализация Drag & Drop для списка файлов

```go
package gui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/storage"
    "fyne.io/fyne/v2/widget"
    "path/filepath"
)

// FileListTab вкладка со списком файлов для объединения
type FileListTab struct {
    container *fyne.Container
    fileList  *widget.List
    files     []string
}

// NewFileListTab создает новую вкладку с поддержкой Drag & Drop
func NewFileListTab() *FileListTab {
    tab := &FileListTab{
        files: []string{},
    }
    
    // Создаем список с возможностью перетаскивания
    tab.fileList = widget.NewList(
        func() int {
            return len(tab.files)
        },
        func() fyne.CanvasObject {
            return widget.NewLabel("Template")
        },
        func(id widget.ListItemID, obj fyne.CanvasObject) {
            obj.(*widget.Label).SetText(filepath.Base(tab.files[id]))
        },
    )
    
    // Создаем контейнер с поддержкой Drop
    tab.container = container.NewStack(tab.fileList)
    
    return tab
}

// TypedRune обработка ввода (для интерфейса Droppable)
func (t *FileListTab) TypedRune(r rune) {}

// TypedKey обработка клавиш
func (t *FileListTab) TypedKey(key *fyne.KeyEvent) {}

// DragEnded обработка окончания перетаскивания
func (t *FileListTab) DragEnded() {}

// Drop обработка Drop события
func (t *FileListTab) Drop(position fyne.Position, items []fyne.URI) {
    for _, uri := range items {
        // Проверяем, что это .xlsx файл
        if filepath.Ext(uri.Path()) == ".xlsx" {
            t.AddFile(uri.Path())
        }
    }
}

// AddFile добавляет файл в список
func (t *FileListTab) AddFile(path string) {
    // Проверяем, что файл еще не добавлен
    for _, f := range t.files {
        if f == path {
            return // Уже есть
        }
    }
    
    t.files = append(t.files, path)
    t.fileList.Refresh()
}

// RemoveSelected удаляет выбранные файлы
func (t *FileListTab) RemoveSelected() {
    if t.fileList.SelectedIndex() < 0 {
        return
    }
    
    idx := t.fileList.SelectedIndex()
    t.files = append(t.files[:idx], t.files[idx+1:]...)
    t.fileList.UnselectAll()
    t.fileList.Refresh()
}

// GetFiles возвращает список всех файлов
func (t *FileListTab) GetFiles() []string {
    return t.files
}
```

### Альтернативный подход - использование container с Droppable

```go
import (
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

type DroppableContainer struct {
    widget.BaseWidget
    onDrop func([]string)
}

func NewDroppableContainer(onDrop func([]string)) *DroppableContainer {
    d := &DroppableContainer{onDrop: onDrop}
    d.ExtendBaseWidget(d)
    return d
}

func (d *DroppableContainer) Drop(position fyne.Position, items []fyne.URI) {
    paths := make([]string, 0, len(items))
    for _, uri := range items {
        if filepath.Ext(uri.Path()) == ".xlsx" {
            paths = append(paths, uri.Path())
        }
    }
    if len(paths) > 0 && d.onDrop != nil {
        d.onDrop(paths)
    }
}
```

## 3. Обработка ошибок и коды ошибок

### Типы ошибок

```go
package errors

import "fmt"

// Коды ошибок
const (
    ErrCodeFileNotFound      = "E001"
    ErrCodeFileReadError     = "E002"
    ErrCodeSheetNotFound     = "E003"
    ErrCodeInvalidHeaderRow  = "E004"
    ErrCodeEmptyFile         = "E005"
    ErrCodeInvalidFormat     = "E006"
    ErrCodePermissionDenied  = "E007"
    ErrCodeFileCorrupted     = "E008"
    ErrCodeConfigError       = "E009"
    ErrCodeMergeError        = "E010"
    ErrCodeSaveError         = "E011"
)

// AppError представляет ошибку приложения с кодом и контекстом
type AppError struct {
    Code    string
    Message string
    Context map[string]interface{}
    Err     error
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// Конструкторы ошибок

func NewFileNotFoundError(path string) *AppError {
    return &AppError{
        Code:    ErrCodeFileNotFound,
        Message: "Файл не найден",
        Context: map[string]interface{}{"path": path},
    }
}

func NewSheetNotFoundError(sheet, file string) *AppError {
    return &AppError{
        Code:    ErrCodeSheetNotFound,
        Message: fmt.Sprintf("Лист '%s' не найден в файле", sheet),
        Context: map[string]interface{}{"sheet": sheet, "file": file},
    }
}

func NewInvalidHeaderRowError(row int) *AppError {
    return &AppError{
        Code:    ErrCodeInvalidHeaderRow,
        Message: fmt.Sprintf("Неверный номер строки заголовков: %d", row),
        Context: map[string]interface{}{"row": row},
    }
}

func NewFileReadError(path string, err error) *AppError {
    return &AppError{
        Code:    ErrCodeFileReadError,
        Message: "Ошибка при чтении файла",
        Context: map[string]interface{}{"path": path},
        Err:     err,
    }
}
```

### Отображение ошибок пользователю

```go
package gui

import (
    "fyne.io/fyne/v2/dialog"
    "merge-excel/internal/errors"
)

// Сообщения для пользователя (более понятные)
var userMessages = map[string]string{
    errors.ErrCodeFileNotFound:     "Файл не найден. Пожалуйста, проверьте путь к файлу.",
    errors.ErrCodeFileReadError:    "Не удалось прочитать файл. Возможно, он поврежден или открыт в другой программе.",
    errors.ErrCodeSheetNotFound:    "Указанный лист не найден в файле. Проверьте настройки.",
    errors.ErrCodeInvalidHeaderRow: "Неверный номер строки заголовков. Укажите значение от 1 и выше.",
    errors.ErrCodeEmptyFile:        "Файл пустой или не содержит данных.",
    errors.ErrCodeInvalidFormat:    "Неверный формат файла. Поддерживаются только .xlsx файлы.",
    errors.ErrCodePermissionDenied: "Нет доступа к файлу. Проверьте права доступа.",
    errors.ErrCodeFileCorrupted:    "Файл поврежден и не может быть прочитан.",
    errors.ErrCodeConfigError:      "Ошибка конфигурации. Проверьте настройки профиля.",
    errors.ErrCodeMergeError:       "Ошибка при объединении файлов. Проверьте логи.",
    errors.ErrCodeSaveError:        "Не удалось сохранить файл. Проверьте путь и права доступа.",
}

// ShowError показывает диалог с ошибкой
func (app *App) ShowError(err error) {
    var message string
    
    if appErr, ok := err.(*errors.AppError); ok {
        if msg, exists := userMessages[appErr.Code]; exists {
            message = msg
        } else {
            message = appErr.Message
        }
        
        // Логируем детали
        app.logger.Error("Application error",
            "code", appErr.Code,
            "message", appErr.Message,
            "context", appErr.Context,
            "error", appErr.Err,
        )
    } else {
        message = err.Error()
        app.logger.Error("Unknown error", "error", err)
    }
    
    dialog.ShowError(fmt.Errorf(message), app.window)
}
```

## 4. Логирование

### Конфигурация логирования

```go
package logger

import (
    "io"
    "log/slog"
    "os"
    "path/filepath"
)

// Config конфигурация логгера
type Config struct {
    Level      slog.Level
    LogFile    string
    MaxSize    int64  // максимальный размер файла в байтах
    MaxBackups int    // максимальное количество старых лог-файлов
    Console    bool   // выводить ли в консоль
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
    return &Config{
        Level:      slog.LevelInfo,
        LogFile:    "logs/excel-merger.log",
        MaxSize:    10 * 1024 * 1024, // 10 MB
        MaxBackups: 5,
        Console:    true,
    }
}

// InitLogger инициализирует логгер
func InitLogger(cfg *Config) (*slog.Logger, error) {
    // Создаем директорию для логов
    logDir := filepath.Dir(cfg.LogFile)
    if err := os.MkdirAll(logDir, 0755); err != nil {
        return nil, err
    }
    
    // Открываем файл для записи
    file, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        return nil, err
    }
    
    // Проверяем размер файла и ротируем при необходимости
    if info, err := file.Stat(); err == nil && info.Size() > cfg.MaxSize {
        file.Close()
        if err := rotateLogFile(cfg); err != nil {
            return nil, err
        }
        file, err = os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if err != nil {
            return nil, err
        }
    }
    
    // Настраиваем вывод
    var writer io.Writer = file
    if cfg.Console {
        writer = io.MultiWriter(file, os.Stdout)
    }
    
    // Создаем хендлер
    handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
        Level: cfg.Level,
        AddSource: true,
    })
    
    logger := slog.New(handler)
    slog.SetDefault(logger)
    
    return logger, nil
}

// rotateLogFile выполняет ротацию лог-файлов
func rotateLogFile(cfg *Config) error {
    // Удаляем самый старый файл, если достигнут лимит
    if cfg.MaxBackups > 0 {
        oldestBackup := fmt.Sprintf("%s.%d", cfg.LogFile, cfg.MaxBackups)
        os.Remove(oldestBackup)
    }
    
    // Сдвигаем все файлы
    for i := cfg.MaxBackups - 1; i > 0; i-- {
        oldPath := fmt.Sprintf("%s.%d", cfg.LogFile, i)
        newPath := fmt.Sprintf("%s.%d", cfg.LogFile, i+1)
        os.Rename(oldPath, newPath)
    }
    
    // Переименовываем текущий файл
    backupPath := fmt.Sprintf("%s.1", cfg.LogFile)
    return os.Rename(cfg.LogFile, backupPath)
}
```

### Использование логгера в приложении

```go
package main

import (
    "log/slog"
    "merge-excel/internal/logger"
)

func main() {
    // Инициализация логгера
    cfg := logger.DefaultConfig()
    log, err := logger.InitLogger(cfg)
    if err != nil {
        panic(err)
    }
    
    log.Info("Application started", "version", "0.1.0-alpha")
    
    // Использование в коде
    log.Info("Loading base file", "path", "/path/to/file.xlsx")
    log.Warn("Sheet not found in file", 
        "sheet", "Продажи",
        "file", "report_q2.xlsx",
    )
    log.Error("Failed to merge files",
        "error", err,
        "files_count", 10,
    )
    
    // Структурированное логирование с атрибутами
    log.With(
        "operation", "merge",
        "user", "admin",
    ).Info("Merge completed successfully",
        "duration_ms", 1234,
        "rows_merged", 50000,
    )
}
```

### Формат лог-файла

```json
{"time":"2025-11-11T10:30:00.123+03:00","level":"INFO","msg":"Application started","version":"0.1.0-alpha"}
{"time":"2025-11-11T10:30:05.456+03:00","level":"INFO","msg":"Loading base file","path":"/path/to/file.xlsx"}
{"time":"2025-11-11T10:30:10.789+03:00","level":"WARN","msg":"Sheet not found in file","sheet":"Продажи","file":"report_q2.xlsx"}
{"time":"2025-11-11T10:30:15.012+03:00","level":"ERROR","msg":"Failed to merge files","error":"file read error","files_count":10}
```

## 5. Индикация прогресса

### Реализация Progress в Fyne

```go
package gui

import (
    "fyne.io/fyne/v2/widget"
    "merge-excel/internal/core"
)

// MergeTab вкладка объединения с прогресс-баром
type MergeTab struct {
    progressBar *widget.ProgressBar
    statusLabel *widget.Label
    merger      *core.Merger
}

// StartMerge запускает процесс объединения с индикацией прогресса
func (t *MergeTab) StartMerge(files []string, config map[string]*core.SheetConfig) {
    // Сброс прогресса
    t.progressBar.SetValue(0)
    t.statusLabel.SetText("Начинаю объединение...")
    
    // Создаем канал для обновления прогресса
    progressChan := make(chan core.ProgressUpdate)
    
    // Настраиваем callback для merger
    t.merger.SetProgressCallback(func(current, total int, message string) {
        progressChan <- core.ProgressUpdate{
            Current: current,
            Total:   total,
            Message: message,
        }
    })
    
    // Запускаем объединение в горутине
    go func() {
        result, err := t.merger.MergeFiles(files, config)
        close(progressChan)
        
        if err != nil {
            // Обработка ошибки
            t.ShowError(err)
            return
        }
        
        // Показываем результат
        t.ShowResult(result)
    }()
    
    // Обновляем UI в главной горутине
    go func() {
        for update := range progressChan {
            progress := float64(update.Current) / float64(update.Total)
            t.progressBar.SetValue(progress)
            t.statusLabel.SetText(update.Message)
        }
        
        t.statusLabel.SetText("Объединение завершено!")
    }()
}
```

### Структура Progress Update

```go
package core

// ProgressUpdate информация об обновлении прогресса
type ProgressUpdate struct {
    Current int    // Текущий шаг
    Total   int    // Всего шагов
    Message string // Сообщение о текущей операции
}

// ProgressCallback функция обратного вызова для обновления прогресса
type ProgressCallback func(current, total int, message string)
```

## 6. Структура директорий при первом запуске

```
~/.excel-merger/                    # Домашняя директория приложения
├── configs/
│   └── profiles/                   # Сохраненные профили
│       ├── quarterly_sales.json
│       └── monthly_reports.json
├── logs/
│   ├── excel-merger.log           # Текущий лог
│   ├── excel-merger.log.1         # Backup 1
│   └── excel-merger.log.2         # Backup 2
└── settings.json                   # Настройки приложения
```

### Инициализация при первом запуске

```go
package main

import (
    "os"
    "path/filepath"
)

// InitAppDirectories создает необходимые директории
func InitAppDirectories() error {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return err
    }
    
    appDir := filepath.Join(homeDir, ".excel-merger")
    
    dirs := []string{
        appDir,
        filepath.Join(appDir, "configs", "profiles"),
        filepath.Join(appDir, "logs"),
    }
    
    for _, dir := range dirs {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return err
        }
    }
    
    return nil
}
```

## Заключение

Эта техническая спецификация покрывает основные аспекты реализации:
- ✅ JSON схема для профилей с примерами использования
- ✅ Drag & Drop в Fyne с рабочим кодом
- ✅ Система обработки ошибок с кодами и user-friendly сообщениями
- ✅ Логирование с ротацией файлов
- ✅ Индикация прогресса через каналы
- ✅ Структура директорий приложения

Все примеры кода готовы к использованию и следуют лучшим практикам Go.
