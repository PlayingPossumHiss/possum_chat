document.addEventListener("DOMContentLoaded", function() {
    createApp();
})

function createApp() {
    const button = document.querySelector('#command_button');

    button.addEventListener('click', function() {
        let commandField = document.querySelector('#command');
        let xhr = new XMLHttpRequest();
        xhr.open('POST', '/api/v1/command');
        let request = JSON.stringify({
            command: commandField.value,
        });
        xhr.send(request);

        xhr.onload = function() {
            if (xhr.status != 200) {
                alert(xhr.response);
            }
        }

        commandField.value = '';
    });
}