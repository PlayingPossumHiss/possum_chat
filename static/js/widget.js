let app
let vote = [];

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
           vote: vote,
        },
    });
    setInterval(function() {
        let url = '/api/v1/widget_content';
        refreshMessages(url, function(xhr){
            vote = xhr.response.vote;
            app.vote = vote;
        });
    }, 250);
}
