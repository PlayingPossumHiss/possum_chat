let app
let urlParams = extractParams()
let messages = [];
let online = {
    twitch: undefined,
    kick: undefined,
    youtube: undefined,
    vk_play_live: undefined,
};

document.addEventListener("DOMContentLoaded", function() {
    createApp();
})

function refreshMessages(url, callback) {
    let xhr = new XMLHttpRequest();
    xhr.open('GET', url);
    xhr.responseType = 'json';
    xhr.send();

    xhr.onload = function() {
        if (xhr.status == 200) {
            callback(xhr);
        }
    }
}

function createApp() { 
    app = new Vue({
        el: '#app',
        data: {
           messages: messages,
           errorCount: 0,
           warnCount: 0,
           online: online,
        },
    });
    initData();
    listenEvents();
}

function listenEvents() {
    let port = Number(window.location.port)+1
    let socket = new WebSocket("ws://"+location.hostname+":"+port+"/chat_events");
    socket.onopen = function() {
        setInterval(function() {
            socket.send("ping")
        }, 5000);
    };

    socket.onmessage = function(rawEvent) {
        if (rawEvent.data == "pong") {
            return;
        }
        let event = JSON.parse(rawEvent.data)
        switch (event.event_type) {
            case "logging_status":
                app.errorCount = event.logging_status.error_count;
                app.warnCount = event.logging_status.warn_count;
                break;
            case "online":
                if (!event.online) {
                    return;
                }
                event.online.forEach(element => {
                    switch (element.source) {
                    case "twitch":
                        newOnline.twitch = element.count
                        break;
                    case "kick":
                        newOnline.kick = element.count
                        break;
                    case "youtube":
                        newOnline.youtube = element.count
                        break;
                    case "vk_play_live":
                        newOnline.vk_play_live = element.count
                        break;
                    }
                });
                break;
            case "message":
                app.messages.unshift(event.message);
                break;
                // TODO: только их теперь чистить надо
        };
    };

    socket.onclose = function() {
        listenEvents();
    };
}

function initData() {
    let url = '/api/v1/messages';
    if (urlParams.forLast != null) {
        url += '?for_last=' + urlParams.forLast;
    }
    refreshMessages(url, function(xhr){
        messages = xhr.response.messages.reverse();
        app.messages = messages;
        newOnline = {
            twitch: undefined,
            kick: undefined,
            youtube: undefined,
            vk_play_live: undefined,
        };
        if (xhr.response.online) {
            xhr.response.online.forEach(element => {
                switch (element.source) {
                case "twitch":
                    newOnline.twitch = element.count
                    break;
                case "kick":
                    newOnline.kick = element.count
                    break;
                case "youtube":
                    newOnline.youtube = element.count
                    break;
                case "vk_play_live":
                    newOnline.vk_play_live = element.count
                    break;
                }
            });
        }
        app.online = newOnline;
    });
}

function extractParams() {
    let queryString = window.location.search;
    let urlParams = new URLSearchParams(queryString);
    return {
        forLast: urlParams.get('for_last'),
    }
}
