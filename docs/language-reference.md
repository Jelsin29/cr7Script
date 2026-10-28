# CR7Script Language Reference

> **Status: approved design; implementation in progress.** The owner-approved 0.1 decisions are summarized below. The lexer is the first implementation slice; no interpreter or CLI executes the examples yet.
>
> **Documentation language:** English · **Language vocabulary:** Spanish-first · **Identity:** a small scripting language inspired by Cristiano Ronaldo, especially his Real Madrid era.

CR7Script gives programming concepts a football identity while keeping their behavior predictable. The approved 0.1 baseline specifies a tree-walking scripting language implemented in Go 1.27 or newer; Go is not CR7Script's syntax or a promise to transpile to Go. The lexer/frontend is the first implementation slice; examples remain specifications until an interpreter verifies them.

## Quick example

```cr7
jugada comandante() {
    leyenda dorsal = 7;
    cantera goles = 0;

    entrena goles < 3 {
        goles = goles + 1;
        siuuu("Gol", goles, "del", dorsal);
    }

    si goles == 3 {
        siuuu("Hat-trick del Comandante");
    } sino {
        siuuu("A seguir entrenando");
    }
}
```

The proposed runner calls the top-level `comandante()` once. Expected stdout:

```text
Gol 1 del 7
Gol 2 del 7
Gol 3 del 7
Hat-trick del Comandante
```

This is a specification example, not a verified executable result.

## Language identity and boundaries

| CR7 reference | Proposed language role | Important boundary |
|---|---|---|
| CR7 / `.cr7` | Language name and source extension | Unofficial fan-inspired project |
| El Comandante | `comandante()` entry function | Ordinary function when imported or used in the REPL |
| El Bicho | `el_bicho(text)` uppercase transformation | Invented programming behavior |
| Mr. Champions | `mr_champions` statistics module | Mathematical operations, not sporting rankings |
| Rueda de prensa | `rueda_de_prensa` quote module | Immutable sourced text constants; no implicit output |
| SIUUU | `siuuu(...)` output function | Prints only when explicitly called |
| Hala Madrid | `hala_madrid()` greeting | Madrid-era presentation reference |
| Chilena | `chilena(list)` shallow reversal | Original list remains unchanged |
| Hat-trick | `hat_trick(value)` | Returns exactly three references to the value |
| Training | `entrena` while loop | Repeats while a boolean condition is true |
| A play | `jugada` function declaration | Reusable function |
| A comeback | `remontada` recovery block | Does not roll back earlier side effects |
| VAR | `var` catch clause | Reads a runtime error; does not retry automatically |

These programming meanings are invented. Nicknames, gestures, and football terms are not quotations. Avoid attributing unverified viral quotes to Ronaldo. Choose one spelling per operation; do not add `siu`, `siiiuuu`, and `siuuuu` aliases. Keep long phrases for occasional library names, not routine control flow.

### Verified spoken-phrase catalog

This catalog adds sourced phrases rather than unattributed motivational sayings. Spanish wording is retained; explanations remain English. Quotation marks identify a short excerpt reproduced from the linked source, not a transcription of an entire interview. Published punctuation may differ from speech.

**Evidence policy:** original broadcast/audio pages and contemporaneous reporting with attributed direct speech qualify as documentary support. A page offering a recording does not mean that its audiovisual content was independently played or transcribed for this revision. Evidence notes distinguish the available recording from a published quotation. Reposts alone do not establish the original occasion or exact wording. This update records the owner-provided source catalog; its linked material has not been independently re-audited here.

| ID / proposed export | Confirmed Spanish excerpt | Occasion and date | Evidence and context |
|---|---|---|---|
| Q1 / `RICO_GUAPO_Y_GRAN_JUGADOR` | “soy guapo, rico y un gran jugador” | Mixed-zone comments after Dinamo Zagreb–Real Madrid, September 14, 2011; report published September 15 | [Atlas video report published by Público](https://www.publico.es/deportes/video-cristiano-ronaldo-me-envidia-soy-rico-guapo-gran-jugador.html) reproduces the explanation he gave for hostile treatment. This is an excerpt, not the full sentence. |
| Q2 / `ESTOY_TRISTE` | “No lo hago cuando estoy triste” | Mixed-zone comments after Real Madrid–Granada, September 2, 2012 | [RTVE contemporaneous coverage](https://www.rtve.es/deportes/20120902/cristiano-estoy-triste-dentro-del-club-saben-por/561162.shtml) quotes his explanation for not celebrating his goals. The surrounding answer describes a professional reason; do not turn it into a claim about a medical condition. |
| Q3 / `TODOS_A_MI_NIVEL` | “Si todos estuviesen en mi nivel seríamos los primeros” | Post-derby media remarks, February 27, 2016 | [RNE / RTVE audio page](https://www.rtve.es/play/audios/tablero-deportivo/tablero-deportivo-ronaldo-si-todos-estuviesen-nivel-seriamos-primeros/3502911/) reproduces this wording in its description. Its headline instead uses “a mi nivel”; retain the description's wording rather than silently merging variants. This was a contested post-match statement, not a general teamwork slogan. |
| Q4 / `CINCO_CHAMPIONS` | “Cinco Champions tengo yo y el Atleti, cero” | Comments to journalists after Atlético–Juventus, February 20, 2019; broadcast February 21 | [RTVE's original broadcast page](https://www.rtve.es/play/videos/telediario-1/cristiano-cinco-champions-tengo-yo-atleti-cero/5001923/) identifies the remark and context. **Juventus era, not Real Madrid era.** This records a historical boast, not a live statistics API. |
| Q5 / `THE_BEST_SOY_YO` | “el premio The Best fue para el 'the best', que soy yo” | Post-award media interview, reported January 10, 2017 | [AS interview transcript](https://as.com/us/2017/01/10/futbol/1484003226_319451.html) reproduces his answer after the FIFA award. This entry relies on attributed published speech, not a newly reviewed original recording. |
| Q6 / `SOY_EL_MEJOR` | “Soy el mejor jugador de la historia” | France Football interview, reported December 8, 2017 | [RTVE reporting of the interview](https://www.rtve.es/deportes/20171208/ronaldo-soy-mejor-jugador-historia/1643981.shtml). This is his stated opinion, not a factual ranking established by CR7Script. |
| Q7 / `HALA_MADRID` | “Uno, dos, tres: ¡Hala Madrid!” | Real Madrid presentation, July 6, 2009 | [RTVE event recording page](https://www.rtve.es/play/videos/programa/uno-dos-tres-hala-madrid/539094/). **Public presentation, not an interview or press conference.** Retained as an existing cultural reference and clearly separated from the interview-based additions. |

The export names are CR7Script adaptations, not claims that Ronaldo used those identifiers. Proposed constants contain exactly the corresponding excerpt above, with typographic quotation marks removed; they do not append commentary or automatically print anything.

### Requested phrase still awaiting verification: “losh dosh”

User-supplied lead: [Instagram reel DVWnG4qDYqH](https://www.instagram.com/reel/DVWnG4qDYqH/?hl=en).

- The reel could not be retrieved during the supplied revision because the fetch was throttled.
- Its video, speaker, question, original source, date, and exact answer have **not** been verified.
- “losh dosh” is the user's phonetic/meme spelling. It must not be presented as an authenticated verbatim transcript.
- No event, interviewer, or meaning is inferred from unrelated search results containing “los dos.”
- `losh_dosh` / `los_dos` is **not** an accepted keyword, export, or function in this draft. If footage and context are verified later, a two-value operation could be considered separately; that programming meaning would still be invented.
- To clear this entry, obtain the clip or an accessible original broadcast/interview, identify the question and answer, and attach a source plus timestamp. A repost date must not be substituted for the interview date.

**Excluded from the spoken-phrase catalog:** the love/hatred motivational line previously considered as inspiration. [Contemporaneous reporting](https://www.mundodeportivo.com/futbol/real-madrid/20150824/201378108430/motivo-enigmatico-mensaje-cristiano-ronaldo.html) describes an English-language Instagram post, not an interview or press-conference statement. That evidence does not meet this revision's spoken-context requirement. No claim is made that he never said it elsewhere.

### Proposed 0.1 boundaries

| Area | Draft decision |
|---|---|
| Runtime | Tree-walking interpreter; bytecode is a possible later optimization only |
| Types | Dynamic values with strict runtime checks |
| Evaluation | Eager, left to right |
| Scope | Lexical block scope; closures supported |
| Concurrency | Single-threaded; no async or parallel syntax |
| Memory | Host-managed allocation and collection |
| Mutation | Explicit assignment and named collection operations |
| Source | UTF-8 `.cr7` files |
| Statements | Simple statements require semicolons; blocks use braces |
| Dependencies | Go 1.27 or newer; standard library and local modules only |
| Initial tooling | Runner, syntax checker, REPL, version and help commands (all proposed) |

**Deferred from 0.1:** classes, inheritance, operator overloading, macros, generics, threads, network access, package installation, JIT, native code generation, and self-hosting. This runtime is learning-oriented and is **not a security sandbox**.

## Source and lexical rules

- Identifiers match `[A-Za-z_][A-Za-z0-9_]*` and are case-sensitive.
- Keywords are lowercase ASCII. Strings and comments may contain Unicode.
- Indentation is conventional, not syntactic; use four spaces.
- `//` starts a line comment; `/* ... */` is a non-nesting block comment.
- Simple statements end in `;`. There is no automatic semicolon insertion.
- Blocks, function declarations, and compound statements need no trailing semicolon.
- There are no significant newlines, interpolation, multiline strings, or trailing commas in 0.1.
- A keyword is reserved only when the full identifier matches; `golazo` is not `gol`.
- A `{` at statement start begins a block. Parenthesize a standalone dictionary expression.

```cr7
// Comments can explain a play.
cantera mensaje = "¡Vamos, equipo!";
/* This is also a comment. */
```

### Literals

```cr7
7
3.5
1e3
"Cristiano"
"Primera linea\nSegunda linea"
gol
fuera
nada
[7, 10, 11]
{"nombre": "Cristiano", "dorsal": 7}
```

Numbers follow `[0-9]+(\.[0-9]+)?([eE][+-]?[0-9]+)?`; minus is unary. Hex, binary, numeric separators, and `.5` are unsupported. A syntactically valid literal that is non-finite (for example, `1e999`) is rejected during front-end analysis with `NumeroNoFinito` at the literal's source location, before program execution. Strings use double quotes and recognize `\"`, `\\`, `\n`, `\r`, `\t`, and `\uXXXX`. Unicode escapes must denote scalar values; surrogate escapes, unknown escapes, and raw line breaks are lexical errors.

## Keywords

| Keyword | Meaning |
|---|---|
| `cantera` | Declare a mutable binding |
| `leyenda` | Declare a non-reassignable binding |
| `jugada` | Declare a function |
| `devuelve` | Return a value |
| `si`, `sino` | Conditional branch |
| `entrena` | While loop |
| `temporada`, `en` | Iterate a collection |
| `vestuario`, `siguiente` | Break / continue the nearest loop |
| `remontada`, `var` | Recoverable block / catch binding |
| `falta` | Raise or rethrow a runtime error |
| `fichar`, `como` | Import a module / name its alias |
| `y`, `o`, `no` | Boolean AND, OR, NOT |
| `gol`, `fuera` | Boolean true, false |
| `nada` | Null |

`comandante`, `siuuu`, `calma`, and other library names are identifiers, not keywords. `var` is a catch clause, never a variable declaration.

## Values and types

| `tipo(value)` | Proposed meaning |
|---|---|
| `numero` | Finite IEEE-754 binary64 |
| `texto` | Immutable Unicode text |
| `booleano` | Boolean |
| `nulo` | Absent value |
| `lista` | Mutable ordered collection |
| `diccionario` | Mutable string-keyed collection |
| `funcion` | User or native function |
| `modulo` | Imported read-only namespace |
| `error` | Runtime error value |

Rules:

1. Do not implicitly convert numbers, strings, and booleans. Conditions and `y`/`o` require booleans; `si 7 { ... }` is an error.
2. `/` is floating-point division (`7 / 2` is `3.5`). Division and remainder by zero raise `DivisionPorCero`.
3. `%` uses a quotient truncated toward zero (`-7 % 3` is `-1`).
4. Results that are infinity or NaN raise `NumeroNoFinito`.
5. Integers are exactly representable through magnitude `2^53 - 1`; larger integer values may be rounded. This is unsuitable for exact financial arithmetic.
6. Indexes are non-negative safe integers; negative indexes are unsupported.
7. Text length and indexing count Unicode scalar values, not UTF-8 bytes or grapheme clusters.
8. `numero(text)` parses a signed decimal number under the literal grammar, allowing surrounding ASCII whitespace; invalid input raises `ConversionInvalida`.

Equality compares scalar values; unlike types are unequal, and numbers do not equal booleans. Lists, dictionaries, functions, modules, and errors compare by identity, not by recursively comparing contents.

## Bindings and scope

```cr7
cantera goles = 0;
goles = goles + 1;

leyenda dorsal = 7;
// dorsal = 10;  // ReasignacionDeLeyenda
```

- Every declaration needs an initializer. Assignment updates the nearest visible binding; it never creates a variable.
- Same-scope redeclaration is an error; nested scopes may shadow outer names.
- Blocks create scopes. Function parameters live in the function body's outer scope and are mutable bindings; reassigning a parameter changes the local binding only, while passed collection references retain identity.
- Names declared with `jugada` are mutable function bindings and may be reassigned.
- `leyenda` prevents rebinding, not mutation of a referenced collection.
- Collections are references; assignment does not copy them.
- The resolver reports an unknown lexical name as `NombreNoDefinido` and a read before initialization as `NombreSinInicializar`. A later local declaration shadows outer names from its lexical scope, so an early read cannot fall back to an outer binding.
- Core prelude names cannot be redeclared. Import aliases are immutable module-level bindings.

```cr7
leyenda plantilla = [7];
convoca(plantilla, 10);  // The list changes; its binding does not.
cantera otra = plantilla;
otra[0] = 9;
siuuu(plantilla[0]);    // 9
```

## Operators

Precedence, highest first:

| Level | Operators | Behavior |
|---|---|---|
| 1 | calls `f(...)`, indexing `x[i]`, property `m.name` | Postfix, left associative |
| 2 | `-`, `no` | Unary, right associative |
| 3 | `*`, `/`, `%` | Numeric, left associative |
| 4 | `+`, `-` | Left associative |
| 5 | `<`, `<=`, `>`, `>=` | Compare numbers |
| 6 | `==`, `!=` | Equality |
| 7 | `y` | Short-circuit AND |
| 8 | `o` | Short-circuit OR |

`+` takes two numbers or two texts, never a mixed pair. Relational operators take numbers. Chained comparisons are rejected; write `1 < x y x < 10`. `y` and `o` return booleans and skip an unnecessary right operand. Assignment is a statement, not an expression; no chained assignment, `++`, or `+=` in 0.1.

```cr7
cantera listo = goles >= 3 y no lesionado;
cantera etiqueta = "CR" + texto(7);
```

## Conditions and iteration

```cr7
si goles > 3 {
    siuuu("Exhibicion");
} sino si goles == 3 {
    siuuu("Hat-trick");
} sino {
    siuuu("Sigue entrenando");
}
```

`entrena condition { ... }` repeats while the condition is true. `temporada name en iterable { ... }` iterates lists in order, text as one-scalar strings, and dictionaries by insertion-ordered keys. Iteration captures a shallow snapshot at loop entry. Each iteration creates a fresh mutable loop binding and body scope; closures capture that iteration's binding. Other values are not iterable. `vestuario;` and `siguiente;` outside a loop in the same function raise the static error `ControlFueraDeBucle`.

```cr7
temporada rival en ["Porto", "Juventus", "Barcelona"] {
    siuuu(rival);
}

temporada jornada en rango(1, 4) {
    siuuu(jornada);  // 1, 2, 3
}
```

A training example with `intentos` incremented from 0 to 7, `siguiente` at 2, and `vestuario` at 5 prints `1`, `3`, and `4` on separate lines.

## Functions and entry point

```cr7
jugada rematar(potencia, precision) {
    devuelve potencia * precision;
}

jugada comandante() {
    siuuu(rematar(7, 10));
    devuelve 0;
}
```

Functions are first-class lexical closures. Arguments evaluate left to right and pass by value; collection references retain identity. Calls require exact arity; user-defined variadic parameters and overloads are absent in 0.1. Recursion is supported. Reaching the end or using `devuelve;` returns `nada`. Nested functions initialize when execution reaches their declaration. Module-level functions initialize before constants, allowing forward calls among top-level functions. Module-level `leyenda` initializers then evaluate in source order; the resolver reports a read of a later constant as `NombreSinInicializar`. Duplicate parameter names raise `DeclaracionDuplicada`; `devuelve` outside a function raises `DevolucionFueraDeFuncion`.

A runnable entry module declares exactly one top-level `jugada comandante()` with no parameters. A runner invokes it once after module initialization. It must return `nada` or an integer exit code from 0 through 125; other results raise `SalidaInvalida`. Imported modules need no entry point and do not auto-run an exported `comandante`. The REPL has no required or automatic entry call.

## Collections

```cr7
cantera plantilla = ["Cristiano", "Benzema", "Bale"];
siuuu(plantilla[0]);
plantilla[1] = "Marcelo";

cantera jugador = {"nombre": "Cristiano", "goles": 3};
jugador["goles"] = 4;
jugador["club"] = "Madrid";
```

- List indexes range from 0 to `palmares(list) - 1`. Assigning at the list length does not append; use `convoca`.
- Missing list/text indexes raise `IndiceFueraDeCampo`; missing dictionary keys raise `ClaveAusente`.
- Dictionary keys are text. Use `consulta` for a default, including when stored value is `nada`.
- Dictionary literals reject duplicate keys; updating a key preserves its insertion position.
- Dictionary field access uses brackets. Dot access is for module exports and read-only error fields.
- Strings cannot be assigned by index. Slices and destructuring are deferred.
- Self-referential collections are allowed; display detects cycles.

## Errors and recovery

```cr7
jugada penalti(acierto) {
    si no acierto {
        falta "Penalti fallado";
    }
    devuelve "Gol";
}

jugada comandante() {
    remontada {
        siuuu(penalti(fuera));
    } var error {
        siuuu("Revision VAR:", error.codigo, error.mensaje);
    }
}
```

`remontada` requires exactly one `var identifier { ... }` handler. It catches runtime errors from the protected block, not lexical, parse, resolution, or module-loading errors. Catch bindings are immutable and scoped to the handler. `falta` takes text (creates `FaltaDelUsuario`) or an existing error (rethrows it). Effects before failure remain; recovery is not a transaction or retry. `devuelve`, `vestuario`, and `siguiente` are control flow, not catchable errors. There is no `finally` in 0.1.

Errors expose read-only `.codigo`, `.mensaje`, `.archivo`, `.linea`, and `.columna`. Locations are one-based; columns count Unicode scalar values. CLI diagnostics additionally print stack frames. Human messages are English by default; codes and syntax are Spanish-themed.

| Stable code | Meaning |
|---|---|
| `SintaxisInvalida` | Invalid source grammar or token |
| `NombreNoDefinido` | Unknown identifier |
| `NombreSinInicializar` | Binding read too early |
| `DeclaracionDuplicada` | Duplicate name in scope |
| `DevolucionFueraDeFuncion` | Return outside a function |
| `ControlFueraDeBucle` | Break or continue outside a loop |
| `ReasignacionDeLeyenda` | Reassignment of a constant or read-only import alias |
| `TipoIncorrecto` | Invalid operand or argument type |
| `AridadIncorrecta` | Wrong number of arguments |
| `DivisionPorCero` | Zero divisor |
| `NumeroNoFinito` | Infinite or NaN result |
| `ConversionInvalida` | Invalid numeric conversion |
| `IndiceFueraDeCampo` | Invalid collection/text index |
| `ClaveAusente` | Missing dictionary key |
| `ListaVacia` | Operation requires an element |
| `RangoInvalido` | Invalid range bounds/direction |
| `EntradaSalida` | Input/output failure |
| `FaltaDelUsuario` | Explicit user-raised failure |
| `ModuloNoEncontrado` | Unresolvable module |
| `FichajeCircular` | Cyclic module dependency |
| `ComandanteInvalido` | Missing/malformed entry point |
| `SalidaInvalida` | Invalid entry return value |
| `LimiteDeEjecucion` | Interpreter call-depth limit reached |

Example diagnostic shape (proposed):

```text
Tarjeta roja [DivisionPorCero]: cannot divide by zero.
  at rematar (partido.cr7:4:12)
  at comandante (partido.cr7:9:5)
```

## Modules

```cr7
fichar "mr_champions" como champions;
fichar "./jugadas.cr7" como tactica;

jugada comandante() {
    siuuu(champions.soy_el_mejor([7, 10, 17]));
    siuuu(tactica.rematar(7, 10));
}
```

Imports are top-level and appear before other module declarations. Non-relative names resolve only to registered standard modules; no remote downloads. Local imports start with `./` or `../`, include `.cr7`, resolve relative to the importer, and reject absolute paths. Canonical filesystem paths identify/cache modules; symlinks resolve; cycles are rejected with the import chain. Modules initialize once per process, in dependency order. Top-level code allows imports, functions, and initialized `leyenda` bindings only; executable statements and mutable bindings belong inside functions. Top-level functions/constants are exported. Import namespaces are read-only, though referenced collections remain mutable. The runner is not a sandbox; path resolution does not isolate untrusted programs.

## Standard library proposal

### Core prelude

| Name | Proposed contract |
|---|---|
| `siuuu(value...)` | Print values separated by spaces, then newline; zero args prints a blank line |
| `entrevista(prompt)` | Print prompt without newline, flush, read one line; EOF returns `nada` |
| `hala_madrid()` | Print `Uno, dos, tres... ¡Hala Madrid!` plus newline |
| `el_bicho(text)` | Locale-independent Unicode simple uppercase |
| `texto(value)` | Stable display conversion |
| `numero(text)` | Explicit decimal conversion |
| `tipo(value)` | Return type label from this reference |
| `palmares(value)` | Length of list, dictionary, or text |
| `calma(ms)` | Block at least non-negative safe-integer milliseconds, barring interruption |
| `rango(start,end)` | Safe-integer list, inclusive start/exclusive end, step 1 |
| `chilena(list)` | New shallow reversed list; original unchanged |
| `hat_trick(value)` | New `[value, value, value]`; reference values are not cloned |
| `convoca(list,value)` | Append in place |
| `banquillo(list)` | Remove/return last item; empty raises `ListaVacia` |
| `consulta(dict,key,fallback)` | Return present value, even `nada`; otherwise fallback |
| `contiene(collection,item)` | List membership, dictionary key test, or text substring |
| `exige(condition,message)` | False raises `FaltaDelUsuario` with message |

`rango` rejects unsafe/non-integer bounds and `start > end`; equal bounds return an empty list. Large ranges allocate memory. `entrevista` strips one LF or CRLF, preserves other whitespace, returns a final unterminated line, then returns `nada` at EOF. I/O failures raise `EntradaSalida`.

Display is deterministic: booleans `gol`/`fuera`, null `nada`, top-level text unquoted, whole-valued numbers without `.0`, negative zero as `0`; non-integral numbers use shortest round-trip decimal with lowercase `e` when needed. Lists/dictionaries use brackets and comma-space separators, escaped quoted text elements, and dictionary insertion order. A cycle on the current display traversal stack renders `<ciclo>`; opaque values render `<funcion>`, `<modulo>`, or `<error CODE>`.

### `mr_champions`

- `soy_el_mejor(numbers)`: maximum; empty list raises `ListaVacia`.
- `minimo(numbers)`: minimum; empty raises `ListaVacia`.
- `goles_totales(numbers)`: left-to-right sum; empty returns 0.
- `promedio(numbers)`: arithmetic mean; empty raises `ListaVacia`.
- `pichichi(players)`: return original dictionary with greatest numeric `goles`; first wins ties. Require non-empty list of dictionaries; missing key raises `ClaveAusente`, invalid type raises `TipoIncorrecto`.

Every numeric list element must be a number and arithmetic must remain finite.

### Text and numeric modules

`relato` proposes `minusculas(text)`, `separa(text, separator)`, `une(texts, separator)`, and `reemplaza(text, old, replacement)`. Split preserves empty fields; join takes a list of texts; replace all non-overlapping matches. Empty separators/search strings raise `TipoIncorrecto`.

`marcador` proposes `absoluto(number)`, `redondea(number)` (half away from zero), and `potencia(base, exponent)` (non-finite result raises `NumeroNoFinito`). Neither module provides filesystem or network access in 0.1.

### `rueda_de_prensa` — sourced quote constants

This proposed standard module exposes the seven documented excerpts above as immutable text exports. It introduces no keywords, operator aliases, or changes to existing functions. The quotations are data, not executable language syntax.

```cr7
fichar "rueda_de_prensa" como frases;

jugada comandante() {
    siuuu(frases.RICO_GUAPO_Y_GRAN_JUGADOR);
    siuuu(frases.TODOS_A_MI_NIVEL);
    siuuu(frases.THE_BEST_SOY_YO);
}
```

| Export | Catalog entry | Runtime type |
|---|---|---|
| `frases.RICO_GUAPO_Y_GRAN_JUGADOR` | Q1 | `texto` |
| `frases.ESTOY_TRISTE` | Q2 | `texto` |
| `frases.TODOS_A_MI_NIVEL` | Q3 | `texto` |
| `frases.CINCO_CHAMPIONS` | Q4; explicitly post-Madrid | `texto` |
| `frases.THE_BEST_SOY_YO` | Q5 | `texto` |
| `frases.SOY_EL_MEJOR` | Q6 | `texto` |
| `frases.HALA_MADRID` | Q7; presentation rather than interview | `texto` |

The constants preserve source-excerpt punctuation, including Q7's colon. The existing `hala_madrid()` greeting deliberately uses an ellipsis; it is an existing output adaptation, not an alternative verbatim transcript. Existing `mr_champions.soy_el_mejor(numbers)` still computes a maximum; the uppercase quote constant merely stores text. The unverified reel contributes no export until verification is complete.

## Grammar proposal

The grammar below is a compact parser target; semantic restrictions above take precedence. `{ rule }` means repetition, `[ rule ]` optional.

```ebnf
program       = { importDecl }, { moduleDecl }, EOF ;
importDecl    = "fichar", STRING, "como", IDENT, ";" ;
moduleDecl    = constDecl | functionDecl ;
declaration   = mutableDecl | constDecl | functionDecl | statement ;
mutableDecl   = "cantera", IDENT, "=", expression, ";" ;
constDecl     = "leyenda", IDENT, "=", expression, ";" ;
functionDecl  = "jugada", IDENT, "(", [ parameters ], ")", block ;
parameters    = IDENT, { ",", IDENT } ;
block         = "{", { declaration }, "}" ;
statement     = ifStmt | whileStmt | forStmt | tryStmt
              | returnStmt | breakStmt | continueStmt | throwStmt
              | block | simpleStmt ;
ifStmt        = "si", expression, block, [ "sino", ( ifStmt | block ) ] ;
whileStmt     = "entrena", expression, block ;
forStmt       = "temporada", IDENT, "en", expression, block ;
tryStmt       = "remontada", block, "var", IDENT, block ;
returnStmt    = "devuelve", [ expression ], ";" ;
breakStmt     = "vestuario", ";" ;
continueStmt  = "siguiente", ";" ;
throwStmt     = "falta", expression, ";" ;
simpleStmt    = expression, [ "=", expression ], ";" ;
expression    = logicalOr ;
logicalOr     = logicalAnd, { "o", logicalAnd } ;
logicalAnd    = equality, { "y", equality } ;
equality      = comparison, { ( "==" | "!=" ), comparison } ;
comparison    = term, [ ( "<" | "<=" | ">" | ">=" ), term ] ;
term          = factor, { ( "+" | "-" ), factor } ;
factor        = unary, { ( "*" | "/" | "%" ), unary } ;
unary         = ( "no" | "-" ), unary | postfix ;
postfix       = primary, { "(", [ arguments ], ")"
                         | "[", expression, "]"
                         | ".", IDENT } ;
arguments     = expression, { ",", expression } ;
primary       = NUMBER | STRING | "gol" | "fuera" | "nada" | IDENT
              | "(", expression, ")" | listLiteral | dictLiteral ;
listLiteral   = "[", [ arguments ], "]" ;
dictLiteral   = "{", [ pair, { ",", pair } ], "}" ;
pair          = STRING, ":", expression ;
```

Only identifiers and final index-access expressions are valid assignment targets. Dot assignment is rejected. A `{` at statement start is a block; parenthesize a standalone dictionary expression. Imports are not allowed in blocks. A REPL accepts block declarations and standalone expressions without a semicolon for value echo.

## Owner-approved implementation decisions

These decisions resolve the draft's previously open questions and are part of the approved 0.1 baseline:

1. **Host:** Go is confirmed, with Go 1.27 as the minimum supported version.
2. **Module constants:** Initialize top-level functions first, then evaluate module constants in source order. A reference to a later constant is a resolver error, `NombreSinInicializar`.
3. **Non-finite literals:** Reject a non-finite numeric literal such as `1e999` during front-end analysis as `NumeroNoFinito`, before executing program code, at the literal's source location.
4. **Names and `check`:** The resolver reports unknown names as `NombreNoDefinido` and reads before initialization as `NombreSinInicializar`. `check` validates syntax and lexical bindings without executing the program.

## Inspiration and attribution

References support historical inspiration only, not the invented programming semantics:

- 2009 presentation / Hala Madrid: [RTVE, July 6, 2009](https://www.rtve.es/play/videos/programa/uno-dos-tres-hala-madrid/539094/).
- SIU celebration: [Premier League report](https://www.premierleague.com/ar/news/4638156); the spelling is not normative.
- Chilena: [FIFA feature](https://inside.fifa.com/en/news/ronaldo-gravity-defying-acrobatics).
- El Bicho / El Comandante nicknames: [AS report](https://as.com/us/us/2019/03/12/futbol/1552399674_237315.html).
- Mr. Champions League nickname: [GOAL report](https://www.goal.com/en/news/why-cristiano-ronaldo-mr-champions-league-nickname-meaning-explained/bltfcfc7d3f39ee6633).
- Best-player self-description report: [RTVE, December 8, 2017](https://www.rtve.es/deportes/20171208/ronaldo-soy-mejor-jugador-historia/1643981.shtml).
- Go implementation rationale: [Go GC guide](https://go.dev/doc/gc-guide).

### Spoken-phrase revision sources

The source links in the Q1–Q7 catalog above are part of this specification. The evidence notes intentionally distinguish published text from an accessible recording and do not claim independent audiovisual review during this documentation update.

8. **Q1 / mixed-zone statement, 2011:** [Atlas report and video published by Público](https://www.publico.es/deportes/video-cristiano-ronaldo-me-envidia-soy-rico-guapo-gran-jugador.html).
9. **Q2 / post-Granada explanation, 2012:** [RTVE contemporaneous report](https://www.rtve.es/deportes/20120902/cristiano-estoy-triste-dentro-del-club-saben-por/561162.shtml).
10. **Q3 / post-derby remarks, 2016:** [RNE / RTVE audio record and published description](https://www.rtve.es/play/audios/tablero-deportivo/tablero-deportivo-ronaldo-si-todos-estuviesen-nivel-seriamos-primeros/3502911/).
11. **Q4 / five Champions remark, 2019:** [RTVE broadcast record](https://www.rtve.es/play/videos/telediario-1/cristiano-cinco-champions-tengo-yo-atleti-cero/5001923/).
12. **Q5 / The Best interview, 2017:** [AS published interview transcript](https://as.com/us/2017/01/10/futbol/1484003226_319451.html).
13. **Excluded social-post context:** [Mundo Deportivo contemporaneous report, August 24, 2015](https://www.mundodeportivo.com/futbol/real-madrid/20150824/201378108430/motivo-enigmatico-mensaje-cristiano-ronaldo.html). Establishes a social-post context, not the requested spoken context.
14. **Unverified user lead:** [Instagram reel](https://www.instagram.com/reel/DVWnG4qDYqH/?hl=en). Retrieval failed in the supplied revision; not evidence for any accepted quote.

CR7Script is an unofficial fan-inspired project. No endorsement by Cristiano Ronaldo, Real Madrid, UEFA, or other organizations is asserted. Names, likenesses, logos, and commercialization may require separate permission; this specification grants no rights.
