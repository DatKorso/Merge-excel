package gui

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	
	"github.com/DatKorso/Merge-excel/internal/native"
)

// FileListTab вкладка со списком файлов для объединения
type FileListTab struct {
	app *App

	// UI элементы
	fileList      *widget.List
	addBtn        *widget.Button
	removeBtn     *widget.Button
	clearBtn      *widget.Button
	fileCountLabel *widget.Label

	// Данные
	files       []string
	selectedIdx int
}

// NewFileListTab создает новую вкладку списка файлов
func NewFileListTab(app *App) *FileListTab {
	tab := &FileListTab{
		app:         app,
		files:       []string{},
		selectedIdx: -1,
	}

	return tab
}

// Build создает UI вкладки
func (t *FileListTab) Build() fyne.CanvasObject {
	// Список файлов
	t.fileList = widget.NewList(
		func() int {
			return len(t.files)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Template"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			box := obj.(*fyne.Container)
			label := box.Objects[0].(*widget.Label)
			label.SetText(fmt.Sprintf("%d. %s", id+1, filepath.Base(t.files[id])))
		},
	)

	// Счетчик файлов
	t.fileCountLabel = widget.NewLabel("Файлов: 0")

	// Кнопка добавления файлов
	t.addBtn = widget.NewButton("Добавить файлы...", func() {
		t.onAddFiles()
	})

	// Кнопка удаления выбранного файла
	t.removeBtn = widget.NewButton("Удалить выбранный", func() {
		t.onRemoveSelected()
	})
	t.removeBtn.Disable()

	// Кнопка очистки списка
	t.clearBtn = widget.NewButton("Очистить список", func() {
		t.onClearList()
	})
	t.clearBtn.Disable()

	// Обработчик выбора в списке
	t.fileList.OnSelected = func(id widget.ListItemID) {
		t.selectedIdx = int(id)
		t.removeBtn.Enable()
	}

	t.fileList.OnUnselected = func(id widget.ListItemID) {
		t.selectedIdx = -1
		t.removeBtn.Disable()
	}

	// Панель с кнопками
	buttonsBox := container.NewVBox(
		t.addBtn,
		t.removeBtn,
		t.clearBtn,
		widget.NewSeparator(),
		t.fileCountLabel,
	)

	// Инструкция
	instructionLabel := widget.NewLabel(
		"Добавьте файлы Excel для объединения.\n" +
			"Файлы должны иметь ту же структуру, что и базовый файл.\n\n" +
			"Вы можете:\n" +
			"• Нажать 'Добавить файлы...'\n" +
			"• Перетащить файлы в это окно (Drag & Drop)",
	)
	instructionLabel.Wrapping = fyne.TextWrapWord

	// Создаем основной layout
	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Файлы для объединения:"),
			instructionLabel,
			widget.NewSeparator(),
		),
		buttonsBox,
		nil,
		nil,
		t.fileList,
	)

	return content
}

// onAddFiles обработчик добавления файлов через диалог
func (t *FileListTab) onAddFiles() {
	// Открываем нативный диалог выбора файла
	filename, err := native.FileOpenDialog(
		"Добавить Excel файл",
		"Excel файлы",
		"xlsx",
	)
	
	// Проверяем отмену пользователем
	if native.IsCancelled(err) {
		return
	}
	
	if err != nil {
		t.app.ShowError(err)
		return
	}

	t.addFile(filename)
}

// OnFilesDropped обработчик Drag & Drop (публичный метод для вызова из App)
func (t *FileListTab) OnFilesDropped(uris []fyne.URI) {
	fmt.Printf("onFilesDropped called with %d URIs\n", len(uris))
	
	for _, uri := range uris {
		path := uri.Path()
		fmt.Printf("Processing URI: %s (ext: %s)\n", path, filepath.Ext(path))
		
		if filepath.Ext(path) == ".xlsx" {
			t.addFile(path)
		} else {
			fmt.Printf("Skipping non-xlsx file: %s\n", path)
		}
	}
}


// addFile добавляет файл в список
func (t *FileListTab) addFile(path string) {
	// Проверяем расширение
	if filepath.Ext(path) != ".xlsx" {
		t.app.ShowError(fmt.Errorf("Неподдерживаемый формат файла. Только .xlsx файлы разрешены"))
		return
	}

	// Проверяем, не является ли это базовым файлом
	if path == t.app.GetBaseFile() {
		t.app.ShowError(fmt.Errorf("Нельзя добавить базовый файл в список для объединения"))
		return
	}

	// Проверяем, что файл еще не добавлен
	for _, f := range t.files {
		if f == path {
			t.app.ShowInfo("Файл уже добавлен", "Файл '"+filepath.Base(path)+"' уже есть в списке")
			return
		}
	}

	// Добавляем файл
	t.files = append(t.files, path)
	t.fileList.Refresh()
	t.updateFileCount()

	// Включаем кнопки
	if len(t.files) > 0 {
		t.clearBtn.Enable()
	}

	t.app.logger.Info("File added to merge list", "path", path, "total_files", len(t.files))
}

// onRemoveSelected обработчик удаления выбранного файла
func (t *FileListTab) onRemoveSelected() {
	if t.selectedIdx < 0 || t.selectedIdx >= len(t.files) {
		return
	}

	removedFile := t.files[t.selectedIdx]
	t.files = append(t.files[:t.selectedIdx], t.files[t.selectedIdx+1:]...)

	t.selectedIdx = -1
	t.fileList.UnselectAll()
	t.fileList.Refresh()
	t.updateFileCount()

	if len(t.files) == 0 {
		t.clearBtn.Disable()
	}

	t.app.logger.Info("File removed from merge list", "path", removedFile, "total_files", len(t.files))
}

// onClearList обработчик очистки списка
func (t *FileListTab) onClearList() {
	t.app.ShowConfirm(
		"Очистить список",
		fmt.Sprintf("Удалить все файлы (%d) из списка?", len(t.files)),
		func(confirm bool) {
			if confirm {
				t.files = []string{}
				t.fileList.UnselectAll()
				t.fileList.Refresh()
				t.updateFileCount()
				t.clearBtn.Disable()
				t.removeBtn.Disable()

				t.app.logger.Info("File list cleared")
			}
		},
	)
}

// updateFileCount обновляет счетчик файлов
func (t *FileListTab) updateFileCount() {
	t.fileCountLabel.SetText(fmt.Sprintf("Файлов: %d", len(t.files)))
}

// GetFiles возвращает список всех файлов
func (t *FileListTab) GetFiles() []string {
	return t.files
}

// DropZone область для приема перетаскиваемых файлов
type DropZone struct {
	widget.BaseWidget
	content fyne.CanvasObject
	onDrop  func([]fyne.URI)
}

// NewDropZone создает новую зону для Drag & Drop
func NewDropZone(content fyne.CanvasObject, onDrop func([]fyne.URI)) *DropZone {
	z := &DropZone{
		content: content,
		onDrop:  onDrop,
	}
	z.ExtendBaseWidget(z)
	return z
}

// CreateRenderer создает рендерер
func (z *DropZone) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(z.content)
}

// Drop обработчик Drop события
func (z *DropZone) Drop(items []fyne.URI) {
	fmt.Printf("DropZone: Drop event! %d items\n", len(items))
	if z.onDrop != nil {
		z.onDrop(items)
	}
}

// Dragged обработчик перетаскивания
func (z *DropZone) Dragged(ev *fyne.DragEvent) {
	fmt.Println("DropZone: Dragged")
}

// DragEnd обработчик окончания перетаскивания
func (z *DropZone) DragEnd() {
	fmt.Println("DropZone: DragEnd")
}



