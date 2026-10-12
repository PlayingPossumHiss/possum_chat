document.addEventListener("DOMContentLoaded", function() {
    createApp();
})

function createApp() {
    let variants = []
    let is_started = false
    callApi("/api/v1/widget_content", "GET", null, function(xhr) {
        let vote = xhr.response.vote;
        is_started = vote.length != 0;
        variants.forEach(element => {
            variants.push({candidate:element.text, votes: element.count});
        })
    })

    app = new Vue({
        el: '#app',
        data: {
           variants: variants,
           is_started: is_started,
        },
        methods: {
            delete_variant: function(position) {
                this.variants.splice(position, 1);
            },
            add_variant: function() {
                this.variants.push({candidate:"", votes: 0})
            },
            start_election: function() {
                let request = {candidates:[]};
                this.variants.forEach(element => {
                    request.candidates.push(element.candidate)
                });
                callApi("/api/v1/command/start_election", "POST", request);
                refresh();
            },
            stop_election: function() {
                callApi("/api/v1/command/stop_election", "POST");
                refresh();
            },
        },
    });

    let refresh = function() {
        callApi("/api/v1/widget_content", "GET", null, function(xhr) {
            let vote = xhr.response.vote;
            app.is_started = vote.length != 0;
            if (!app.is_started) {
                return;
            }
            var new_variants = [];
            vote.forEach(element => {
                new_variants.push({candidate:element.text, votes: element.count});
            });
            app.variants = new_variants
        })
    }; 

    setInterval(refresh, 250);
}
