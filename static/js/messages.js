const MESSAGE_KEEP_MS = 20 * 60 * 1000; // удаляем сообщения старше 20 минут
const RECONNECT_DELAY_MS = 3000;
const PING_INTERVAL_MS = 5000;

let app
let urlParams = extractParams()
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
        } else if (onError) {
            onError();
        }
    }
}

function createApp() {
    app = new Vue({
        el: '#app',
        data: {
            messages: [],
            errorCount: 0,
            warnCount: 0,
            online: online,
        },
    });
    initData();
    listenEvents();
    setInterval(pruneOldMessages, 1000);
}

function listenEvents() {
    let port = Number(window.location.port) + 1;
    let socket = new WebSocket("ws://" + location.hostname + ":" + port + "/chat_events");
    let pingTimer = null;

    socket.onopen = function() {
        pingTimer = setInterval(function() {
            if (socket.readyState === WebSocket.OPEN) {
                socket.send("ping");
            }
        }, PING_INTERVAL_MS);
    };

    socket.onmessage = function(rawEvent) {
        if (rawEvent.data === "pong") {
            return;
        }
        let event;
        try {
            event = JSON.parse(rawEvent.data);
        } catch (e) {
            return;
        }
        switch (event.event_type) {
        case "logging_status":
            app.errorCount = event.logging_status.error_count;
            app.warnCount = event.logging_status.warn_count;
            break;
        case "online":
            if (event.online) {
                applyOnline(event.online);
            }
            break;
        case "message":
            onWsMessage(event.message);
            break;
        }
    };

    socket.onclose = function() {
        if (pingTimer !== null) {
            clearInterval(pingTimer);
            pingTimer = null;
        }
        setTimeout(function() {
            listenEvents();
        }, RECONNECT_DELAY_MS);
    };
}

function initData() {
    let url = '/api/v1/messages';
    if (urlParams.forLast != null) {
        url += '?for_last=' + urlParams.forLast;
    }
    refreshMessages(url, function(xhr) {
        app.messages = xhr.response.messages.reverse();
        if (xhr.response.online) {
            applyOnline(xhr.response.online);
        }
    });
}

function applyOnline(onlineList) {
    let newOnline = {
        twitch: undefined,
        kick: undefined,
        youtube: undefined,
        vk_play_live: undefined,
    };
    onlineList.forEach(function(element) {
        switch (element.source) {
        case "twitch":
            newOnline.twitch = element.count;
            break;
        case "kick":
            newOnline.kick = element.count;
            break;
        case "youtube":
            newOnline.youtube = element.count;
            break;
        case "vk_play_live":
            newOnline.vk_play_live = element.count;
            break;
        }
    });
    app.online = newOnline;
}

// Пока идёт первичная (или повторная после реконнекта) загрузка снапшота,
// копим WS-сообщения, чтобы не потерять их и не задублировать.
function onWsMessage(message) {
    app.messages.unshift(message);
}

function pruneOldMessages() {
    let now = Date.now();
    app.messages = app.messages.filter(function(message) {
        let createdAt = Date.parse(message.created_at);
        if (isNaN(createdAt)) {
            return true;
        }
        return (now - createdAt) < MESSAGE_KEEP_MS;
    });
}

function extractParams() {
    let queryString = window.location.search;
    let urlParams = new URLSearchParams(queryString);
    return {
        forLast: urlParams.get('for_last'),
    }
}
