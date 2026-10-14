document.addEventListener("DOMContentLoaded", function() {
    createApp();
})

function createApp() {
    const button = document.querySelector('#command_button');
    const commandField = document.querySelector('#message');

    callApi(
        '/api/v1/widget_content',
        'GET',
        null,
        function(xhr) {
            commandField.value = xhr.response.text;
        },
    )

    button.addEventListener('click', function() {
        const message = commandField.value.trim();
        callApi(
            '/api/v1/command/message',
            'POST',
            {text: message},
        )
    });
}
