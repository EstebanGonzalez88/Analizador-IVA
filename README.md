# 🧠 Analizador Léxico, Sintáctico, Semántico y Generador de C3D para IVA Acumulado

## 📋 Requisitos Cumplidos

### ✅ 1. Analizador Léxico (REAL)
- **Detecta tokens dinámicamente** desde la entrada del usuario
- **Identifica tipos correctos:**
  - `IDENTIFICADOR`: variables (subtotal, retencion, iva)
  - `DECIMAL`: constantes como `0.16`
  - `ENTERO`: números enteros
  - `OPERADOR`: `+`, `-`, `*`
  - `ASIGNACION`: `=`
  - `PARENTESIS`: `(`, `)`
- **Salida:** Lista de tokens con tipo y valor

### ✅ 2. Analizador Sintáctico (REAL)
- **Parser Recursive Descent** que construye un árbol de sintaxis
- **Respeta precedencia de operadores:**
  - Multiplicación `*` tiene mayor precedencia que suma `+`
  - Paréntesis agrupan expresiones
- **Genera Árbol Sintáctico Abstracto (AST)** desde los tokens
- **Validación de estructura gramatical**

### ✅ 3. Analizador Semántico (REAL)
- **Verifica que las variables estén declaradas**
- **Variables de entrada reconocidas:**
  - `subtotal` ✓
  - `retencion` ✓
- **Valida que los identificadores desconocidos no se usen**
- **Detecta errores de variables no inicializadas**

### ✅ 4. Código Intermedio de 3 Direcciones (C3D)
```
→ t1 = subtotal * 0.16
→ t2 = t1 + retencion
→ iva = t2
```

**3 variables temporales mínimas:** `t1`, `t2` (y resultado final)

### ✅ 5. Árbol de Análisis Sintáctico
Genera un árbol ASCII mostrando:
- Operadores internos
- Operandos en hojas
- Respeto a precedencia (multiplicación arriba de suma)

**Ejemplo:**
```
+
├─ *
│  ├─ subtotal
│  └─ 0.16
└─ retencion
```

---

## 🚀 Cómo Usar

### Backend (Go)
```bash
cd go
go run main.go
```
Servidor disponible en: `http://localhost:8080`

### Frontend
Accede a la interfaz web y:
1. Ingresa la expresión (ej: `subtotal * 0.16 + retencion`)
2. Click en "Analizar"
3. Visualiza todos los análisis

---

## 📐 Fases del Análisis Completo

```
INPUT: "subtotal * 0.16 + retencion"
   ↓
┌─────────────────────────────────┐
│ 1️⃣  ANÁLISIS LÉXICO             │
│ Tokens: [subtotal] [*] [0.16]...│
└─────────────────────────────────┘
   ↓
┌─────────────────────────────────┐
│ 2️⃣  ANÁLISIS SINTÁCTICO         │
│ Construcción de AST             │
│ Precedencia: * antes que +      │
└─────────────────────────────────┘
   ↓
┌─────────────────────────────────┐
│ 3️⃣  ANÁLISIS SEMÁNTICO          │
│ Verificar variables inicializadas│
│ Tipos correctos                 │
└─────────────────────────────────┘
   ↓
┌─────────────────────────────────┐
│ 4️⃣  GENERACIÓN C3D              │
│ t1, t2, variables temporales    │
└─────────────────────────────────┘
   ↓
┌─────────────────────────────────┐
│ 5️⃣  ÁRBOL SINTÁCTICO ASCII       │
│ Visualización jerárquica        │
└─────────────────────────────────┘
```

---

## 🔧 Estructura Técnica

### Go (Backend)
- **analizarLexico()**: Tokeniza la expresión con regex
- **Parser (RDP)**: Analiza gramática y construye AST
- **analizarSemantico()**: Valida variables
- **generarC3D()**: Transversal postorden del AST
- **generarArbol()**: Serializa AST a formato ASCII

### JavaScript (Frontend)
- Envía expresión a `/analyze` (POST JSON)
- Formatea y visualiza resultados
- Manejo de errores con mensajes claros

---

## 📝 Ejemplos de Entrada

### ✅ Válido:
```
subtotal * 0.16 + retencion
(subtotal + retencion) * 0.16
subtotal * 0.16
```

### ❌ Inválido:
```
subtotal * 0.16 + desconocida    → variable no declarada
subtotal * 0.16)                 → paréntesis no cerrado
+ 0.16                           → factor esperado
```

---

## 📊 Conceptos Implementados

| Concepto | Implementación |
|----------|---|
| **Léxico** | Expresión regular con patrones para cada token |
| **Sintáctico** | Parser recursive descent con precedencia |
| **Semántico** | Validación de variables en árbol |
| **C3D** | Recorrido postorden generando 3AC |
| **AST** | Estructura de nodos con operadores y operandos |

---

## ✨ Mejoras Implementadas

✅ Análisis REAL de entrada (no hardcoded)
✅ Árbol dinámico basado en precedencia
✅ Validación de variables
✅ Mensajes de error descriptivos
✅ Interfaz mejorada con estilos
✅ Visualización de tokens tipados
✅ C3D con lógica de transversal postorden
