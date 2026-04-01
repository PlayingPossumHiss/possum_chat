var app

document.addEventListener("DOMContentLoaded", function() {
    refreshMessages(createApp);
})


let messages = [];

function refreshMessages(callback) {
    let queryString = window.location.search;
    let urlParams = new URLSearchParams(queryString);
    let for_last = urlParams.get('for_last')
    let xhr = new XMLHttpRequest();
    let url = '/api/v1/messages'
    if (for_last != null) {
        url += '?for_last=' + for_last
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
        },
    });
    setInterval(function() {
        refreshMessages(function(){
            app.messages = messages;
        });
    }, 50)
}
