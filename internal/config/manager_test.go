package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DatKorso/Merge-excel/internal/core"
)

func TestNewManager(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	manager, err := NewManager(logger)
	if err != nil {
		t.Fatalf("не удалось создать менеджер: %v", err)
	}

	if manager.profilesDir == "" {
		t.Error("profilesDir не должен быть пустым")
	}

	// Проверяем, что директории созданы
	if _, err := os.Stat(manager.profilesDir); os.IsNotExist(err) {
		t.Error("директория профилей не была создана")
	}

	t.Logf("Config dir: %s", manager.configDir)
	t.Logf("Profiles dir: %s", manager.profilesDir)
}

func TestSaveAndLoadProfile(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	manager, err := NewManager(logger)
	if err != nil {
		t.Fatalf("не удалось создать менеджер: %v", err)
	}

	// Создаем тестовый профиль
	profile := core.NewProfile("test_save_load")
	profile.BaseFileName = "base_file.xlsx"
	profile.Sheets = []core.SheetConfig{
		{
			SheetName: "Sheet1",
			Enabled:   true,
			HeaderRow: 5,
			Headers:   []string{"Column1", "Column2", "Column3"},
		},
		{
			SheetName: "Sheet2",
			Enabled:   false,
			HeaderRow: 1,
			Headers:   []string{},
		},
	}

	// Сохраняем профиль
	filename := "test_profile_save_load"
	err = manager.SaveProfile(profile, filename)
	if err != nil {
		t.Fatalf("не удалось сохранить профиль: %v", err)
	}

	// Загружаем профиль
	loadedProfile, err := manager.LoadProfile(filename)
	if err != nil {
		t.Fatalf("не удалось загрузить профиль: %v", err)
	}

	// Проверяем данные
	if loadedProfile.ProfileName != profile.ProfileName {
		t.Errorf("имя профиля не совпадает: ожидалось %s, получено %s",
			profile.ProfileName, loadedProfile.ProfileName)
	}

	if loadedProfile.BaseFileName != profile.BaseFileName {
		t.Errorf("базовый файл не совпадает: ожидалось %s, получено %s",
			profile.BaseFileName, loadedProfile.BaseFileName)
	}

	if len(loadedProfile.Sheets) != len(profile.Sheets) {
		t.Errorf("количество листов не совпадает: ожидалось %d, получено %d",
			len(profile.Sheets), len(loadedProfile.Sheets))
	}

	// Проверяем листы
	for i, sheet := range loadedProfile.Sheets {
		expected := profile.Sheets[i]
		if sheet.SheetName != expected.SheetName {
			t.Errorf("имя листа %d не совпадает: ожидалось %s, получено %s",
				i, expected.SheetName, sheet.SheetName)
		}
		if sheet.Enabled != expected.Enabled {
			t.Errorf("статус enabled листа %d не совпадает", i)
		}
		if sheet.HeaderRow != expected.HeaderRow {
			t.Errorf("headerRow листа %d не совпадает: ожидалось %d, получено %d",
				i, expected.HeaderRow, sheet.HeaderRow)
		}
	}

	// Очищаем после теста
	manager.DeleteProfile(filename)
}

func TestListProfiles(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	manager, err := NewManager(logger)
	if err != nil {
		t.Fatalf("не удалось создать менеджер: %v", err)
	}

	// Создаем несколько тестовых профилей
	profileNames := []string{"test_list_1", "test_list_2", "test_list_3"}
	
	for _, name := range profileNames {
		profile := core.NewProfile(name)
		profile.BaseFileName = "test.xlsx"
		profile.Sheets = []core.SheetConfig{
			{SheetName: "Sheet1", Enabled: true, HeaderRow: 1},
		}
		
		if err := manager.SaveProfile(profile, name); err != nil {
			t.Fatalf("не удалось сохранить профиль %s: %v", name, err)
		}
	}

	// Получаем список профилей
	profiles, err := manager.ListProfiles()
	if err != nil {
		t.Fatalf("не удалось получить список профилей: %v", err)
	}

	// Проверяем, что все наши профили в списке
	found := 0
	for _, profile := range profiles {
		for _, name := range profileNames {
			if profile.Filename == name {
				found++
				t.Logf("Найден профиль: %s (имя: %s, листов: %d)",
					profile.Filename, profile.Name, profile.SheetsCount)
			}
		}
	}

	if found < len(profileNames) {
		t.Errorf("ожидалось найти %d профилей, найдено %d", len(profileNames), found)
	}

	// Очищаем после теста
	for _, name := range profileNames {
		manager.DeleteProfile(name)
	}
}

func TestDeleteProfile(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	manager, err := NewManager(logger)
	if err != nil {
		t.Fatalf("не удалось создать менеджер: %v", err)
	}

	// Создаем профиль
	profile := core.NewProfile("test_delete")
	profile.BaseFileName = "test.xlsx"
	filename := "test_profile_delete"

	err = manager.SaveProfile(profile, filename)
	if err != nil {
		t.Fatalf("не удалось сохранить профиль: %v", err)
	}

	// Проверяем, что профиль существует
	if !manager.ProfileExists(filename) {
		t.Fatal("профиль должен существовать после сохранения")
	}

	// Удаляем профиль
	err = manager.DeleteProfile(filename)
	if err != nil {
		t.Fatalf("не удалось удалить профиль: %v", err)
	}

	// Проверяем, что профиль удален
	if manager.ProfileExists(filename) {
		t.Error("профиль не должен существовать после удаления")
	}

	// Попытка удалить несуществующий профиль
	err = manager.DeleteProfile(filename)
	if err == nil {
		t.Error("ожидалась ошибка при удалении несуществующего профиля")
	}
}

func TestProfileExists(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	manager, err := NewManager(logger)
	if err != nil {
		t.Fatalf("не удалось создать менеджер: %v", err)
	}

	filename := "test_profile_exists"

	// Проверяем несуществующий профиль
	if manager.ProfileExists(filename) {
		t.Error("несуществующий профиль не должен существовать")
	}

	// Создаем профиль
	profile := core.NewProfile("test_exists")
	profile.BaseFileName = "test.xlsx"

	err = manager.SaveProfile(profile, filename)
	if err != nil {
		t.Fatalf("не удалось сохранить профиль: %v", err)
	}

	// Проверяем существующий профиль
	if !manager.ProfileExists(filename) {
		t.Error("сохраненный профиль должен существовать")
	}

	// Проверяем с расширением .json
	if !manager.ProfileExists(filename + ".json") {
		t.Error("должен находить профиль даже с расширением .json")
	}

	// Очищаем
	manager.DeleteProfile(filename)
}

func TestExportImportProfile(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	manager, err := NewManager(logger)
	if err != nil {
		t.Fatalf("не удалось создать менеджер: %v", err)
	}

	// Создаем временную директорию для экспорта
	tempDir := t.TempDir()

	// Создаем профиль
	profile := core.NewProfile("test_export_import")
	profile.BaseFileName = "export_test.xlsx"
	profile.Sheets = []core.SheetConfig{
		{SheetName: "Sheet1", Enabled: true, HeaderRow: 5, Headers: []string{"A", "B", "C"}},
	}

	filename := "test_profile_export"
	err = manager.SaveProfile(profile, filename)
	if err != nil {
		t.Fatalf("не удалось сохранить профиль: %v", err)
	}

	// Экспортируем профиль
	err = manager.ExportProfile(filename, tempDir)
	if err != nil {
		t.Fatalf("не удалось экспортировать профиль: %v", err)
	}

	// Проверяем, что файл экспортирован
	exportedFile := filepath.Join(tempDir, filename+".json")
	if _, err := os.Stat(exportedFile); os.IsNotExist(err) {
		t.Fatal("экспортированный файл не найден")
	}

	// Удаляем оригинальный профиль
	manager.DeleteProfile(filename)

	// Импортируем обратно
	err = manager.ImportProfile(exportedFile)
	if err != nil {
		t.Fatalf("не удалось импортировать профиль: %v", err)
	}

	// Проверяем, что профиль импортирован
	if !manager.ProfileExists(filename) {
		t.Fatal("импортированный профиль не найден")
	}

	// Загружаем и проверяем данные
	importedProfile, err := manager.LoadProfile(filename)
	if err != nil {
		t.Fatalf("не удалось загрузить импортированный профиль: %v", err)
	}

	if importedProfile.ProfileName != profile.ProfileName {
		t.Errorf("имя импортированного профиля не совпадает")
	}

	// Очищаем
	manager.DeleteProfile(filename)
}

func TestSaveProfileValidation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	manager, err := NewManager(logger)
	if err != nil {
		t.Fatalf("не удалось создать менеджер: %v", err)
	}

	t.Run("nil профиль", func(t *testing.T) {
		err := manager.SaveProfile(nil, "test")
		if err == nil {
			t.Error("ожидалась ошибка для nil профиля")
		}
	})

	t.Run("невалидный профиль", func(t *testing.T) {
		profile := &core.Profile{
			ProfileName: "", // пустое имя - невалидно
			Version:     "1.0",
			CreatedAt:   time.Now(),
		}

		err := manager.SaveProfile(profile, "test")
		if err == nil {
			t.Error("ожидалась ошибка для невалидного профиля")
		}
	})
}

func TestLoadProfileErrors(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	manager, err := NewManager(logger)
	if err != nil {
		t.Fatalf("не удалось создать менеджер: %v", err)
	}

	t.Run("несуществующий профиль", func(t *testing.T) {
		_, err := manager.LoadProfile("несуществующий_профиль")
		if err == nil {
			t.Error("ожидалась ошибка для несуществующего профиля")
		}
	})
}
