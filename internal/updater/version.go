package updater

import (
	"fmt"
	"strconv"
	"strings"
)

// Version представляет семантическую версию
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string // alpha, beta, rc и т.д.
}

// ParseVersion парсит строку версии в структуру Version
// Поддерживаемые форматы: v0.1.0, 0.1.0, v0.1.0-alpha, 0.1.0-beta
func ParseVersion(versionStr string) (*Version, error) {
	// Убираем префикс 'v' если он есть
	versionStr = strings.TrimPrefix(versionStr, "v")

	// Разделяем на версию и prerelease
	parts := strings.SplitN(versionStr, "-", 2)
	versionPart := parts[0]
	prerelease := ""
	if len(parts) > 1 {
		prerelease = parts[1]
	}

	// Парсим основную версию (major.minor.patch)
	versionNumbers := strings.Split(versionPart, ".")
	if len(versionNumbers) < 2 || len(versionNumbers) > 3 {
		return nil, fmt.Errorf("неверный формат версии: %s", versionStr)
	}

	major, err := strconv.Atoi(versionNumbers[0])
	if err != nil {
		return nil, fmt.Errorf("неверный major номер: %w", err)
	}

	minor, err := strconv.Atoi(versionNumbers[1])
	if err != nil {
		return nil, fmt.Errorf("неверный minor номер: %w", err)
	}

	patch := 0
	if len(versionNumbers) == 3 {
		patch, err = strconv.Atoi(versionNumbers[2])
		if err != nil {
			return nil, fmt.Errorf("неверный patch номер: %w", err)
		}
	}

	return &Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: prerelease,
	}, nil
}

// IsNewer проверяет, является ли текущая версия новее чем другая
func (v *Version) IsNewer(other *Version) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	if v.Patch != other.Patch {
		return v.Patch > other.Patch
	}

	// Если номера версий одинаковые, сравниваем prerelease
	// Релиз без prerelease считается новее чем с prerelease
	if v.Prerelease == "" && other.Prerelease != "" {
		return true
	}
	if v.Prerelease != "" && other.Prerelease == "" {
		return false
	}

	// Если оба имеют prerelease, считаем версии равными
	return false
}

// String возвращает строковое представление версии
func (v *Version) String() string {
	version := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		version += "-" + v.Prerelease
	}
	return version
}

// CompareVersions сравнивает две версии и возвращает true если latest новее current
func CompareVersions(current, latest string) (bool, error) {
	currentVer, err := ParseVersion(current)
	if err != nil {
		return false, fmt.Errorf("ошибка парсинга текущей версии: %w", err)
	}

	latestVer, err := ParseVersion(latest)
	if err != nil {
		return false, fmt.Errorf("ошибка парсинга последней версии: %w", err)
	}

	return latestVer.IsNewer(currentVer), nil
}
