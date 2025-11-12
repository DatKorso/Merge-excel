# Руководство по сборке Excel Merger

## Быстрая сборка для разработки

```bash
make build-dev
./excel-merger
```

## Сборка релиза для текущей платформы

Создает оптимизированный бинарник в папке `build/`:

```bash
make build-release
```

Результат: `build/excel-merger`

## Кросс-компиляция для всех платформ

### Требования

1. **Fyne CLI** (обязательно):
```bash
go install fyne.io/tools/cmd/fyne@latest
```

2. **MinGW** (для Windows на macOS):
```bash
brew install mingw-w64
```

### Сборка для всех платформ

```bash
make package-all
```

Создаст в `build/`:
- `excel-merger.app` (macOS app bundle)
- `excel-merger-win-x64.exe` (Windows)
- `excel-merger-linux-amd64.tar.xz` (Linux) - требует сборку на Linux

### Сборка для отдельных платформ

**macOS:**
```bash
make package-macos
# Результат: build/excel-merger.app
```

**Windows (на macOS с MinGW):**
```bash
make package-windows
# Результат: build/excel-merger-win-x64.exe (~25 MB)
```

**Linux:**
```bash
make package-linux
# На macOS не работает напрямую - используйте Linux или Docker
```

## Решение проблем

### "Fyne CLI не установлен"

```bash
go install fyne.io/tools/cmd/fyne@latest
# Убедитесь, что ~/go/bin в PATH
export PATH=$PATH:~/go/bin
```

### "MinGW не установлен" (для Windows сборки)

```bash
brew install mingw-w64
```

### Сборка для Linux на macOS

Прямая кросс-компиляция для Linux на macOS не работает из-за CGO. Варианты:

**1. Использовать Docker (рекомендуется):**
```bash
# Установите Docker Desktop
# Затем используйте fyne-cross:
go install github.com/fyne-io/fyne-cross@latest
fyne-cross linux -arch=amd64 -icon assets/icon.png ./cmd/excel-merger
```

**2. Собрать на Linux машине:**
```bash
# На Ubuntu/Debian:
sudo apt-get install gcc libgl1-mesa-dev xorg-dev
go install fyne.io/tools/cmd/fyne@latest
make package-linux
```

**3. GitHub Actions (автоматически):**
Используйте CI/CD для сборки на нативных платформах.

## Команды Makefile

| Команда | Описание |
|---------|----------|
| `make build` | Обычная сборка |
| `make build-dev` | Быстрая сборка для разработки |
| `make build-release` | Релиз для текущей платформы → `build/` |
| `make package-macos` | .app для macOS (требует fyne CLI) |
| `make package-windows` | .exe для Windows (требует fyne CLI + MinGW) |
| `make package-linux` | Для Linux (требует Linux или Docker) |
| `make package-all` | Сборка macOS + Windows (+ Linux если возможно) |
| `make clean` | Очистить build/ и временные файлы |

## Размер бинарников

- macOS .app: ~30-40 MB
- Windows .exe: ~25 MB
- Linux: ~23 MB

Размер обусловлен встроенными GUI-компонентами Fyne.

## Ручная сборка с fyne package

### macOS
```bash
cd cmd/excel-merger
~/go/bin/fyne package --os darwin --icon ../../assets/icon.png --release
# Результат: excel-merger.app
```

### Windows (на macOS)
```bash
cd cmd/excel-merger
CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
  ~/go/bin/fyne package --os windows --icon ../../assets/icon.png --release --app-id com.github.excel-merger
# Результат: excel-merger.exe
```

### Linux (на Linux)
```bash
cd cmd/excel-merger
~/go/bin/fyne package --os linux --icon ../../assets/icon.png --release
# Результат: excel-merger.tar.xz
```
