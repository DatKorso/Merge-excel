package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DatKorso/Merge-excel/internal/core"
)

// Manager управляет профилями конфигурации
type Manager struct {
	configDir   string
	profilesDir string
	logger      *slog.Logger
}

// NewManager создает новый менеджер конфигураций
func NewManager(logger *slog.Logger) (*Manager, error) {
	if logger == nil {
		logger = slog.Default()
	}

	// Получаем домашнюю директорию пользователя
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить домашнюю директорию: %w", err)
	}

	// Директории приложения
	appDir := filepath.Join(homeDir, ".excel-merger")
	configDir := filepath.Join(appDir, "configs")
	profilesDir := filepath.Join(configDir, "profiles")

	// Создаем директории если их нет
	dirs := []string{appDir, configDir, profilesDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("не удалось создать директорию %s: %w", dir, err)
		}
	}

	logger.Info("менеджер конфигураций инициализирован",
		"config_dir", configDir,
		"profiles_dir", profilesDir,
	)

	return &Manager{
		configDir:   configDir,
		profilesDir: profilesDir,
		logger:      logger,
	}, nil
}

// SaveProfile сохраняет профиль в JSON файл
func (m *Manager) SaveProfile(profile *core.Profile, filename string) error {
	if profile == nil {
		return fmt.Errorf("профиль не может быть nil")
	}

	// Валидируем профиль
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("профиль невалиден: %w", err)
	}

	// Обновляем время изменения
	profile.UpdatedAt = time.Now()

	// Если время создания не установлено, устанавливаем
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = time.Now()
	}

	// Убираем расширение если оно есть
	filename = strings.TrimSuffix(filename, ".json")

	// Полный путь к файлу
	filePath := filepath.Join(m.profilesDir, filename+".json")

	// Сериализуем в JSON с отступами
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("не удалось сериализовать профиль: %w", err)
	}

	// Записываем в файл
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("не удалось записать файл профиля: %w", err)
	}

	m.logger.Info("профиль сохранен",
		"profile", profile.ProfileName,
		"file", filePath,
		"sheets_count", len(profile.Sheets),
	)

	return nil
}

// LoadProfile загружает профиль из JSON файла
func (m *Manager) LoadProfile(filename string) (*core.Profile, error) {
	// Убираем расширение если оно есть
	filename = strings.TrimSuffix(filename, ".json")

	// Полный путь к файлу
	filePath := filepath.Join(m.profilesDir, filename+".json")

	// Проверяем существование файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("файл профиля не найден: %s", filename)
	}

	// Читаем файл
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл профиля: %w", err)
	}

	// Десериализуем из JSON
	var profile core.Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("не удалось десериализовать профиль: %w", err)
	}

	// Валидируем загруженный профиль
	if err := profile.Validate(); err != nil {
		return nil, fmt.Errorf("загруженный профиль невалиден: %w", err)
	}

	m.logger.Info("профиль загружен",
		"profile", profile.ProfileName,
		"file", filePath,
		"sheets_count", len(profile.Sheets),
	)

	return &profile, nil
}

// ListProfiles возвращает список всех доступных профилей
func (m *Manager) ListProfiles() ([]ProfileInfo, error) {
	entries, err := os.ReadDir(m.profilesDir)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать директорию профилей: %w", err)
	}

	var profiles []ProfileInfo

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Проверяем расширение .json
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// Получаем информацию о файле
		info, err := entry.Info()
		if err != nil {
			m.logger.Warn("не удалось получить информацию о файле",
				"file", entry.Name(),
				"error", err,
			)
			continue
		}

		// Пытаемся загрузить профиль для получения деталей
		filename := strings.TrimSuffix(entry.Name(), ".json")
		profile, err := m.LoadProfile(filename)
		if err != nil {
			m.logger.Warn("не удалось загрузить профиль",
				"file", entry.Name(),
				"error", err,
			)
			// Добавляем базовую информацию
			profiles = append(profiles, ProfileInfo{
				Filename:  filename,
				Name:      filename,
				ModTime:   info.ModTime(),
				Size:      info.Size(),
				IsCorrupt: true,
			})
			continue
		}

		profiles = append(profiles, ProfileInfo{
			Filename:    filename,
			Name:        profile.ProfileName,
			BaseFile:    profile.BaseFileName,
			SheetsCount: len(profile.Sheets),
			CreatedAt:   profile.CreatedAt,
			UpdatedAt:   profile.UpdatedAt,
			ModTime:     info.ModTime(),
			Size:        info.Size(),
			IsCorrupt:   false,
		})
	}

	m.logger.Info("получен список профилей", "count", len(profiles))

	return profiles, nil
}

// DeleteProfile удаляет профиль
func (m *Manager) DeleteProfile(filename string) error {
	// Убираем расширение если оно есть
	filename = strings.TrimSuffix(filename, ".json")

	// Полный путь к файлу
	filePath := filepath.Join(m.profilesDir, filename+".json")

	// Проверяем существование файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("файл профиля не найден: %s", filename)
	}

	// Удаляем файл
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("не удалось удалить файл профиля: %w", err)
	}

	m.logger.Info("профиль удален", "file", filename)

	return nil
}

// ProfileExists проверяет существование профиля
func (m *Manager) ProfileExists(filename string) bool {
	filename = strings.TrimSuffix(filename, ".json")
	filePath := filepath.Join(m.profilesDir, filename+".json")
	_, err := os.Stat(filePath)
	return err == nil
}

// ExportProfile экспортирует профиль в указанную директорию
func (m *Manager) ExportProfile(filename, destPath string) error {
	// Убираем расширение если оно есть
	filename = strings.TrimSuffix(filename, ".json")

	// Полный путь к файлу источника
	srcPath := filepath.Join(m.profilesDir, filename+".json")

	// Проверяем существование файла
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("файл профиля не найден: %s", filename)
	}

	// Читаем файл
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл профиля: %w", err)
	}

	// Записываем в новое место
	destFile := filepath.Join(destPath, filename+".json")
	if err := os.WriteFile(destFile, data, 0644); err != nil {
		return fmt.Errorf("не удалось записать файл профиля: %w", err)
	}

	m.logger.Info("профиль экспортирован",
		"source", srcPath,
		"destination", destFile,
	)

	return nil
}

// ImportProfile импортирует профиль из указанного пути
func (m *Manager) ImportProfile(srcPath string) error {
	// Проверяем существование файла
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("файл профиля не найден: %s", srcPath)
	}

	// Читаем и валидируем профиль
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл профиля: %w", err)
	}

	var profile core.Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return fmt.Errorf("не удалось десериализовать профиль: %w", err)
	}

	if err := profile.Validate(); err != nil {
		return fmt.Errorf("импортируемый профиль невалиден: %w", err)
	}

	// Получаем имя файла
	filename := filepath.Base(srcPath)
	filename = strings.TrimSuffix(filename, ".json")

	// Сохраняем в директорию профилей
	if err := m.SaveProfile(&profile, filename); err != nil {
		return fmt.Errorf("не удалось сохранить импортированный профиль: %w", err)
	}

	m.logger.Info("профиль импортирован",
		"source", srcPath,
		"profile", profile.ProfileName,
	)

	return nil
}

// GetProfilesDir возвращает путь к директории профилей
func (m *Manager) GetProfilesDir() string {
	return m.profilesDir
}

// GetConfigDir возвращает путь к директории конфигурации
func (m *Manager) GetConfigDir() string {
	return m.configDir
}

// ProfileInfo информация о профиле
type ProfileInfo struct {
	Filename    string    // Имя файла (без расширения)
	Name        string    // Имя профиля
	BaseFile    string    // Базовый файл
	SheetsCount int       // Количество листов
	CreatedAt   time.Time // Дата создания
	UpdatedAt   time.Time // Дата обновления
	ModTime     time.Time // Дата модификации файла
	Size        int64     // Размер файла в байтах
	IsCorrupt   bool      // Файл поврежден
}
