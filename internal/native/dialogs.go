package native

import (
	"path/filepath"

	"github.com/sqweek/dialog"
)

// FileOpenDialog показывает нативный диалог открытия файла
// Возвращает путь к выбранному файлу или ошибку
// Если пользователь отменил выбор, возвращается dialog.Cancelled
func FileOpenDialog(title string, filter string, ext string) (string, error) {
	dlg := dialog.File().Title(title)
	
	if filter != "" && ext != "" {
		dlg = dlg.Filter(filter, ext)
	}
	
	filename, err := dlg.Load()
	if err != nil {
		return "", err
	}
	
	return filename, nil
}

// FileSaveDialog показывает нативный диалог сохранения файла
// Возвращает путь для сохранения или ошибку
// Если пользователь отменил выбор, возвращается dialog.Cancelled
func FileSaveDialog(title string, defaultName string, filter string, ext string) (string, error) {
	dlg := dialog.File().Title(title)
	
	if filter != "" && ext != "" {
		dlg = dlg.Filter(filter, ext)
	}
	
	// Устанавливаем имя файла по умолчанию, если указано
	if defaultName != "" {
		// Для dialog.Save() нужно установить начальный путь
		// Получаем домашнюю директорию и добавляем имя файла
		homeDir, err := dialog.Directory().Title("").Browse()
		if err == dialog.Cancelled {
			return "", dialog.Cancelled
		}
		if err == nil && homeDir != "" {
			defaultPath := filepath.Join(homeDir, defaultName)
			// Сохраняем с предложенным путём
			filename, err := dialog.File().
				Title(title).
				Filter(filter, ext).
				SetStartFile(defaultPath).
				Save()
			return filename, err
		}
	}
	
	filename, err := dlg.Save()
	if err != nil {
		return "", err
	}
	
	return filename, nil
}

// FileSaveDialogSimple упрощенная версия диалога сохранения
// без предварительного выбора директории
func FileSaveDialogSimple(title string, filter string, ext string) (string, error) {
	dlg := dialog.File().Title(title)
	
	if filter != "" && ext != "" {
		dlg = dlg.Filter(filter, ext)
	}
	
	filename, err := dlg.Save()
	if err != nil {
		return "", err
	}
	
	return filename, nil
}

// IsCancelled проверяет, является ли ошибка отменой диалога пользователем
func IsCancelled(err error) bool {
	return err == dialog.Cancelled
}
