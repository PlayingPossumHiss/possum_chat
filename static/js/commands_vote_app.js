document.addEventListener("DOMContentLoaded", function() {
    createApp();
})

function createApp() {
    let variants = []
    app = new Vue({
        el: '#app',
        data: {
           variants: variants,
           is_started: false,
        },
        methods: {
            delete_variant: function(position) {
                this.variants.splice(position, 1);
            },
            add_variant: function() {
                this.variants.push({candidate:"", votes: 0})
            },
            start_election: function() {},
            stop_election: function() {},
        },
    });
}
