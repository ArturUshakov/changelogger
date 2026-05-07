# Go Changeloger

Отдельная Go-реализация `changeloger`, которую можно собрать в один бинарник и запускать без PHP и Composer на машине пользователя.

## Сборка

```bash
go build -o ./bin/changeloger ./cmd/changeloger
```

После сборки пользователю нужен только файл `bin/changeloger`.

## Установка

После публикации release-артефактов установка выполняется одной командой:

```bash
curl -fsSL https://github.com/hiddenpathz/changeloger/releases/latest/download/changeloger-install | sh
```

Installer сам определит ОС и поставит бинарник в `/usr/local/bin/changeloger`.
После этого `changeloger` можно запускать из корня любого проекта с `.env` и `CHANGELOG.md`.

Если нужен другой каталог установки:

```bash
curl -fsSL https://github.com/hiddenpathz/changeloger/releases/latest/download/changeloger-install | CHANGELOGER_INSTALL_DIR="$HOME/bin" sh
```

## Release-артефакты

Собрать два бинарника для публикации:

```bash
./scripts/build-release.sh
```

Скрипт создаст:

```text
dist/changeloger-darwin-universal
dist/changeloger-linux-amd64
dist/changeloger-install
dist/checksums.txt
```

Локально проверить installer без GitHub Releases можно так:

```bash
./scripts/build-release.sh
CHANGELOGER_BASE_URL="file://$PWD/dist" CHANGELOGER_INSTALL_DIR="$HOME/bin" ./changeloger-install
```

## Публикация релиза без CI

1. Соберите артефакты локально:

   ```bash
   ./scripts/build-release.sh
   ```

2. Создайте GitHub Release в `https://github.com/hiddenpathz/changeloger`.

3. Приложите к release все файлы из `dist/`:

   ```text
   changeloger-darwin-universal
   changeloger-linux-amd64
   changeloger-install
   checksums.txt
   ```

4. После публикации проверьте установку:

   ```bash
   curl -fsSL https://github.com/hiddenpathz/changeloger/releases/latest/download/changeloger-install | sh
   changeloger
   ```

## Запуск

Запускать нужно из корня проекта, где лежит `.env`:

```bash
changeloger
changeloger https://gitlab.some.ru/your.repo.ru/-/tags/
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
