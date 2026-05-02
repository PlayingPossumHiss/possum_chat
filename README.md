# Хотелось бы иметь мультичат под линукс

## Запуск приложения

При запуске для каждого источника, формируется строка с кнопкой запуска. В подписе кнопке указан текущий статус подключения

## Конфиг

При запуске в папке с приложением должен лежать файл config.json следующего вида

``` json
{
    "port": 8081,
    "view": {
        "css_style": "",
        "time_to_hide_message": "180s",
        "time_to_delete_message": "1h"
    },
    "loging": {
        "log_path": "",
        "level": "INFO"
    },
    "connections": [
        {
            "source": "twitch",
            "key": "playingpossumhiss"
        },
        {
            "source": "vk_play_live",
            "key": "playingpossum"
        },
        {
            "source": "youtube",
            "key": "PlayingPossumHiss"
        },
        {
            "source": "donation_alerts",
            "key": "am I stupid enough to left my token here? (yes, I am)"
        }
    ]
}
```

- port - порт на котором запускаемся
- loging - правила логирования
    - log_path - писать ли лог в файл (надо указать путь). Если пустая строка, то просто пишем в консоль
    - level - уровень логирования "DEBUG", "INFO", "WARN", "ERROR"
- view - описание отображения
    - css_style - кастомный стиль
    - time_to_hide_message - через сколько скрывать сообщения, если 0, то сообщения будут скрываться сразу, определяется параметром for_last виджета
    - time_to_delete_message - через сколько удалять сообщения, если 0, то сообщения живут вечно
- connections - к чему подключаемся
    - source - источник (один из: youtube, vk_play_live, twitch, donation_alerts)
    - key - имя канала
        - для donation_alerts токен из строки виджета (и постарайтесь его не палить, это все же креды)

## Виджет

http://127.0.0.1:8081/messages.html - виджет для OBS будет тут после запуска
http://127.0.0.1:8081/messages.html?for_last=1h - если хотим отображать все комментарии за последний час
http://127.0.0.1:8081/messages.html?for_last=1h&use_scroll=true - если в дополнение к этому хотим, чтобы список можно было скролить (удобно для просмотра на втором экране). Так же можно увидеть ошибки, если они были в логах
