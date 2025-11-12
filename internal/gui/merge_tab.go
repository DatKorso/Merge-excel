package gui

import (
	"fmt"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/DatKorso/Merge-excel/internal/core"
	apperrors "github.com/DatKorso/Merge-excel/internal/errors"
)

// MergeTab вкладка объединения файлов
type MergeTab struct {
	app *App

	// UI элементы
	startBtn      *widget.Button
	saveBtn       *widget.Button
	progressBar   *widget.ProgressBar
	statusLabel   *widget.Label
	detailsLabel  *widget.Label
	resultPreview *widget.Label

	// Состояние
	mergeResult   *core.MergeResult
	mergeInProgress bool
}

// NewMergeTab создает новую вкладку объединения
func NewMergeTab(app *App) *MergeTab {
	tab := &MergeTab{
		app:             app,
		mergeInProgress: false,
	}

	return tab
}

// Build создает UI вкладки
func (t *MergeTab) Build() fyne.CanvasObject {
	// Кнопка запуска объединения
	t.startBtn = widget.NewButton("Начать объединение", func() {
		t.onStartMerge()
	})
	t.startBtn.Importance = widget.HighImportance

	// Кнопка сохранения результата
	t.saveBtn = widget.NewButton("Сохранить результат...", func() {
		t.onSaveResult()
	})
	t.saveBtn.Disable()

	// Прогресс бар
	t.progressBar = widget.NewProgressBar()
	t.progressBar.Min = 0
	t.progressBar.Max = 1

	// Метка статуса
	t.statusLabel = widget.NewLabel("Готов к объединению")
	t.statusLabel.Wrapping = fyne.TextWrapWord

	// Детальная информация
	t.detailsLabel = widget.NewLabel("")
	t.detailsLabel.Wrapping = fyne.TextWrapWord

	// Предпросмотр результата
	t.resultPreview = widget.NewLabel("")
	t.resultPreview.Wrapping = fyne.TextWrapWord

	// Инструкция
	instructionLabel := widget.NewLabel(
		"Объединение файлов:\n\n" +
			"1. Убедитесь, что базовый файл выбран и проанализирован\n" +
			"2. Добавьте файлы для объединения во второй вкладке\n" +
			"3. Нажмите 'Начать объединение'\n" +
			"4. Дождитесь завершения процесса\n" +
			"5. Сохраните результат",
	)
	instructionLabel.Wrapping = fyne.TextWrapWord

	// Контейнер с кнопками
	buttonsBox := container.NewHBox(
		t.startBtn,
		t.saveBtn,
	)

	// Панель прогресса
	progressBox := container.NewVBox(
		widget.NewLabel("Прогресс:"),
		t.progressBar,
		t.statusLabel,
		widget.NewSeparator(),
		t.detailsLabel,
	)

	// Основной контейнер - используем Border для растягивания области результата
	mainContainer := container.NewBorder(
		// Top - все элементы управления
		container.NewVBox(
			instructionLabel,
			widget.NewSeparator(),
			buttonsBox,
			widget.NewSeparator(),
			progressBox,
			widget.NewSeparator(),
			widget.NewLabel("Результат:"),
		),
		nil, // Bottom
		nil, // Left
		nil, // Right
		// Center - растягивается на всё доступное пространство
		container.NewScroll(t.resultPreview),
	)

	return mainContainer
}

// onStartMerge обработчик начала объединения
func (t *MergeTab) onStartMerge() {
	if t.mergeInProgress {
		t.app.ShowInfo("Объединение в процессе", "Дождитесь завершения текущего объединения")
		return
	}

	// Проверяем готовность
	if err := t.validateReadiness(); err != nil {
		t.app.ShowError(err)
		return
	}

	// Получаем данные
	profile := t.app.GetProfile()
	files := t.app.fileListTab.GetFiles()

	// Показываем предупреждение для больших объемов
	if len(files) >= 5 {
		t.app.ShowConfirm(
			"Предупреждение",
			fmt.Sprintf(
				"Вы собираетесь объединить %d файлов.\n\n"+
					"⚠️ Объединение может занять продолжительное время.\n\n"+
					"При обработке больших файлов полоса прогресса может временно остановиться — "+
					"это нормально и происходит при чтении файлов. "+
					"Пожалуйста, дождитесь завершения операции.\n\n"+
					"Продолжить?",
				len(files),
			),
			func(confirmed bool) {
				if confirmed {
					t.startMergeProcess(profile, files)
				}
			},
		)
		return
	}

	// Для малого количества файлов запускаем сразу
	t.startMergeProcess(profile, files)
}

// startMergeProcess запускает процесс объединения
func (t *MergeTab) startMergeProcess(profile *core.Profile, files []string) {

	// Сброс состояния
	t.progressBar.SetValue(0)
	t.statusLabel.SetText("Начинаю объединение...")
	t.detailsLabel.SetText("")
	t.resultPreview.SetText("")
	t.startBtn.Disable()
	t.saveBtn.Disable()
	t.mergeInProgress = true

	// Создаем канал для обновления прогресса
	progressChan := make(chan core.ProgressUpdate, 10)
	doneChan := make(chan error, 1)

	// Настраиваем callback для merger
	t.app.merger.SetProgressCallback(func(current, total int, message string) {
		progressChan <- core.ProgressUpdate{
			Current: current,
			Total:   total,
			Message: message,
		}
	})

	// Запускаем объединение в горутине
	go func() {
		startTime := time.Now()

		// Создаем конфигурацию для объединения
		sheetConfigs := make(map[string]*core.SheetConfig)
		for i := range profile.Sheets {
			if profile.Sheets[i].Enabled {
				sheetConfigs[profile.Sheets[i].SheetName] = &profile.Sheets[i]
			}
		}

		// Получаем путь к базовому файлу
		baseFile := t.app.GetBaseFile()

		result, err := t.app.merger.MergeFiles(baseFile, files, sheetConfigs)
		
		doneChan <- err
		close(progressChan)

		if result != nil {
			result.Duration = time.Since(startTime)
			t.mergeResult = result
		}
	}()

	// Обновляем UI в главной горутине
	go func() {
		for update := range progressChan {
			// Копируем значения для замыкания
			currentUpdate := update
			fyne.Do(func() {
				if currentUpdate.Total > 0 {
					progress := float64(currentUpdate.Current) / float64(currentUpdate.Total)
					t.progressBar.SetValue(progress)
				}
				t.statusLabel.SetText(currentUpdate.Message)
				
				// Обновляем детали
				t.detailsLabel.SetText(fmt.Sprintf(
					"Обработано: %d из %d",
					currentUpdate.Current,
					currentUpdate.Total,
				))
			})
		}

		// Ждем завершения
		err := <-doneChan
		
		fyne.Do(func() {
			t.mergeInProgress = false
			t.startBtn.Enable()

			if err != nil {
				t.statusLabel.SetText("Ошибка при объединении")
				t.progressBar.SetValue(0)
				t.app.ShowError(err)
				t.app.logger.Error("Merge failed", "error", err)
				return
			}

			// Объединение успешно
			t.statusLabel.SetText("Объединение завершено успешно!")
			t.progressBar.SetValue(1)
			t.saveBtn.Enable()

			t.showMergeResult()

			t.app.logger.Info("Merge completed successfully",
				"duration_ms", t.mergeResult.Duration.Milliseconds(),
				"total_rows", t.mergeResult.TotalRows,
			)
		})
	}()
}

// validateReadiness проверяет готовность к объединению
func (t *MergeTab) validateReadiness() error {
	// Проверяем профиль
	profile := t.app.GetProfile()
	if profile == nil {
		return apperrors.NewConfigError("Профиль не создан. Выберите базовый файл и проанализируйте его")
	}

	// Проверяем базовый файл
	if t.app.GetBaseFile() == "" {
		return apperrors.NewConfigError("Базовый файл не выбран")
	}

	// Проверяем наличие включенных листов
	hasEnabledSheets := false
	for _, sheet := range profile.Sheets {
		if sheet.Enabled {
			hasEnabledSheets = true
			break
		}
	}
	if !hasEnabledSheets {
		return apperrors.NewConfigError("Нет включенных листов для объединения")
	}

	// Проверяем список файлов
	files := t.app.fileListTab.GetFiles()
	if len(files) == 0 {
		return apperrors.NewConfigError("Список файлов для объединения пуст")
	}

	return nil
}

// showMergeResult показывает результат объединения
func (t *MergeTab) showMergeResult() {
	if t.mergeResult == nil {
		return
	}

	result := fmt.Sprintf(
		"✅ Объединение выполнено успешно!\n\n"+
			"Обработано файлов: %d\n"+
			"Обработано листов: %d\n"+
			"Всего строк объединено: %d\n"+
			"Время выполнения: %s\n\n",
		t.mergeResult.ProcessedFiles,
		t.mergeResult.ProcessedSheets,
		t.mergeResult.TotalRows,
		t.mergeResult.Duration.Round(time.Millisecond),
	)

	// Добавляем детали по листам
	if len(t.mergeResult.SheetStats) > 0 {
		result += "Детали по листам:\n"
		for sheetName, stats := range t.mergeResult.SheetStats {
			result += fmt.Sprintf("  • %s: %d строк\n", sheetName, stats.RowsMerged)
		}
	}

	// Обновление UI должно происходить в UI-потоке
	// Но этот метод уже вызывается из fyne.Do(), поэтому просто обновляем
	t.resultPreview.SetText(result)
}

// onSaveResult обработчик сохранения результата
func (t *MergeTab) onSaveResult() {
	if t.mergeResult == nil || t.mergeResult.WorkbookData == nil {
		t.app.ShowError(apperrors.NewConfigError("Нет результата для сохранения"))
		return
	}

	// Устанавливаем имя файла по умолчанию
	defaultName := "merged.xlsx"

	// Создаем диалог сохранения
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			t.app.ShowError(err)
			return
		}
		if writer == nil {
			return // Пользователь отменил сохранение
		}
		defer writer.Close()

		savePath := writer.URI().Path()

		// Убеждаемся что путь имеет расширение .xlsx
		if filepath.Ext(savePath) != ".xlsx" {
			savePath += ".xlsx"
		}

		// Сохраняем объединенный файл
		if err := t.mergeResult.WorkbookData.Save(savePath); err != nil {
			t.app.ShowError(err)
			return
		}

		t.app.ShowInfo(
			"Файл сохранен",
			fmt.Sprintf("Результат успешно сохранен в:\n%s\n\nОбъединено строк: %d", 
				savePath, t.mergeResult.TotalRows),
		)

		t.app.logger.Info("Merge result saved", 
			"path", savePath,
			"total_rows", t.mergeResult.TotalRows,
			"processed_files", t.mergeResult.ProcessedFiles,
		)

	}, t.app.window)

	// Устанавливаем имя файла по умолчанию
	saveDialog.SetFileName(defaultName)
	
	saveDialog.Show()
}

// Reset сбрасывает состояние вкладки
func (t *MergeTab) Reset() {
	t.progressBar.SetValue(0)
	t.statusLabel.SetText("Готов к объединению")
	t.detailsLabel.SetText("")
	t.resultPreview.SetText("")
	t.mergeResult = nil
	t.saveBtn.Disable()
	t.startBtn.Enable()
	t.mergeInProgress = false
}
