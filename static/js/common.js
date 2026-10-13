function callApi(url, method, body, callback, onError) {
    let xhr = new XMLHttpRequest();
    xhr.open(method, url);
    xhr.responseType = 'json';

    xhr.onload = function() {
        if (xhr.status == 200) {
            if (callback) {
                callback(xhr);
            }
            return;
        }
        if (onError) {
            onError(xhr);
        } else {
            console.error('API ' + method + ' ' + url + ' failed: ' + xhr.status);
        }
    };

    if (body !== null && body !== undefined) {
        xhr.setRequestHeader('Content-Type', 'application/json');
        xhr.send(JSON.stringify(body));
        return;
    }
    xhr.send();
}
