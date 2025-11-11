package core

import "time"

// Profile представляет сохраненный профиль настроек
type Profile struct {
	Version      string          `json:"version"`
	ProfileName  string          `json:"profile_name"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	BaseFileName string          `json:"base_file_name"`
	Sheets       []SheetConfig   `json:"sheets"`
	Settings     ProfileSettings `json:"settings"`
}

// SheetConfig настройки для одного листа
type SheetConfig struct {
	SheetName string   `json:"sheet_name"`
	Enabled   bool     `json:"enabled"`
	HeaderRow int      `json:"header_row"` // 1-based index
	Headers   []string `json:"headers"`
}

// ProfileSettings дополнительные настройки профиля
type ProfileSettings struct {
	SkipEmptyRows bool `json:"skip_empty_rows"`
	ShowWarnings  bool `json:"show_warnings"`
	PreviewRows   int  `json:"preview_rows"`
}

// NewProfile создает новый профиль с настройками по умолчанию
func NewProfile(name string) *Profile {
	now := time.Now()
	return &Profile{
		Version:     "1.0",
		ProfileName: name,
		CreatedAt:   now,
		UpdatedAt:   now,
		Sheets:      []SheetConfig{},
		Settings: ProfileSettings{
			SkipEmptyRows: true,
			ShowWarnings:  true,
			PreviewRows:   100,
		},
	}
}

// AddSheet добавляет конфигурацию листа в профиль
func (p *Profile) AddSheet(config SheetConfig) {
	p.Sheets = append(p.Sheets, config)
	p.UpdatedAt = time.Now()
}

// GetSheetConfig возвращает конфигурацию листа по имени
func (p *Profile) GetSheetConfig(sheetName string) *SheetConfig {
	for i := range p.Sheets {
		if p.Sheets[i].SheetName == sheetName {
			return &p.Sheets[i]
		}
	}
	return nil
}

// UpdateSheet обновляет конфигурацию листа
func (p *Profile) UpdateSheet(sheetName string, config SheetConfig) bool {
	for i := range p.Sheets {
		if p.Sheets[i].SheetName == sheetName {
			p.Sheets[i] = config
			p.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// RemoveSheet удаляет конфигурацию листа из профиля
func (p *Profile) RemoveSheet(sheetName string) bool {
	for i := range p.Sheets {
		if p.Sheets[i].SheetName == sheetName {
			p.Sheets = append(p.Sheets[:i], p.Sheets[i+1:]...)
			p.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// GetEnabledSheets возвращает список включенных листов
func (p *Profile) GetEnabledSheets() []SheetConfig {
	enabled := []SheetConfig{}
	for _, sheet := range p.Sheets {
		if sheet.Enabled {
			enabled = append(enabled, sheet)
		}
	}
	return enabled
}

// Validate проверяет корректность профиля
func (p *Profile) Validate() error {
	if p.ProfileName == "" {
		return &AppError{
			Code:    "E009",
			Message: "Имя профиля не может быть пустым",
		}
	}

	if p.BaseFileName == "" {
		return &AppError{
			Code:    "E009",
			Message: "Базовый файл не указан",
		}
	}

	for i, sheet := range p.Sheets {
		if sheet.SheetName == "" {
			return &AppError{
				Code:    "E009",
				Message: "Имя листа не может быть пустым",
				Context: map[string]interface{}{"sheet_index": i},
			}
		}
		if sheet.HeaderRow < 1 {
			return &AppError{
				Code:    "E004",
				Message: "Номер строки заголовков должен быть больше 0",
				Context: map[string]interface{}{"sheet": sheet.SheetName, "header_row": sheet.HeaderRow},
			}
		}
	}

	return nil
}

// AppError ошибка приложения (временное определение, будет импортироваться из errors)
type AppError struct {
	Code    string
	Message string
	Context map[string]interface{}
}

func (e *AppError) Error() string {
	return e.Message
}
