document.addEventListener("DOMContentLoaded", function() {
    createApp();
})

function createApp() {
    const button = document.querySelector('#command_button');
    const commandField = document.querySelector('#command');

    button.addEventListener('click', function() {
        const command = commandField.value.trim();
        if (command === '') {
            return;
        }

        const xhr = new XMLHttpRequest();
        xhr.open('POST', '/api/v1/command');
        xhr.setRequestHeader('Content-Type', 'application/json');
        xhr.onload = function() {
            if (xhr.status != 200) {
                alert(xhr.response);
            }
        }
        xhr.send(JSON.stringify({
            command: command,
        }));

        commandField.value = '';
    });
}
