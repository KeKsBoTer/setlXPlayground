var codeContainer = document.getElementById("code")
var linesContainer = document.getElementById("lines")
var runButton = document.getElementById("run")

function run() {
    var elm = codeContainer
    var out = document.getElementById("output")
    out.innerHTML = "<p class='info'>" + "Waiting for remote server..." + "</p>"
    fetch("/run", {
        method: "POST",
        headers: {
            "Content-Type": "text/plain; charset=utf-8",
        },
        body: elm.value
    }).then(function (response) {
        if (response.ok)
            return response.json()
        throw response.statusText;
    }).then(function (json) {
        var lines = ""
        for (var msg of json.Events) {
            lines += "<p>" + atob(msg.Text) + "</p>"
        }
        lines += "<p class='info'>" + "Program exited." + "</p>"
        out.innerHTML = lines;
    }).catch(function (e) {
        out.innerHTML = "<p class='error'>" + e + "</p>"
    })
}

document.onreadystatechange = function () {
    checkLines(codeContainer)
}

codeContainer.addEventListener('scroll', function (e) {
    linesContainer.style.transform = "translateY(" + -e.target.scrollTop + "px)"
}, {
    passive: true
});

codeContainer.onkeyup = function (e) {
    checkLines(e.target)
}
codeContainer.onkeydown = function (e) {
    checkLines(e.target)
}

function checkLines(area) {
    var lines = area.value.split("\n").length
    if (lines < 100)
        lines = 100
    var l = ""
    for (var i = 1; i <= lines; i++) {
        l += "<div>" + i + "</div>"
    }
    linesContainer.innerHTML = l
}