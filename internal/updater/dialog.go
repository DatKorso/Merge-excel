package updater

import (
	"fmt"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowUpdateDialog показывает диалоговое окно с информацией об обновлении
func ShowUpdateDialog(window fyne.Window, info *ReleaseInfo) {
	if info == nil || !info.IsNewer {
		return
	}

	// Создаем содержимое диалога
	content := createUpdateContent(info)

	// Создаем кнопки
	downloadButton := widget.NewButton("Скачать обновление", func() {
		openURL(info.DownloadURL)
	})
	downloadButton.Importance = widget.HighImportance

	laterButton := widget.NewButton("Напомнить позже", func() {
		// Просто закрываем диалог
	})

	skipButton := widget.NewButton("Пропустить эту версию", func() {
		// TODO: В будущем можно добавить сохранение пропущенной версии
	})

	// Создаем кастомный диалог
	d := dialog.NewCustom(
		"🎉 Доступно обновление",
		"Закрыть",
		container.NewVBox(
			content,
			container.NewGridWithColumns(3,
				downloadButton,
				laterButton,
				skipButton,
			),
		),
		window,
	)

	d.Resize(fyne.NewSize(600, 400))
	d.Show()
}

// createUpdateContent создает содержимое диалога с информацией об обновлении
func createUpdateContent(info *ReleaseInfo) fyne.CanvasObject {
	// Заголовок с версией
	versionLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("Версия %s", info.Version),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Дата релиза
	dateLabel := widget.NewLabel(
		fmt.Sprintf("Дата релиза: %s", info.ReleaseDate.Format("02.01.2006")),
	)
	dateLabel.Alignment = fyne.TextAlignCenter

	// Описание изменений
	changelogLabel := widget.NewLabel("Что нового:")
	changelogLabel.TextStyle = fyne.TextStyle{Bold: true}

	changelog := info.Changelog
	if changelog == "" {
		changelog = "Описание изменений недоступно"
	}

	// Используем RichText для changelog с возможностью прокрутки
	changelogText := widget.NewRichTextFromMarkdown(changelog)
	changelogScroll := container.NewScroll(changelogText)
	changelogScroll.SetMinSize(fyne.NewSize(550, 200))

	// Собираем все вместе
	return container.NewVBox(
		versionLabel,
		dateLabel,
		widget.NewSeparator(),
		changelogLabel,
		changelogScroll,
	)
}

// openURL открывает URL в браузере по умолчанию
func openURL(urlStr string) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return
	}
	
	// Используем fyne для открытия URL
	_ = fyne.CurrentApp().OpenURL(parsedURL)
}
