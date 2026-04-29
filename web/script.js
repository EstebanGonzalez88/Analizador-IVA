function analyze(){

    let expr = document.getElementById("expr").value

    fetch("/analyze",{
        method:"POST",
        headers:{
            "Content-Type":"application/json"
        },
        body: JSON.stringify({expr})
    })
    .then(r=>r.json())
    .then(data=>{

        let div = document.getElementById("result")
        div.innerHTML = ""

        // Si hay error
        if(data.error){
            div.innerHTML += `<div class="card error"><h3> Error</h3>${data.error}</div>`
            return
        }

        // Léxico
        let lexicoHTML = data.lexico.map(t => `<span class="token">${t.tipo}: "${t.valor}"</span>`).join(" ")
        div.innerHTML += `
        <div class="card">
        <h3>Análisis Léxico</h3>
        <div class="tokens">${lexicoHTML}</div>
        </div>
        `

        // Sintáctico
        div.innerHTML += `
        <div class="card">
        <h3>Análisis Sintáctico</h3>
        ${data.sintactico}
        </div>
        `

        // Semántico
        div.innerHTML += `
        <div class="card">
        <h3>Análisis Semántico</h3>
        ${data.semantico}
        </div>
        `

        // C3D
        div.innerHTML += `
        <div class="card">
        <h3>Código Intermedio (3 Direcciones)</h3>
        <div class="c3d">${data.c3d.map(x => `<div>→ ${x}</div>`).join("")}</div>
        </div>
        `

        // Árbol
        div.innerHTML += `
        <div class="card">
        <h3>Árbol de Análisis Sintáctico</h3>
        <pre>${data.arbol}</pre>
        </div>
        `
    })
}