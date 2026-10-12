function callApi(url, method, body, callback) {
    let xhr = new XMLHttpRequest();
    xhr.open(method, url);
    xhr.responseType = 'json';
    xhr.send(JSON.stringify(body));

    xhr.onload = function() {
        if (xhr.status == 200 && callback) {
            callback(xhr);
        }
    }
}
