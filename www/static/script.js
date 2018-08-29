var linesContainer = document.getElementById("lines")
var runButton = document.getElementById("run")
var codeEditor;


window.onload = function () {
    let codeArea = document.getElementById("code")
    codeEditor = CodeMirror.fromTextArea(codeArea, {
        lineNumbers: true,
        autofocus: true,
        indentUnit: 4,
        mode: "setlx"
    })
}

function run() {
    let code = codeEditor.getValue()
    // check if code is empty
    if (!code)
        return

    let out = document.getElementById("output")
    out.innerHTML = "<span class='info'>" + "Waiting for remote server..." + "</span>"
    fetch("/run", {
        method: "POST",
        headers: {
            "Content-Type": "text/plain; charset=utf-8",
        },
        body: code
    }).then(async (response) => {
        if (response.ok)
            return response.json()
        let errorMsg = await response.text()
        throw errorMsg ? errorMsg : response.statusText
    }).then((json) => {
        out.innerHTML = ""
        for (var msg of json.Events) {
            out.innerHTML += `<span class="${msg.Kind}">${msg.Text}</span>`
            out.scrollTo(0, out.scrollHeight)
        }
        out.innerHTML += `<span class="info">Program exited.</span>`
        out.scrollTo(0, out.scrollHeight)
    }).catch((e) => {
        out.innerHTML = `<span class="stderr">${e}</span>`
    })
}