package core

import (
	"testing"
	"time"
)

func TestNewProfile(t *testing.T) {
	profileName := "Test Profile"
	profile := NewProfile(profileName)

	if profile.ProfileName != profileName {
		t.Errorf("Expected profile name %s, got %s", profileName, profile.ProfileName)
	}

	if profile.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", profile.Version)
	}

	if !profile.Settings.SkipEmptyRows {
		t.Error("Expected SkipEmptyRows to be true by default")
	}

	if profile.Settings.PreviewRows != 100 {
		t.Errorf("Expected PreviewRows to be 100, got %d", profile.Settings.PreviewRows)
	}
}

func TestAddSheet(t *testing.T) {
	profile := NewProfile("Test")

	sheet := SheetConfig{
		SheetName: "Продажи",
		Enabled:   true,
		HeaderRow: 5,
		Headers:   []string{"Дата", "Товар", "Сумма"},
	}

	profile.AddSheet(sheet)

	if len(profile.Sheets) != 1 {
		t.Errorf("Expected 1 sheet, got %d", len(profile.Sheets))
	}

	if profile.Sheets[0].SheetName != "Продажи" {
		t.Errorf("Expected sheet name Продажи, got %s", profile.Sheets[0].SheetName)
	}
}

func TestGetSheetConfig(t *testing.T) {
	profile := NewProfile("Test")

	sheet := SheetConfig{
		SheetName: "Продажи",
		Enabled:   true,
		HeaderRow: 5,
	}

	profile.AddSheet(sheet)

	// Проверка существующего листа
	found := profile.GetSheetConfig("Продажи")
	if found == nil {
		t.Error("Expected to find sheet Продажи")
	}

	// Проверка несуществующего листа
	notFound := profile.GetSheetConfig("Несуществующий")
	if notFound != nil {
		t.Error("Expected nil for non-existent sheet")
	}
}

func TestUpdateSheet(t *testing.T) {
	profile := NewProfile("Test")

	sheet := SheetConfig{
		SheetName: "Продажи",
		Enabled:   true,
		HeaderRow: 5,
	}

	profile.AddSheet(sheet)

	// Обновляем конфигурацию
	updatedSheet := SheetConfig{
		SheetName: "Продажи",
		Enabled:   false,
		HeaderRow: 3,
	}

	oldTime := profile.UpdatedAt
	time.Sleep(10 * time.Millisecond) // Небольшая задержка для изменения времени

	if !profile.UpdateSheet("Продажи", updatedSheet) {
		t.Error("Expected UpdateSheet to return true")
	}

	config := profile.GetSheetConfig("Продажи")
	if config.HeaderRow != 3 {
		t.Errorf("Expected HeaderRow 3, got %d", config.HeaderRow)
	}

	if config.Enabled {
		t.Error("Expected Enabled to be false")
	}

	if !profile.UpdatedAt.After(oldTime) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestRemoveSheet(t *testing.T) {
	profile := NewProfile("Test")

	sheet1 := SheetConfig{SheetName: "Лист1", Enabled: true, HeaderRow: 1}
	sheet2 := SheetConfig{SheetName: "Лист2", Enabled: true, HeaderRow: 1}

	profile.AddSheet(sheet1)
	profile.AddSheet(sheet2)

	if len(profile.Sheets) != 2 {
		t.Errorf("Expected 2 sheets, got %d", len(profile.Sheets))
	}

	// Удаляем первый лист
	if !profile.RemoveSheet("Лист1") {
		t.Error("Expected RemoveSheet to return true")
	}

	if len(profile.Sheets) != 1 {
		t.Errorf("Expected 1 sheet after removal, got %d", len(profile.Sheets))
	}

	if profile.Sheets[0].SheetName != "Лист2" {
		t.Errorf("Expected remaining sheet to be Лист2, got %s", profile.Sheets[0].SheetName)
	}
}

func TestGetEnabledSheets(t *testing.T) {
	profile := NewProfile("Test")

	sheet1 := SheetConfig{SheetName: "Enabled1", Enabled: true, HeaderRow: 1}
	sheet2 := SheetConfig{SheetName: "Disabled", Enabled: false, HeaderRow: 1}
	sheet3 := SheetConfig{SheetName: "Enabled2", Enabled: true, HeaderRow: 1}

	profile.AddSheet(sheet1)
	profile.AddSheet(sheet2)
	profile.AddSheet(sheet3)

	enabled := profile.GetEnabledSheets()

	if len(enabled) != 2 {
		t.Errorf("Expected 2 enabled sheets, got %d", len(enabled))
	}

	for _, sheet := range enabled {
		if !sheet.Enabled {
			t.Errorf("Expected only enabled sheets, got %s with Enabled=%v", sheet.SheetName, sheet.Enabled)
		}
	}
}

func TestValidate(t *testing.T) {
	// Валидный профиль
	profile := NewProfile("Valid Profile")
	profile.BaseFileName = "base.xlsx"
	profile.AddSheet(SheetConfig{
		SheetName: "Продажи",
		Enabled:   true,
		HeaderRow: 5,
	})

	if err := profile.Validate(); err != nil {
		t.Errorf("Expected valid profile to pass validation, got error: %v", err)
	}

	// Профиль без имени
	invalidProfile1 := NewProfile("")
	invalidProfile1.BaseFileName = "base.xlsx"
	if err := invalidProfile1.Validate(); err == nil {
		t.Error("Expected validation to fail for empty profile name")
	}

	// Профиль без базового файла
	invalidProfile2 := NewProfile("No Base File")
	if err := invalidProfile2.Validate(); err == nil {
		t.Error("Expected validation to fail for missing base file")
	}

	// Профиль с невалидным HeaderRow
	invalidProfile3 := NewProfile("Invalid HeaderRow")
	invalidProfile3.BaseFileName = "base.xlsx"
	invalidProfile3.AddSheet(SheetConfig{
		SheetName: "Лист1",
		Enabled:   true,
		HeaderRow: 0, // Невалидное значение
	})
	if err := invalidProfile3.Validate(); err == nil {
		t.Error("Expected validation to fail for HeaderRow < 1")
	}
}
