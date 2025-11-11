package gui

import (
	"fmt"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/DatKorso/Merge-excel/internal/config"
	"github.com/DatKorso/Merge-excel/internal/core"
	apperrors "github.com/DatKorso/Merge-excel/internal/errors"
	"github.com/DatKorso/Merge-excel/internal/excel"
)

// App главная структура приложения
type App struct {
	fyneApp       fyne.App
	window        fyne.Window
	logger        *slog.Logger
	configManager *config.Manager
	analyzer      *core.BaseAnalyzer
	merger        *core.Merger
	excelWriter   *excel.Writer

	// Вкладки
	baseFileTab *BaseFileTab
	fileListTab *FileListTab
	mergeTab    *MergeTab

	// Текущее состояние
	currentProfile *core.Profile
	baseFilePath   string
}

// NewApp создает новое приложение
func NewApp(logger *slog.Logger, cfgManager *config.Manager) *App {
	application := &App{
		fyneApp:       app.NewWithID("com.excel-merger.app"),
		logger:        logger,
		configManager: cfgManager,
		excelWriter:   excel.NewWriter(),
	}

	application.analyzer = core.NewBaseAnalyzer(nil, logger)
	application.merger = core.NewMerger(nil, logger)

	return application
}

// Run запускает приложение
func (a *App) Run() {
	a.window = a.fyneApp.NewWindow("Excel Merger - Объединение файлов Excel")
	a.window.Resize(fyne.NewSize(900, 700))

	// Создаем вкладки
	a.baseFileTab = NewBaseFileTab(a)
	a.fileListTab = NewFileListTab(a)
	a.mergeTab = NewMergeTab(a)

	// Создаем контейнер с вкладками
	tabs := container.NewAppTabs(
		container.NewTabItem("1. Базовый файл", a.baseFileTab.Build()),
		container.NewTabItem("2. Файлы для объединения", a.fileListTab.Build()),
		container.NewTabItem("3. Объединение", a.mergeTab.Build()),
	)

	// Устанавливаем активную вкладку
	tabs.SelectIndex(0)

	// Создаем меню
	mainMenu := a.createMainMenu()
	a.window.SetMainMenu(mainMenu)

	// Устанавливаем содержимое окна
	a.window.SetContent(tabs)

	// Настраиваем Drag & Drop для всего окна
	a.window.SetOnDropped(func(pos fyne.Position, items []fyne.URI) {
		fmt.Printf("Window Drop event! Position: %v, Items: %d\n", pos, len(items))
		
		// Проверяем, на какой вкладке мы находимся
		if tabs.CurrentTabIndex() == 1 { // Вкладка "Файлы для объединения"
			a.fileListTab.OnFilesDropped(items)
		}
	})

	// Обработчик закрытия
	a.window.SetCloseIntercept(func() {
		a.onClose()
	})

	a.logger.Info("Application window created")
	a.window.ShowAndRun()
}

// createMainMenu создает главное меню приложения
func (a *App) createMainMenu() *fyne.MainMenu {
	// Меню "Файл"
	fileMenu := fyne.NewMenu("Файл",
		fyne.NewMenuItem("Открыть профиль...", func() {
			a.onLoadProfile()
		}),
		fyne.NewMenuItem("Сохранить профиль...", func() {
			a.onSaveProfile()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Выход", func() {
			a.fyneApp.Quit()
		}),
	)

	// Меню "Помощь"
	helpMenu := fyne.NewMenu("Помощь",
		fyne.NewMenuItem("О программе", func() {
			a.showAboutDialog()
		}),
	)

	return fyne.NewMainMenu(fileMenu, helpMenu)
}

// ShowError показывает диалог с ошибкой пользователю
func (a *App) ShowError(err error) {
	var message string

	if appErr, ok := err.(*apperrors.AppError); ok {
		if msg, exists := apperrors.UserMessages[appErr.Code]; exists {
			message = msg
		} else {
			message = appErr.Message
		}

		// Логируем детали
		a.logger.Error("Application error",
			"code", appErr.Code,
			"message", appErr.Message,
			"context", appErr.Context,
			"error", appErr.Err,
		)
	} else {
		message = err.Error()
		a.logger.Error("Unknown error", "error", err)
	}

	dialog.ShowError(fmt.Errorf("%s", message), a.window)
}

// ShowInfo показывает информационное сообщение
func (a *App) ShowInfo(title, message string) {
	dialog.ShowInformation(title, message, a.window)
}

// ShowConfirm показывает диалог подтверждения
func (a *App) ShowConfirm(title, message string, callback func(bool)) {
	dialog.ShowConfirm(title, message, callback, a.window)
}

// onLoadProfile обработчик загрузки профиля
func (a *App) onLoadProfile() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			a.ShowError(err)
			return
		}
		if reader == nil {
			return // Пользователь отменил выбор
		}
		defer reader.Close()

		profile, err := a.configManager.LoadProfile(reader.URI().Path())
		if err != nil {
			a.ShowError(err)
			return
		}

		a.currentProfile = profile
		a.baseFileTab.LoadProfile(profile)
		a.ShowInfo("Профиль загружен", "Профиль '"+profile.ProfileName+"' успешно загружен")

		a.logger.Info("Profile loaded", "name", profile.ProfileName)
	}, a.window)
}

// onSaveProfile обработчик сохранения профиля
func (a *App) onSaveProfile() {
	if a.currentProfile == nil {
		a.ShowError(apperrors.NewConfigError("Нет профиля для сохранения"))
		return
	}

	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			a.ShowError(err)
			return
		}
		if writer == nil {
			return // Пользователь отменил сохранение
		}
		defer writer.Close()

		if err := a.configManager.SaveProfile(a.currentProfile, writer.URI().Path()); err != nil {
			a.ShowError(err)
			return
		}

		a.ShowInfo("Профиль сохранен", "Профиль '"+a.currentProfile.ProfileName+"' успешно сохранен")

		a.logger.Info("Profile saved", "name", a.currentProfile.ProfileName, "path", writer.URI().Path())
	}, a.window)
}

// showAboutDialog показывает диалог "О программе"
func (a *App) showAboutDialog() {
	about := widget.NewLabel(
		"Excel Merger v0.1.0-alpha\n\n" +
			"Приложение для объединения нескольких файлов Excel\n" +
			"с одинаковой структурой в один файл.\n\n" +
			"© 2025",
	)
	about.Wrapping = fyne.TextWrapWord
	about.Alignment = fyne.TextAlignCenter

	dialog.ShowCustom("О программе", "Закрыть", about, a.window)
}

// onClose обработчик закрытия приложения
func (a *App) onClose() {
	a.logger.Info("Application closing")
	a.window.Close()
}

// UpdateProfile обновляет текущий профиль
func (a *App) UpdateProfile(profile *core.Profile) {
	a.currentProfile = profile
}

// GetProfile возвращает текущий профиль
func (a *App) GetProfile() *core.Profile {
	return a.currentProfile
}

// SetBaseFile устанавливает путь к базовому файлу
func (a *App) SetBaseFile(path string) {
	a.baseFilePath = path
	if a.currentProfile != nil {
		a.currentProfile.BaseFileName = path
	}
}

// GetBaseFile возвращает путь к базовому файлу
func (a *App) GetBaseFile() string {
	return a.baseFilePath
}
