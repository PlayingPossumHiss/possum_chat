let app
let urlParams = extractParams()
let messages = [];

document.addEventListener("DOMContentLoaded", function() {
    refreshMessages(createApp);
})

function refreshMessages(callback) {
    let xhr = new XMLHttpRequest();
    let url = '/api/v1/messages'
    if (urlParams.forLast != null) {
        url += '?for_last=' + urlParams.forLast
    }
    xhr.open('GET', url);
    xhr.responseType = 'json';
    xhr.send();

    xhr.onload = function() {
        if (xhr.status == 200) {
            messages = xhr.response.messages.reverse();
            callback();
        }
    }
}

function createApp() { 
    app = new Vue({
        el: '#app',
        data: {
           messages: messages,
           useScroll: urlParams.useScroll,
        },
    });
    setInterval(function() {
        refreshMessages(function(){
            app.messages = messages;
        });
    }, 50)
}

function extractParams() {
    let queryString = window.location.search;
    let urlParams = new URLSearchParams(queryString);
    return {
        forLast: urlParams.get('for_last'),
        useScroll: urlParams.get('use_scroll') == "true",
    }
}
