let app
let urlParams = extractParams()
let messages = [];

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
           useScroll: urlParams.useScroll,
           errorCount: 0,
           warnCount: 0,
        },
    });
    setInterval(function() {
        let url = '/api/v1/messages';
        if (urlParams.forLast != null) {
            url += '?for_last=' + urlParams.forLast;
        }
        refreshMessages(url, function(xhr){
            messages = xhr.response.messages.reverse();
            app.messages = messages;
        });
    }, 50);
    if (urlParams.useScroll) {
        setInterval(function() {
            refreshMessages('/api/v1/logging_status', function(xhr){
                app.errorCount = xhr.response.error_count;
                app.warnCount = xhr.response.warn_count;
            });
        }, 1000);
    }
}

function extractParams() {
    let queryString = window.location.search;
    let urlParams = new URLSearchParams(queryString);
    return {
        forLast: urlParams.get('for_last'),
        useScroll: urlParams.get('use_scroll') == "true",
    }
}
