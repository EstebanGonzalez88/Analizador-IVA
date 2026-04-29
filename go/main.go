package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/xlab/treeprint"
)

type Resultado struct {
	Lexico     []Token   `json:"lexico"`
	Sintactico string    `json:"sintactico"`
	Semantico  string    `json:"semantico"`
	C3D        []string  `json:"c3d"`
	Arbol      string    `json:"arbol"`
	ArbolJSON  *NodoJSON `json:"arbolJson,omitempty"`
	Error      string    `json:"error"`
}

type Token struct {
	Tipo  string `json:"tipo"`
	Valor string `json:"valor"`
}

type Nodo struct {
	Valor      string
	Izq        *Nodo
	Der        *Nodo
	EsOperador bool
}

type NodoJSON struct {
	Valor string    `json:"valor"`
	Izq   *NodoJSON `json:"izq,omitempty"`
	Der   *NodoJSON `json:"der,omitempty"`
}

var variables = map[string]bool{
	"iva":       false,
	"subtotal":  true,
	"retencion": true,
}

var tempCounter = 0

func main() {
	http.Handle("/", http.FileServer(http.Dir("../web")))
	http.HandleFunc("/analyze", handler)

	fmt.Println("Servidor en http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var input struct {
		Expr string `json:"expr"`
	}

	json.NewDecoder(r.Body).Decode(&input)
	res := analizar(input.Expr)
	json.NewEncoder(w).Encode(res)
}

// 🔍 ANÁLISIS LÉXICO
func analizarLexico(expr string) ([]Token, error) {
	expr = strings.TrimSpace(expr)
	tokens := []Token{}

	// Patrones para identificar tokens
	patterns := []struct {
		tipo   string
		patron string
	}{
		{"IDENTIFICADOR", `[a-zA-Z_][a-zA-Z0-9_]*`},
		{"DECIMAL", `\d+\.\d+`},
		{"ENTERO", `\d+`},
		{"ASIGNACION", `=`},
		{"MULTIPLICACION", `\*`},
		{"SUMA", `\+`},
		{"PARENTESIS", `[()]{1}`},
	}

	posicion := 0
	for posicion < len(expr) {
		// Saltar espacios
		if expr[posicion] == ' ' {
			posicion++
			continue
		}

		encontrado := false
		for _, p := range patterns {
			regex := regexp.MustCompile(`^` + p.patron)
			matches := regex.FindStringIndex(expr[posicion:])
			if matches != nil {
				valor := expr[posicion : posicion+matches[1]]
				if p.tipo == "IDENTIFICADOR" {
					if _, existe := variables[valor]; !existe {
						return nil, fmt.Errorf("identificador inválido: %s", valor)
					}
				}
				tokens = append(tokens, Token{p.tipo, valor})
				posicion += matches[1]
				encontrado = true
				break
			}
		}

		if !encontrado {
			return nil, fmt.Errorf("token inválido en posición %d: %c", posicion, expr[posicion])
		}
	}

	return tokens, nil
}

// 🌳 ANÁLISIS SINTÁCTICO (Recursive Descent Parser)
func analizarSintactico(tokens []Token) (*Nodo, error) {
	parser := &Parser{tokens: tokens, pos: 0}
	arbol, err := parser.parseExpresion()
	if err != nil {
		return nil, err
	}
	if parser.pos < len(tokens) {
		return nil, fmt.Errorf("tokens no consumidos al final")
	}
	return arbol, nil
}

type Parser struct {
	tokens []Token
	pos    int
}

func (p *Parser) actual() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{"EOF", ""}
}

func (p *Parser) avanzar() {
	p.pos++
}

func (p *Parser) parseExpresion() (*Nodo, error) {
	// Manejar asignación: IDENTIFICADOR = EXPRESION
	if p.actual().Tipo == "IDENTIFICADOR" && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Tipo == "ASIGNACION" {
		varIzq := p.actual().Valor
		p.avanzar() // salta IDENTIFICADOR
		p.avanzar() // salta ASIGNACION

		// Parsear la expresión del lado derecho
		expDer, err := p.parseAdicion()
		if err != nil {
			return nil, err
		}

		// Crear nodo de asignación
		return &Nodo{Valor: "=", Izq: &Nodo{Valor: varIzq, EsOperador: false}, Der: expDer, EsOperador: true}, nil
	}

	// Si no hay asignación, procesar como expresión normal
	return p.parseAdicion()
}

func (p *Parser) parseAdicion() (*Nodo, error) {
	izq, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	for p.actual().Tipo == "SUMA" {
		op := p.actual().Valor
		p.avanzar()
		der, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		izq = &Nodo{Valor: op, Izq: izq, Der: der, EsOperador: true}
	}

	return izq, nil
}

func (p *Parser) parseTerm() (*Nodo, error) {
	izq, err := p.parseFactor()
	if err != nil {
		return nil, err
	}

	for p.actual().Tipo == "MULTIPLICACION" {
		op := p.actual().Valor
		p.avanzar()
		der, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		izq = &Nodo{Valor: op, Izq: izq, Der: der, EsOperador: true}
	}

	return izq, nil
}

func (p *Parser) parseFactor() (*Nodo, error) {
	if p.actual().Tipo == "PARENTESIS" && p.actual().Valor == "(" {
		p.avanzar()
		expr, err := p.parseExpresion()
		if err != nil {
			return nil, err
		}
		if p.actual().Valor != ")" {
			return nil, fmt.Errorf("paréntesis no cerrado")
		}
		p.avanzar()
		return expr, nil
	}

	if p.actual().Tipo == "IDENTIFICADOR" || p.actual().Tipo == "DECIMAL" || p.actual().Tipo == "ENTERO" {
		valor := p.actual().Valor
		p.avanzar()
		return &Nodo{Valor: valor, EsOperador: false}, nil
	}

	return nil, fmt.Errorf("factor esperado, encontrado: %s", p.actual().Valor)
}

// 🧠 ANÁLISIS SEMÁNTICO
func analizarSemantico(arbol *Nodo) error {
	variables["iva"] = true
	return verificarVariable(arbol)
}

func verificarVariable(nodo *Nodo) error {
	if nodo == nil {
		return nil
	}

	if !nodo.EsOperador {
		// Es una variable o constante
		if esDecimal(nodo.Valor) {
			// Es constante decimal válida
			return nil
		}
		if _, existe := variables[nodo.Valor]; !existe {
			return fmt.Errorf("❌ variable no declarada: %s", nodo.Valor)
		}
		return nil
	}

	// Verificar sub-árboles
	if err := verificarVariable(nodo.Izq); err != nil {
		return err
	}
	return verificarVariable(nodo.Der)
}

// ⚙️ GENERAR C3D
func generarC3D(arbol *Nodo) []string {
	tempCounter = 0
	_, instrucciones := generarC3DRec(arbol)
	return instrucciones
}

func generarC3DRec(nodo *Nodo) (string, []string) {
	if nodo == nil {
		return "", []string{}
	}

	instrucciones := []string{}

	// Si es un átomo (variable o constante)
	if !nodo.EsOperador {
		return nodo.Valor, instrucciones
	}

	// Si es asignación
	if nodo.Valor == "=" {
		// El lado izquierdo debe ser una variable
		varIzq := nodo.Izq.Valor
		// Procesar el lado derecho
		valDer, insDer := generarC3DRec(nodo.Der)
		instrucciones = append(instrucciones, insDer...)
		// Generar la asignación final
		instrucciones = append(instrucciones, fmt.Sprintf("%s = %s", varIzq, valDer))
		return varIzq, instrucciones
	}

	// Si es una operación binaria normal
	// Generar código para subárbol izquierdo
	valIzq, insIzq := generarC3DRec(nodo.Izq)
	instrucciones = append(instrucciones, insIzq...)

	// Generar código para subárbol derecho
	valDer, insDer := generarC3DRec(nodo.Der)
	instrucciones = append(instrucciones, insDer...)

	// Generar instrucción para este operador
	tempCounter++
	tempVar := "t" + fmt.Sprintf("%d", tempCounter)
	instrucciones = append(instrucciones, fmt.Sprintf("%s = %s %s %s", tempVar, valIzq, nodo.Valor, valDer))

	return tempVar, instrucciones
}

// 🌲 GENERAR ÁRBOL ASCII
func generarArbol(nodo *Nodo) string {
	if nodo == nil {
		return ""
	}

	root := treeprint.New()
	root.SetValue(nodo.Valor)
	agregarRama(root, nodo)
	return root.String()
}

func agregarRama(tree treeprint.Tree, nodo *Nodo) {
	if nodo == nil {
		return
	}

	if nodo.Izq != nil {
		branch := tree.AddBranch(nodo.Izq.Valor)
		agregarRama(branch, nodo.Izq)
	}
	if nodo.Der != nil {
		branch := tree.AddBranch(nodo.Der.Valor)
		agregarRama(branch, nodo.Der)
	}
}

func convertirNodoJSON(nodo *Nodo) *NodoJSON {
	if nodo == nil {
		return nil
	}

	return &NodoJSON{
		Valor: nodo.Valor,
		Izq:   convertirNodoJSON(nodo.Izq),
		Der:   convertirNodoJSON(nodo.Der),
	}
}

// FUNCIONES AUXILIARES
func esDecimal(s string) bool {
	_, err := regexp.MatchString(`^\d+\.?\d*$`, s)
	return err == nil
}

func analizar(expr string) Resultado {
	// Resetear variables globales
	tempCounter = 0
	variables = map[string]bool{
		"iva":       false,
		"subtotal":  true,
		"retencion": true,
	}

	// ANÁLISIS LÉXICO
	tokens, err := analizarLexico(expr)
	if err != nil {
		return Resultado{Error: err.Error()}
	}

	// ANÁLISIS SINTÁCTICO
	arbol, err := analizarSintactico(tokens)
	if err != nil {
		return Resultado{
			Lexico: tokens,
			Error:  err.Error(),
		}
	}

	// ANÁLISIS SEMÁNTICO
	err = analizarSemantico(arbol)
	if err != nil {
		return Resultado{
			Lexico:     tokens,
			Sintactico: "✔ Sintaxis correcta",
			Error:      err.Error(),
		}
	}

	// GENERAR C3D
	c3d := generarC3D(arbol)

	// GENERAR ÁRBOL ASCII
	arbolStr := generarArbol(arbol)

	// GENERAR ÁRBOL JSON
	arbolJSON := convertirNodoJSON(arbol)

	return Resultado{
		Lexico:     tokens,
		Sintactico: "✔ Sintaxis correcta",
		Semantico:  "✔ Variables válidas e inicializadas",
		C3D:        c3d,
		Arbol:      arbolStr,
		ArbolJSON:  arbolJSON,
		Error:      "",
	}
}
