# Хотелось бы иметь мультичат по линукс

## Конфиг

При запуске в корне должен лежать файл config.json

``` json
{
    "port": 8081,
    "view": {
        "css_style": "",
        "time_to_hide_message": "60s"
    },
    "connections": [
        {
            "source": "youtube",
            "key": "bMQejy5RoHM",
            "refresh_time": "50ms"
        }
    ]
}
```

- port - порт на котором запускаемся
- view - описание отображения
    - css_style - настомный стиль
    - time_to_hide_message - через сколько скрывать сообщения
- connections - к чему подключаемся
    - source - источник
    - key - ключ
    - refresh_time - как часто опрашивать

## Виджет

http://127.0.0.1:8081/messages.html - виджет для OBS будет тут после запуска

