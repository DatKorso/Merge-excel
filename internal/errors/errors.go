package errors

import "fmt"

// Коды ошибок
const (
	ErrCodeFileNotFound     = "E001"
	ErrCodeFileReadError    = "E002"
	ErrCodeSheetNotFound    = "E003"
	ErrCodeInvalidHeaderRow = "E004"
	ErrCodeEmptyFile        = "E005"
	ErrCodeInvalidFormat    = "E006"
	ErrCodePermissionDenied = "E007"
	ErrCodeFileCorrupted    = "E008"
	ErrCodeConfigError      = "E009"
	ErrCodeMergeError       = "E010"
	ErrCodeSaveError        = "E011"
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

// NewFileNotFoundError создает ошибку "файл не найден"
func NewFileNotFoundError(path string) *AppError {
	return &AppError{
		Code:    ErrCodeFileNotFound,
		Message: "Файл не найден",
		Context: map[string]interface{}{"path": path},
	}
}

// NewSheetNotFoundError создает ошибку "лист не найден"
func NewSheetNotFoundError(sheet, file string) *AppError {
	return &AppError{
		Code:    ErrCodeSheetNotFound,
		Message: fmt.Sprintf("Лист '%s' не найден в файле", sheet),
		Context: map[string]interface{}{"sheet": sheet, "file": file},
	}
}

// NewInvalidHeaderRowError создает ошибку "неверный номер строки заголовков"
func NewInvalidHeaderRowError(row int) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidHeaderRow,
		Message: fmt.Sprintf("Неверный номер строки заголовков: %d", row),
		Context: map[string]interface{}{"row": row},
	}
}

// NewFileReadError создает ошибку чтения файла
func NewFileReadError(path string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeFileReadError,
		Message: "Ошибка при чтении файла",
		Context: map[string]interface{}{"path": path},
		Err:     err,
	}
}

// NewEmptyFileError создает ошибку "файл пустой"
func NewEmptyFileError(path string) *AppError {
	return &AppError{
		Code:    ErrCodeEmptyFile,
		Message: "Файл пустой или не содержит данных",
		Context: map[string]interface{}{"path": path},
	}
}

// NewInvalidFormatError создает ошибку "неверный формат файла"
func NewInvalidFormatError(path string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidFormat,
		Message: "Неверный формат файла. Поддерживаются только .xlsx файлы",
		Context: map[string]interface{}{"path": path},
	}
}

// NewPermissionDeniedError создает ошибку "нет доступа"
func NewPermissionDeniedError(path string) *AppError {
	return &AppError{
		Code:    ErrCodePermissionDenied,
		Message: "Нет доступа к файлу",
		Context: map[string]interface{}{"path": path},
	}
}

// NewFileCorruptedError создает ошибку "файл поврежден"
func NewFileCorruptedError(path string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeFileCorrupted,
		Message: "Файл поврежден и не может быть прочитан",
		Context: map[string]interface{}{"path": path},
		Err:     err,
	}
}

// NewConfigError создает ошибку конфигурации
func NewConfigError(message string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeConfigError,
		Message: message,
		Err:     err,
	}
}

// NewMergeError создает ошибку объединения
func NewMergeError(message string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeMergeError,
		Message: message,
		Err:     err,
	}
}

// NewSaveError создает ошибку сохранения файла
func NewSaveError(path string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeSaveError,
		Message: "Не удалось сохранить файл",
		Context: map[string]interface{}{"path": path},
		Err:     err,
	}
}

// UserMessage возвращает понятное пользователю сообщение об ошибке
func UserMessage(code string) string {
	messages := map[string]string{
		ErrCodeFileNotFound:     "Файл не найден. Пожалуйста, проверьте путь к файлу.",
		ErrCodeFileReadError:    "Не удалось прочитать файл. Возможно, он поврежден или открыт в другой программе.",
		ErrCodeSheetNotFound:    "Указанный лист не найден в файле. Проверьте настройки.",
		ErrCodeInvalidHeaderRow: "Неверный номер строки заголовков. Укажите значение от 1 и выше.",
		ErrCodeEmptyFile:        "Файл пустой или не содержит данных.",
		ErrCodeInvalidFormat:    "Неверный формат файла. Поддерживаются только .xlsx файлы.",
		ErrCodePermissionDenied: "Нет доступа к файлу. Проверьте права доступа.",
		ErrCodeFileCorrupted:    "Файл поврежден и не может быть прочитан.",
		ErrCodeConfigError:      "Ошибка конфигурации. Проверьте настройки профиля.",
		ErrCodeMergeError:       "Ошибка при объединении файлов. Проверьте логи.",
		ErrCodeSaveError:        "Не удалось сохранить файл. Проверьте путь и права доступа.",
	}

	if msg, exists := messages[code]; exists {
		return msg
	}
	return "Произошла неизвестная ошибка"
}
