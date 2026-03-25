var app

document.addEventListener("DOMContentLoaded", function() {
    refreshMessages(createApp);
})


let messages = [];

function refreshMessages(callback) {
    let xhr = new XMLHttpRequest();
    xhr.open('GET', '/api/v1/messages');
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
