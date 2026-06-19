# Changelogger

Changelogger - это консольный инструмент, который автоматически генерирует блок для `CHANGELOG.md` по git-коммитам.

Go-версия поставляется как один бинарник и не требует PHP, Composer или Packagist на машине пользователя.

## Требования к окружению

Для работы Changelogger нужны:

- Git;
- репозиторий с ветками `develop` и `origin/master`;
- git tags в формате `X.Y.Z`, например `1.24.5`;
- глобальный config пользователя;
- файл `CHANGELOG.md` в корне проекта или путь к нему в global config / `.env`.

## Установка

Для установки или обновления выполните:

```bash
curl -fsSL https://github.com/ArturUshakov/changelogger/releases/latest/download/changelogger-install | sh
```

Installer сам определит ОС и установит бинарник в:

```text
/usr/local/bin/changelogger
```

Если нужен другой каталог установки:

```bash
curl -fsSL https://github.com/ArturUshakov/changelogger/releases/latest/download/changelogger-install | CHANGELOGGER_INSTALL_DIR="$HOME/bin" sh
```

После установки команда доступна из любого проекта:

```bash
changelogger
```

Для обновления до последнего GitHub Release:

```bash
changelogger update
```

## Глобальная настройка

Создайте файл:

```text
~/.config/changelogger/config.json
```

Минимальный config:

```json
{
  "taskSystemLink": "https://helpdesk.efko.ru/tasks/view?code={code}",
  "changelogPath": "./CHANGELOG.md"
}
```

`taskSystemLink` можно указывать в одном из форматов:

```text
https://helpdesk.efko.ru/tasks/view?code=
https://helpdesk.efko.ru/tasks/view?code={code}
https://helpdesk.efko.ru
```

Если global config отсутствует, Changelogger создаст его с этим содержимым автоматически.

Настройки поведения можно менять без ручного редактирования файла:

```bash
changelogger config
```

Посмотреть текущие настройки:

```bash
changelogger config show
```

Настроить ссылки на заявки:

```bash
changelogger config task-link ask
changelogger config task-link always
changelogger config task-link never
```

Настроить создание commit:

```bash
changelogger config commit ask
changelogger config commit always
changelogger config commit never
```

Настроить способ записи changelog:

```bash
changelogger config changelog ask
changelogger config changelog current
changelogger config changelog branch
```

Сбросить сохраненный выбор:

```bash
changelogger config reset
```

## Подготовка проекта для работы

Локальный `.env` не обязателен. Changelogger автоматически определяет:

- ссылку на теги из `git remote get-url origin`;
- префикс ветки из коммитов, например `HR` из `[HR-IU291518-W1026701]`.

Если префикса системы нет, например `[IU291518-W1026701]`, ветка будет создана без него.
Если нужно переопределить автоматическое определение, добавьте `.env`:

```env
REPOSITORY_LINK=https://gitlab.some.ru/your.repo.ru/-/tags/
BRANCH_PREFIX=MYPROJECT
```

Если `.env` нет, Changelogger создаст шаблон:

```env
# Changelogger project config
# Optional overrides. Changelogger detects repository link and branch prefix automatically.
# REPOSITORY_LINK=
# BRANCH_PREFIX=
# CHANGELOG_PATH=./CHANGELOG.md
# TASK_SYSTEM_LINK=https://helpdesk.efko.ru/tasks/view?code={code}
```

Любую строку можно раскомментировать, если проекту нужен override global config или auto-detect.

## Запуск

Запускать нужно из каталога проекта:

```bash
changelogger
```

Можно явно передать ссылку на репозиторий до номера тега. В этом случае аргумент переопределит auto-detect и `REPOSITORY_LINK` из `.env`:

```bash
changelogger https://gitlab.some.ru/your.repo.ru/-/tags/
```

## Использование

При старте Changelogger собирает коммиты, показывает изменения, которые попадут в `CHANGELOG.md`, и предлагает версию:

```bash
Изменения которые попадут в CHANGELOG.md:
- Реализовано:
  - Добавлен экспорт [Заявка](https://helpdesk.efko.ru/tasks/view?code=IU212256)

Текущая версия приложения: 1.24.5
Рекомендация: minor (есть feat)
Какую версию нужно поднять?
 1 - major (*.0.0)
 2 - minor (0.*.0)
 3 - fix   (0.0.*)
 Enter - minor
>
```

Рекомендация выбирается по изменениям:

- если есть `feat`, предлагается `minor`;
- если `feat` нет, предлагается `fix`.

Enter принимает рекомендацию. `1`, `2`, `3`, `major`, `minor`, `fix` можно ввести вручную.

Затем Changelogger попросит ввести номер заявки и задачи в формате `IU000000-W0111111`.
Если нажать Enter, будет использована заявка-задача из последнего подходящего коммита.
Перед записью можно выбрать, как менять changelog:

- записать `CHANGELOG.md` в текущей ветке;
- перейти на `develop` и создать новую ветку.

Если выбран второй вариант, будет создана ветка:

```text
feature/MYPROJECT-IU000000-W0111111-assign-to-changelog
```

Если префикс не найден:

```text
feature/IU000000-W0111111-assign-to-changelog
```

После создания ветки Changelogger попросит подтверждение записи в `CHANGELOG.md`.
Если несколько коммитов дают полностью одинаковую строку changelog, строка будет добавлена один раз.

## Правила коммитов

В changelog попадают коммиты с префиксами:

| Ключ     | Запись      |
|----------|-------------|
| feat     | Реализовано |
| refactor | Изменено    |
| change   | Изменено    |
| fix      | Исправлено  |
| remove   | Удалено     |

Секции всегда выводятся в регламентном порядке:

```text
Реализовано
Изменено
Исправлено
Удалено
```

Описание после префикса поднимается с первой буквы в uppercase:

```bash
[IDPSPS-IU212256-W0835055] fix: исправлены опечатки
```

В `CHANGELOG.md` попадет:

```md
- Исправлены опечатки [Заявка](https://helpdesk.efko.ru/tasks/view?code=IU212256)
```

Код заявки автоматически берется из префикса коммита. Например, из `[HR-IU291518-W1026701]` будет взят `IU291518`. Буквы заявки не привязаны к `IU`: из `[HR-AB291518-W1026701]` будет взят `AB291518`.

Префиксы `refactor` и `change` равнозначны и оба попадают в секцию `Изменено`:

```bash
[IDPSPS-IU212256-W0835055] change: обновлен расчет скидок
```

В `CHANGELOG.md` попадет:

```md
- Изменено:
  - Обновлен расчет скидок [Заявка](https://helpdesk.efko.ru/tasks/view?code=IU212256)
```

Коммиты с другими префиксами не попадут в `CHANGELOG.md`, например:

```md
wip - промежуточные коммиты в процессе работы
ci - настройки CI/CD
build - настройки окружения без пользовательского смысла
```

## Пример результата

```md
# История изменений

## [ [1.1.0](https://gitlab.some.ru/your.repo.ru/-/tags/1.1.0) ] - 01.01.2023

- Реализовано:
  - Отображение кода заявки [Заявка](https://helpdesk.efko.ru/tasks/view?code=IU212256)
  - Учет параметра "Срочная доставка"
- Изменено:
  - Дополнены правила расчета стоимости [Заявка](https://helpdesk.efko.ru/tasks/view?code=IU212300)
- Исправлено:
  - Некорректные названия полей в форме [Заявка](https://helpdesk.efko.ru/tasks/view?code=IU212301)
- Удалено:
  - Удален устаревший экран настроек [Заявка](https://helpdesk.efko.ru/tasks/view?code=IU212302)
```

## Завершение работы

После записи `CHANGELOG.md` программа предложит:

- создать commit;
- push ветки;
- удалить локальную ветку после push.

Если на вопрос ответить `n`, действие будет отменено. Изменения можно будет завершить вручную.

## Сборка из исходников

Для локальной сборки нужен Go:

```bash
go build -o ./bin/changelogger ./cmd/changelogger
```

Проверка:

```bash
go test ./...
```

## Публикация релиза

Релиз публикуется автоматически через GitHub Actions после push тэга:

```bash
git tag 1.2.3
git push origin 1.2.3
```

Workflow создаст GitHub Release и приложит файлы:

```text
changelogger-darwin-universal
changelogger-linux-amd64
changelogger-install
checksums.txt
```

Локально собрать release-артефакты можно так:

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

После публикации проверьте установку:

```bash
curl -fsSL https://github.com/ArturUshakov/changelogger/releases/latest/download/changelogger-install | sh
changelogger
```

Локально проверить installer без GitHub Releases можно так:

```bash
./scripts/build-release.sh
CHANGELOGGER_BASE_URL="file://$PWD/dist" CHANGELOGGER_INSTALL_DIR="$HOME/bin" ./changelogger-install
```
