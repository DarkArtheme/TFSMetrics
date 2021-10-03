# TFSMetrics
Данный репозиторий содержит библиотеку, позволяющую выгружать информацию о коммитах из azure репозитория, cli для взаимодействия с библиотекой и экспортер, позволяющий смотреть на метрики в Prometheus и Graphana.

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
