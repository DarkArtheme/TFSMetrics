# go-marathon-team-3

## Road map

- [ ] Библиотека получения и анализа данных из новых и старых версий TFS
- [ ] CLI для библиотеки
- [ ] Exporter

## Инструкция по запуску CLI

1. Из корневой папки проекта запустить команду
> go install ./cmd/cli-metrics
2. Далее можно запускать cli просто введя:
> cli-metrics
3. Введите следующую команду, чтобы настроить параметры подключения:
> cli-metrics config --url "SOME_TFS_URL" --token "YOUR_PERSONAL_ACCESS_TOKEN" --cache "TRUE_IF_YOU_WANT_CACHE"
4. Введите следующую команду, чтобы вывести список всех проектов:
> cli-metrics list
5. Введите следующую команду, чтобы посмотреть список всех коммитов:
> cli-metrics log [ProjectName]

Если ProjectName не задан, то команда выведет коммиты для всех проектов!


Используйте флаг *--help* для получения помощи.