package gui

import (
	"fmt"
	"path/filepath"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/DatKorso/Merge-excel/internal/core"
	apperrors "github.com/DatKorso/Merge-excel/internal/errors"
)

// BaseFileTab вкладка выбора и настройки базового файла
type BaseFileTab struct {
	app *App

	// UI элементы
	filePathLabel      *widget.Label
	selectFileBtn      *widget.Button
	sheetList          *widget.List
	profileNameEntry   *widget.Entry
	useOzonTemplateChk *widget.Check // Чекбокс для шаблона Ozon
	
	// Панель настройки листа
	configPanel       *fyne.Container
	sheetNameLabel    *widget.Label
	headerRowEntry    *widget.Entry
	previewBtn        *widget.Button
	headerPreviewText *widget.Label

	// Данные
	sheets        []core.SheetConfig
	selectedSheet int
	
	// Флаг для предотвращения ложных срабатываний чекбоксов
	updatingUI bool
}

// NewBaseFileTab создает новую вкладку базового файла
func NewBaseFileTab(app *App) *BaseFileTab {
	tab := &BaseFileTab{
		app:           app,
		sheets:        []core.SheetConfig{},
		selectedSheet: -1,
	}

	return tab
}

// Build создает UI вкладки
func (t *BaseFileTab) Build() fyne.CanvasObject {
	// Метка пути к файлу
	t.filePathLabel = widget.NewLabel("Файл не выбран")
	t.filePathLabel.Wrapping = fyne.TextWrapWord

	// Кнопка выбора файла
	t.selectFileBtn = widget.NewButton("Выбрать базовый файл...", func() {
		t.onSelectFile()
	})

	// Поле ввода имени профиля
	t.profileNameEntry = widget.NewEntry()
	t.profileNameEntry.SetPlaceHolder("Введите имя профиля")

	// Чекбокс для использования шаблона Ozon
	t.useOzonTemplateChk = widget.NewCheck("Использовать шаблон Ozon (листы: Шаблон, Озон.Видео, Озон.Видеообложка с заголовками на строке 4)", func(checked bool) {
		t.onOzonTemplateToggled(checked)
	})
	
	// Загружаем настройку из конфига
	if settings := t.app.GetSettings(); settings != nil {
		t.useOzonTemplateChk.Checked = settings.UseOzonTemplate
	}

	// Список листов
	t.sheetList = widget.NewList(
		func() int {
			return len(t.sheets)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewCheck("", nil),
				widget.NewLabel("Sheet Name"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= widget.ListItemID(len(t.sheets)) {
				return
			}
			
			sheet := t.sheets[id]
			box := obj.(*fyne.Container)
			check := box.Objects[0].(*widget.Check)
			label := box.Objects[1].(*widget.Label)

			// Устанавливаем флаг, что мы обновляем UI программно
			t.updatingUI = true
			
			// Важно: сначала удаляем старый обработчик
			check.OnChanged = nil
			check.Checked = sheet.Enabled
			// КРИТИЧНО: вызываем Refresh() для обновления визуального состояния
			check.Refresh()
			
			// Сбрасываем флаг перед установкой обработчика
			t.updatingUI = false
			
			// Устанавливаем новый обработчик с защитой
			check.OnChanged = func(enabled bool) {
				// Игнорируем, если идет программное обновление
				if t.updatingUI {
					return
				}
				if int(id) < len(t.sheets) {
					t.sheets[id].Enabled = enabled
					t.updateProfile()
					t.app.logger.Info("Sheet toggled", "sheet", t.sheets[id].SheetName, "enabled", enabled)
				}
			}

			label.SetText(sheet.SheetName)
		},
	)

	t.sheetList.OnSelected = func(id widget.ListItemID) {
		t.selectedSheet = int(id)
		t.updateConfigPanel()
	}

	// Панель настройки листа
	t.sheetNameLabel = widget.NewLabel("Не выбран")
	t.sheetNameLabel.TextStyle = fyne.TextStyle{Bold: true}
	
	t.headerRowEntry = widget.NewEntry()
	t.headerRowEntry.SetPlaceHolder("Номер строки")
	t.headerRowEntry.Disable() // Включается при выборе листа
	
	t.previewBtn = widget.NewButton("Предпросмотр", func() {
		t.onPreviewHeaders()
	})
	t.previewBtn.Disable() // Включается при выборе листа
	
	t.headerPreviewText = widget.NewLabel("Выберите лист слева для настройки")
	t.headerPreviewText.Wrapping = fyne.TextWrapWord
	
	applyBtn := widget.NewButton("Применить изменения", func() {
		t.onApplySheetConfig()
	})
	applyBtn.Importance = widget.HighImportance
	
	t.configPanel = container.NewVBox(
		widget.NewLabel("Настройка выбранного листа:"),
		widget.NewSeparator(),
		container.NewVBox(
			widget.NewLabel("Лист:"),
			t.sheetNameLabel,
		),
		widget.NewSeparator(),
		container.NewVBox(
			widget.NewLabel("Номер строки с заголовками:"),
			t.headerRowEntry,
			t.previewBtn,
		),
		widget.NewSeparator(),
		applyBtn,
	)

	// Добавляем увеличенный padding вокруг панели настроек (двойной)
	paddedConfigPanel := container.NewPadded(
		container.NewPadded(t.configPanel),
	)

	// HSplit для разделения списка и настроек
	splitContainer := container.NewHSplit(
		container.NewPadded(
			container.NewPadded(
				container.NewBorder(
					widget.NewLabel("Листы в файле:"), nil, nil, nil,
					t.sheetList,
				),
			),
		),
		container.NewScroll(paddedConfigPanel),
	)
	splitContainer.SetOffset(0.5) // 50/50

	// Основной контейнер - используем Border для растягивания splitContainer
	mainContainer := container.NewBorder(
		// Top - верхние элементы
		container.NewVBox(
			widget.NewLabel("Шаг 1: Выберите базовый Excel файл"),
			t.selectFileBtn,
			t.filePathLabel,
			widget.NewSeparator(),
			widget.NewLabel("Имя профиля:"),
			t.profileNameEntry,
			widget.NewSeparator(),
			t.useOzonTemplateChk, // Добавляем чекбокс шаблона
			widget.NewSeparator(),
			widget.NewLabel("Шаг 2: Настройте листы для объединения"),
		),
		nil, // Bottom
		nil, // Left
		nil, // Right
		// Center - растягивается на всё доступное пространство
		splitContainer,
	)

	return mainContainer
}

// onSelectFile обработчик выбора файла
func (t *BaseFileTab) onSelectFile() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			t.app.ShowError(err)
			return
		}
		if reader == nil {
			return // Пользователь отменил выбор
		}
		defer reader.Close()

		path := reader.URI().Path()
		
		// Проверяем расширение файла
		if filepath.Ext(path) != ".xlsx" {
			t.app.ShowError(apperrors.NewInvalidFormatError(path))
			return
		}

		t.filePathLabel.SetText(path)
		t.app.SetBaseFile(path)

		t.app.logger.Info("Base file selected", "path", path)
		
		// Автоматически анализируем файл
		t.analyzeFile(path)
	}, t.app.window)
}

// analyzeFile анализирует выбранный файл и загружает листы
func (t *BaseFileTab) analyzeFile(filePath string) {
	// Анализируем файл
	sheetNames, err := t.app.analyzer.GetSheetNames(filePath)
	if err != nil {
		t.app.ShowError(err)
		return
	}

	// Создаем конфигурации для каждого листа
	t.sheets = make([]core.SheetConfig, 0, len(sheetNames))
	for _, name := range sheetNames {
		t.sheets = append(t.sheets, core.SheetConfig{
			SheetName: name,
			Enabled:   false, // По умолчанию выключены
			HeaderRow: 1,     // По умолчанию первая строка
			Headers:   []string{},
		})
	}

	// Применяем шаблон Ozon, если он включен
	if t.useOzonTemplateChk.Checked {
		template := t.app.configManager.GetOzonTemplate()
		for i := range t.sheets {
			sheet := &t.sheets[i]
			if config, exists := template[sheet.SheetName]; exists {
				sheet.Enabled = config.Enabled
				sheet.HeaderRow = config.HeaderRow
				sheet.FilterValues = config.FilterValues
				
				// Для листа "Шаблон" автоматически определяем столбец фильтрации
				if sheet.SheetName == "Шаблон" && len(config.FilterValues) > 0 {
					columnIndex, err := t.app.analyzer.FindBrandColumnInFirstRows(filePath, sheet.SheetName, sheet.HeaderRow)
					if err != nil {
						t.app.logger.Warn("не удалось найти столбец бренда для фильтрации", "error", err, "sheet", sheet.SheetName)
					} else if columnIndex >= 0 {
						sheet.FilterColumn = columnIndex
						t.app.logger.Info("автоматически определен столбец фильтрации",
							"sheet", sheet.SheetName,
							"column_index", columnIndex,
							"filter_values", sheet.FilterValues)
					} else {
						t.app.logger.Warn("столбец 'Бренд в одежде и обуви*' не найден, фильтрация не будет применена", "sheet", sheet.SheetName)
						sheet.FilterColumn = -1
					}
				}
				
				t.app.logger.Debug("applied Ozon template on load", "sheet", sheet.SheetName, "enabled", sheet.Enabled, "header_row", sheet.HeaderRow)
			}
		}
	}

	// Устанавливаем флаг обновления UI и обновляем список
	t.updatingUI = true
	t.sheetList.Refresh()
	t.updatingUI = false

	// Создаем новый профиль
	profileName := t.profileNameEntry.Text
	if profileName == "" {
		profileName = "Новый профиль"
	}

	profile := core.NewProfile(profileName)
	profile.BaseFileName = filePath
	profile.Sheets = t.sheets

	t.app.UpdateProfile(profile)

	t.app.ShowInfo("Файл загружен", fmt.Sprintf("Найдено листов: %d", len(sheetNames)))
	t.app.logger.Info("File analyzed", "sheets_count", len(sheetNames))
}

// updateConfigPanel обновляет панель настройки для выбранного листа
func (t *BaseFileTab) updateConfigPanel() {
	if t.selectedSheet < 0 || t.selectedSheet >= len(t.sheets) {
		// Сбрасываем панель
		t.sheetNameLabel.SetText("Не выбран")
		t.headerRowEntry.SetText("")
		t.headerRowEntry.Disable()
		t.previewBtn.Disable()
		t.headerPreviewText.SetText("Выберите лист слева для настройки")
		return
	}

	sheet := &t.sheets[t.selectedSheet]
	
	// Обновляем UI
	t.sheetNameLabel.SetText(sheet.SheetName)
	t.headerRowEntry.SetText(strconv.Itoa(sheet.HeaderRow))
	t.headerRowEntry.Enable()
	t.previewBtn.Enable()
	
	if len(sheet.Headers) > 0 {
		t.headerPreviewText.SetText(t.formatHeaders(sheet.Headers))
	} else {
		t.headerPreviewText.SetText("Нажмите 'Предпросмотр' для загрузки заголовков")
	}
}

// onPreviewHeaders обработчик предпросмотра заголовков
func (t *BaseFileTab) onPreviewHeaders() {
	if t.selectedSheet < 0 || t.selectedSheet >= len(t.sheets) {
		return
	}

	headerRow, err := strconv.Atoi(t.headerRowEntry.Text)
	if err != nil || headerRow < 1 {
		t.app.ShowError(apperrors.NewInvalidHeaderRowError(headerRow))
		return
	}

	sheet := &t.sheets[t.selectedSheet]
	baseFile := t.app.GetBaseFile()

	// Читаем заголовки
	headers, err := t.app.analyzer.GetHeaders(baseFile, sheet.SheetName, headerRow)
	if err != nil {
		t.app.ShowError(err)
		return
	}

	sheet.Headers = headers
	t.headerPreviewText.SetText(t.formatHeaders(headers))
	
	t.app.ShowInfo(
		"Заголовки загружены",
		fmt.Sprintf("Найдено %d колонок в строке %d", len(headers), headerRow),
	)
	
	t.app.logger.Info("Headers previewed", "sheet", sheet.SheetName, "header_row", headerRow, "count", len(headers))
}

// onApplySheetConfig применяет настройки листа
func (t *BaseFileTab) onApplySheetConfig() {
	if t.selectedSheet < 0 || t.selectedSheet >= len(t.sheets) {
		return
	}

	headerRow, err := strconv.Atoi(t.headerRowEntry.Text)
	if err != nil || headerRow < 1 {
		t.app.ShowError(apperrors.NewInvalidHeaderRowError(headerRow))
		return
	}

	sheet := &t.sheets[t.selectedSheet]
	sheet.HeaderRow = headerRow
	
	// Автоматически включаем лист после применения настроек
	if !sheet.Enabled {
		sheet.Enabled = true
		t.app.logger.Info("Sheet auto-enabled after config", "sheet", sheet.SheetName)
	}
	
	t.updateProfile()
	
	// Сохраняем текущий выбранный элемент
	selectedID := t.selectedSheet
	
	// Принудительное обновление списка:
	// 1. Снимаем выделение
	t.sheetList.UnselectAll()
	// 2. Обновляем весь список с флагом
	t.updatingUI = true
	t.sheetList.Refresh()
	t.updatingUI = false
	// 3. Возвращаем выделение
	t.sheetList.Select(widget.ListItemID(selectedID))
	
	t.app.logger.Info("Sheet config updated", "sheet", sheet.SheetName, "header_row", headerRow, "enabled", sheet.Enabled)
}

// showSheetConfig показывает конфигурацию выбранного листа
// DEPRECATED: заменено на updateConfigPanel
func (t *BaseFileTab) showSheetConfig(id widget.ListItemID) {
	// Этот метод больше не используется, но оставлен для совместимости
	// Вся логика перенесена в updateConfigPanel
}

// previewHeaders предпросмотр заголовков
// DEPRECATED: заменено на onPreviewHeaders
func (t *BaseFileTab) previewHeaders(id widget.ListItemID, headerRowStr string) {
	// Этот метод больше не используется
}

// formatHeaders форматирует список заголовков для отображения
func (t *BaseFileTab) formatHeaders(headers []string) string {
	if len(headers) == 0 {
		return "Нет заголовков"
	}

	result := ""
	for i, h := range headers {
		if i > 0 {
			result += ", "
		}
		result += h
		if i >= 9 { // Показываем только первые 10
			result += "..."
			break
		}
	}
	return result
}

// updateProfile обновляет профиль в приложении
func (t *BaseFileTab) updateProfile() {
	if profile := t.app.GetProfile(); profile != nil {
		profile.Sheets = t.sheets
		if name := t.profileNameEntry.Text; name != "" {
			profile.ProfileName = name
		}
	}
}

// LoadProfile загружает профиль в UI
func (t *BaseFileTab) LoadProfile(profile *core.Profile) {
	t.filePathLabel.SetText(profile.BaseFileName)
	t.profileNameEntry.SetText(profile.ProfileName)
	t.app.SetBaseFile(profile.BaseFileName)

	t.sheets = profile.Sheets
	
	// Защищаем от ложных срабатываний при обновлении
	t.updatingUI = true
	t.sheetList.Refresh()
	t.updatingUI = false
	
	// Сбрасываем выбор и панель настроек
	t.selectedSheet = -1
	t.updateConfigPanel()

	t.app.logger.Info("Profile loaded into UI", "name", profile.ProfileName)
}

// onOzonTemplateToggled обработчик переключения шаблона Ozon
func (t *BaseFileTab) onOzonTemplateToggled(checked bool) {
	// Сохраняем настройку
	if settings := t.app.GetSettings(); settings != nil {
		settings.UseOzonTemplate = checked
		if err := t.app.configManager.SaveSettings(settings); err != nil {
			t.app.logger.Error("не удалось сохранить настройки", "error", err)
		}
	}
	
	t.app.logger.Info("Ozon template toggled", "enabled", checked)
	
	// Если есть загруженные листы, применяем/снимаем шаблон
	if len(t.sheets) > 0 {
		if checked {
			t.applyOzonTemplate()
		} else {
			t.clearOzonTemplate()
		}
	}
}

// applyOzonTemplate применяет шаблон Ozon к загруженным листам
func (t *BaseFileTab) applyOzonTemplate() {
	template := t.app.configManager.GetOzonTemplate()
	baseFile := t.app.GetBaseFile()
	
	for i := range t.sheets {
		sheet := &t.sheets[i]
		if config, exists := template[sheet.SheetName]; exists {
			sheet.Enabled = config.Enabled
			sheet.HeaderRow = config.HeaderRow
			sheet.FilterValues = config.FilterValues
			sheet.UseTemplateArticles = config.UseTemplateArticles
			
			// Для листа "Шаблон" автоматически определяем столбец фильтрации
			if sheet.SheetName == "Шаблон" && len(config.FilterValues) > 0 {
				columnIndex, err := t.app.analyzer.FindBrandColumnInFirstRows(baseFile, sheet.SheetName, sheet.HeaderRow)
				if err != nil {
					t.app.logger.Warn("не удалось найти столбец бренда для фильтрации", "error", err, "sheet", sheet.SheetName)
				} else if columnIndex >= 0 {
					sheet.FilterColumn = columnIndex
					t.app.logger.Info("автоматически определен столбец фильтрации",
						"sheet", sheet.SheetName,
						"column_index", columnIndex,
						"filter_values", sheet.FilterValues)
				} else {
					t.app.logger.Warn("столбец 'Бренд в одежде и обуви*' не найден, фильтрация не будет применена", "sheet", sheet.SheetName)
					sheet.FilterColumn = -1
				}
			}
			
			t.app.logger.Debug("applied Ozon template", "sheet", sheet.SheetName, "enabled", sheet.Enabled, "header_row", sheet.HeaderRow, "use_template_articles", sheet.UseTemplateArticles)
		} else {
			// Листы, не входящие в шаблон, отключаем
			sheet.Enabled = false
		}
	}
	
	// Обновляем UI
	t.updatingUI = true
	t.sheetList.Refresh()
	t.updatingUI = false
	t.updateConfigPanel()
	t.updateProfile()
	
	t.app.ShowInfo("Шаблон применен", "Применен шаблон Ozon для листов")
}

// clearOzonTemplate снимает настройки шаблона Ozon
func (t *BaseFileTab) clearOzonTemplate() {
	// Сбрасываем все листы в состояние по умолчанию
	for i := range t.sheets {
		t.sheets[i].Enabled = false
		t.sheets[i].HeaderRow = 1
	}
	
	// Обновляем UI
	t.updatingUI = true
	t.sheetList.Refresh()
	t.updatingUI = false
	t.updateConfigPanel()
	t.updateProfile()
	
	t.app.ShowInfo("Шаблон сброшен", "Настройки шаблона Ozon сброшены")
}

