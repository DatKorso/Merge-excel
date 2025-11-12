package updater

import (
	"context"
	"fmt"
	"log/slog"
)

// UpdateChecker проверяет наличие обновлений
type UpdateChecker struct {
	currentVersion string
	githubClient   *GitHubClient
	logger         *slog.Logger
}

// NewUpdateChecker создает новый экземпляр UpdateChecker
func NewUpdateChecker(currentVersion, owner, repo string, logger *slog.Logger) *UpdateChecker {
	return &UpdateChecker{
		currentVersion: currentVersion,
		githubClient:   NewGitHubClient(owner, repo),
		logger:         logger,
	}
}

// CheckForUpdates проверяет наличие новой версии
// Возвращает информацию об обновлении если оно доступно, или nil если обновлений нет
func (uc *UpdateChecker) CheckForUpdates(ctx context.Context) (*ReleaseInfo, error) {
	uc.logger.Info("Проверка обновлений",
		"current_version", uc.currentVersion,
	)

	// Получаем последний релиз из GitHub
	release, err := uc.githubClient.GetLatestRelease(ctx)
	if err != nil {
		uc.logger.Warn("Не удалось получить информацию о последнем релизе",
			"error", err,
		)
		return nil, fmt.Errorf("ошибка получения информации о релизе: %w", err)
	}

	// Пропускаем черновики и pre-release версии
	if release.Draft || release.Prerelease {
		uc.logger.Info("Последний релиз является черновиком или pre-release, пропускаем",
			"tag", release.TagName,
			"draft", release.Draft,
			"prerelease", release.Prerelease,
		)
		return nil, nil
	}

	uc.logger.Info("Получен последний релиз",
		"version", release.TagName,
		"published_at", release.PublishedAt,
	)

	// Сравниваем версии
	isNewer, err := CompareVersions(uc.currentVersion, release.TagName)
	if err != nil {
		uc.logger.Error("Ошибка сравнения версий",
			"current", uc.currentVersion,
			"latest", release.TagName,
			"error", err,
		)
		return nil, fmt.Errorf("ошибка сравнения версий: %w", err)
	}

	if !isNewer {
		uc.logger.Info("Установлена последняя версия",
			"current_version", uc.currentVersion,
			"latest_version", release.TagName,
		)
		return nil, nil
	}

	uc.logger.Info("Доступно обновление",
		"current_version", uc.currentVersion,
		"new_version", release.TagName,
	)

	// Создаем информацию об обновлении
	info := release.ToReleaseInfo()
	info.IsNewer = true

	return info, nil
}
