# Go Changelogger

Отдельная Go-реализация `changelogger`, которую можно собрать в один бинарник и запускать без PHP и Composer на машине пользователя.

## Сборка

```bash
go build -o ./bin/changelogger ./cmd/changelogger
```

После сборки пользователю нужен только файл `bin/changelogger`.

## Установка

После публикации release-артефактов установка выполняется одной командой:

```bash
curl -fsSL https://github.com/hiddenpathz/changelogger/releases/latest/download/changelogger-install | sh
```

Installer сам определит ОС и поставит бинарник в `/usr/local/bin/changelogger`.
После этого `changelogger` можно запускать из корня любого проекта с `.env` и `CHANGELOG.md`.

Если нужен другой каталог установки:

```bash
curl -fsSL https://github.com/hiddenpathz/changelogger/releases/latest/download/changelogger-install | CHANGELOGGER_INSTALL_DIR="$HOME/bin" sh
```

## Release-артефакты

Собрать два бинарника для публикации:

```bash
./scripts/build-release.sh
```

Скрипт создаст:

```text
dist/changelogger-darwin-universal
dist/changelogger-linux-amd64
dist/changelogger-install
dist/checksums.txt
```

Локально проверить installer без GitHub Releases можно так:

```bash
./scripts/build-release.sh
CHANGELOGGER_BASE_URL="file://$PWD/dist" CHANGELOGGER_INSTALL_DIR="$HOME/bin" ./changelogger-install
```

## Публикация релиза без CI

1. Соберите артефакты локально:

   ```bash
   ./scripts/build-release.sh
   ```

2. Создайте GitHub Release в `https://github.com/hiddenpathz/changelogger`.

3. Приложите к release все файлы из `dist/`:

   ```text
   changelogger-darwin-universal
   changelogger-linux-amd64
   changelogger-install
   checksums.txt
   ```

4. После публикации проверьте установку:

   ```bash
   curl -fsSL https://github.com/hiddenpathz/changelogger/releases/latest/download/changelogger-install | sh
   changelogger
   ```

## Запуск

Запускать нужно из корня проекта, где лежит `.env`:

```bash
changelogger
changelogger https://gitlab.some.ru/your.repo.ru/-/tags/
```

Поддерживаемые переменные `.env`:

```env
REPOSITORY_LINK=https://gitlab.some.ru/your.repo.ru/-/tags/
CHANGELOG_PATH=./CHANGELOG.md
BRANCH_PREFIX=MYPROJECT
TASK_SYSTEM_NAME=SomeTaskSystemName
TASK_SYSTEM_LINK=https://some-task-system.ru/tasks/view?code=
```

## Что делает

1. Проверяет `.env`.
2. Получает последний git tag.
3. Просит номер заявки и задачи.
4. Создает ветку `feature/<BRANCH_PREFIX>-<Заявка-Задача>-assign-to-changelog`.
5. Просит выбрать уровень версии: `1` major, `2` minor, `3` fix.
6. Берет коммиты из `lastTag..develop`.
7. Группирует `feat`, `refactor`, `fix`, `remove` в регламентном порядке: `Реализовано`, `Изменено`, `Исправлено`, `Удалено`.
8. Поднимает первый символ описания после префикса коммита в uppercase.
9. Вставляет новый блок в `CHANGELOG.md`.
10. По подтверждению создает commit, push и удаляет локальную ветку.

## Проверка

```bash
go test ./...
```
