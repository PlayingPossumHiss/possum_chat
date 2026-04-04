# Хотелось бы иметь мультичат под линукс

## Конфиг

При запуске в корне должен лежать файл config.json

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
        }
    ]
}
```

- port - порт на котором запускаемся
- loging - правила логирования
    - log_path - писать ли лог в файл (надо указать путь). Если пустая строка, то просто пишем в консоль
    - level - уровень логирования "DEBUG", "INFO", "WARN", "ERROR"
- view - описание отображения
    - css_style - настомный стиль
    - time_to_hide_message - через сколько скрывать сообщения, если 0, то сообщения будут скрываться сразу, определяется параметром for_last виджета
    - time_to_delete_message - через сколько удалять сообщения, если 0, то сообщения живут вечно
- connections - к чему подключаемся
    - source - источник (один из: youtube, vk_play_live, twitch)
    - key - имя канала

## Виджет

http://127.0.0.1:8081/messages.html - виджет для OBS будет тут после запуска
http://127.0.0.1:8081/messages.html?for_last=1h - если хотим отображать все видео за последний час

## Известные проблемы:

- [Ютуб при старте подключается к последней трансляции, даже если она завершилась](https://github.com/PlayingPossumHiss/possum_chat/issues/25)
- [Если соединение с Твичем прервалось, то для реконекта надо перезапустить мультичат](https://github.com/PlayingPossumHiss/possum_chat/issues/26)
