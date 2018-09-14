var linesContainer = document.getElementById("lines")
var runButton = document.getElementById("run")
var snippetUrl = document.getElementById("snippetUrl")
var codeEditor;
var executing = false;


window.onload = function () {
    if (isSharedCodePage())
        updateSnippitUrl(true)

    let codeArea = document.getElementById("code")
    codeEditor = CodeMirror.fromTextArea(codeArea, {
        lineNumbers: true,
        autofocus: true,
        indentUnit: 4,
        mode: "setlx"
    })
    codeEditor.on("change", function (e) {
        if (isSharedCodePage()) {
            window.history.replaceState(undefined, undefined, "/")
            updateSnippitUrl(false)
        }
    })
}

function isSharedCodePage() {
    return !!window.location.pathname.match(/^\/c\/[a-zA-Z0-9]+$/)
}

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}


function run() {
    if (executing)
        return
    let code = codeEditor.getValue()
    // check if code is empty
    if (!code)
        return

    let out = document.getElementById("output")
    executing = true
    out.innerHTML = "<span class='info'>" + "Waiting for remote server..." + "</span>"
    fetch("/run", {
        method: "POST",
        headers: {
            "Content-Type": "text/plain; charset=utf-8",
        },
        body: code
    }).then(async (response) => {
        executing = false
        if (response.ok)
            return response.json()
        let errorMsg = await response.text()
        throw errorMsg ? errorMsg : response.statusText
    }).then(async (json) => {
        out.innerHTML = ""
        let log = ""
        for (var msg of json.Events) {
            log += `<span class="${msg.Kind}">${msg.Text}</span>`
        }
        out.innerHTML += log;
        out.innerHTML += `<span class="info">Program exited.</span>`
        out.scrollTo(0, out.scrollHeight)
    }).catch((e) => {
        executing = false
        out.innerHTML = `<span class="stderr">${e}</span>`
    })
}

function share() {
    // check if user is on shared code page
    if (isSharedCodePage())
        return

    let code = codeEditor.getValue()
    fetch("/share", {
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
    }).then((response) => {
        // redirect to code page
        window.history.replaceState(undefined, undefined, "/c/" + response.id)
        updateSnippitUrl(true)
    }).catch(e => {
        console.log(e)
        alert("something went wrong")
    })
}

function updateSnippitUrl(show) {
    snippetUrl.type = show ? "url" : "hidden"
    snippetUrl.value = show ? window.location.href : ""
    snippetUrl.setSelectionRange(0, snippetUrl.value.length)
}