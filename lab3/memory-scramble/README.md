# PR Lab 3: Multiplayer Game

**Author:** Gurschi Gheorghe

**Data:** 14 Noiembrie 2025

## Comenzi de Rulare

### Inițializare (doar prima dată)

```bash
# Navighează în directorul proiectului
cd memory-scramble

# Inițializează modulul Go
go mod init memory-scramble
```

### Rularea Testelor

```bash
# Rulează toate testele unit cu output detaliat
go test -v board_test.go board.go card.go player.go commands.go

# Sau pentru a vedea doar rezultatul final
go test board_test.go board.go card.go player.go commands.go
```

### Pornirea Serverului

```bash
# Pornește serverul HTTP pe portul 8080
go run board.go card.go player.go commands.go server.go
```

Serverul va porni pe `http://localhost:8080`

```###

```bash
# În alt terminal, după ce serverul pornește
go run simulate.go
```

---

## Structura Proiectului

```
memory-scramble/
├── board.go          # Board ADT - tabla de joc cu toate operațiile
├── card.go           # Card type - structura și validarea unei cărți
├── player.go         # PlayerState - starea unui jucător în joc
├── commands.go       # Logica regulilor jocului (flip, cleanup, replace)
├── server.go         # HTTP server și handler-ele pentru endpoints
├── board_test.go     # Unit tests pentru toate regulile
├── simulate.go       # Script de simulare multi-player
├── index.html        # Client web (interfața jocului)
├── perfect.txt       # Fișierul cu configurația tablei de joc
└── go.mod            # Definiția modulului Go
```

---

## Funcții Principale

### 1. FlipFirstCard - Întoarce Prima Carte

```go
func FlipFirstCard(board *Board, card *Card, row, col int, 
                   playerID string, playerState *PlayerState) bool
```

**Ce face:**

- Implementează **Regulile 1-A, 1-B, 1-C, 1-D** din specificație
- Încearcă să întoarcă prima carte pentru un jucător
- Returnează `true` dacă operația reușește, `false` altfel

**Reguli implementate:**


| Regulă | Situație                        | Acțiune                              | Return  |
| ------- | -------------------------------- | ------------------------------------- | ------- |
| 1-A     | Nu există carte (`Value == ""`) | Eșuează                             | `false` |
| 1-B     | Carte cu fața în jos           | Întoarce cartea, setează controller | `true`  |
| 1-C     | Carte vizibilă, necontrolată   | Preia controlul                       | `true`  |
| 1-D     | Carte controlată de altcineva   | Eșuează (ar aștepta)               | `false` |

**Exemplu de utilizare:**

```go
card := &board.Cards[0][0]
playerState := board.GetPlayerState("player1")

success := FlipFirstCard(board, card, 0, 0, "player1", playerState)

if success {
    fmt.Println("Prima carte întorsă cu succes!")
}
```

**Flow-ul funcției:**

```
FlipFirstCard()
    │
    ├─> Verifică Value == "" ? → return false (Regula 1-A)
    │
    ├─> Verifică !FaceUp ? 
    │   └─> Da: card.FaceUp = true
    │           card.Controller = playerID
    │           return true (Regula 1-B)
    │
    ├─> Verifică Controller == "" ?
    │   └─> Da: card.Controller = playerID
    │           return true (Regula 1-C)
    │
    └─> Altfel: return false (Regula 1-D)
```

---

### 2. FlipSecondCard - Întoarce A Doua Carte

```go
func FlipSecondCard(board *Board, card *Card, row, col int, 
                    playerID string, playerState *PlayerState) bool
```

**Ce face:**

- Implementează **Regulile 2-A, 2-B, 2-C, 2-D, 2-E** din specificație
- Încearcă să întoarcă a doua carte
- Verifică dacă cele două cărți se potrivesc
- Gestionează controlul cărților în funcție de rezultat

**Reguli implementate:**


| Regulă | Situație                   | Acțiune                               | Return  |
| ------- | --------------------------- | -------------------------------------- | ------- |
| 2-A     | Nu există carte            | Renunță la prima carte               | `false` |
| 2-B     | Carte controlată           | Renunță la prima carte               | `false` |
| 2-C     | Carte cu fața în jos      | Întoarce cartea                       | -       |
| 2-D     | Cărți identice (MATCH)    | Păstrează controlul ambelor          | `true`  |
| 2-E     | Cărți diferite (NO MATCH) | Renunță la control, rămân vizibile | `true`  |

**Exemplu de utilizare:**

```go
card := &board.Cards[0][1]

success := FlipSecondCard(board, card, 0, 1, "player1", playerState)

if success && playerState.Matched {
    fmt.Println("MATCH!")
}
```

**Flow-ul funcției:**

```
FlipSecondCard()
    │
    ├─> Verifică Value == "" ?
    │   └─> Da: firstCard.Controller = ""
    │           return false (Regula 2-A)
    │
    ├─> Verifică FaceUp && Controller != "" ?
    │   └─> Da: firstCard.Controller = ""
    │           return false (Regula 2-B)
    │
    ├─> Verifică !FaceUp ?
    │   └─> Da: card.FaceUp = true (Regula 2-C)
    │
    ├─> Compară firstCard.Value == card.Value ?
    │   │
    │   ├─> Da (MATCH):
    │   │   card.Controller = playerID
    │   │   playerState.Matched = true
    │   │   return true (Regula 2-D)
    │   │
    │   └─> Nu (NO MATCH):
    │       firstCard.Controller = ""
    │       card.Controller = ""
    │       playerState.Matched = false
    │       return true (Regula 2-E)
```

---

### 3. CleanupPreviousPlay - Curăță Jocul Anterior

```go
func CleanupPreviousPlay(board *Board, playerState *PlayerState, playerID string)
```

**Ce face:**

- Implementează **Regulile 3-A și 3-B** din specificație
- Se apelează **automat** când jucătorul începe o nouă tură
- Elimină cărțile potrivite sau întoarce cu fața în jos cărțile nepotrivite

**Reguli implementate:**


| Regulă | Situație                           | Acțiune                                           |
| ------- | ----------------------------------- | -------------------------------------------------- |
| 3-A     | Cărți potrivite (Matched=true)    | Elimină ambele cărți (`Value=""`)               |
| 3-B     | Cărți nepotrivite (Matched=false) | Întoarce cu fața în jos (dacă`Controller==""`) |

**Condiții pentru Regula 3-B:**
O carte se întoarce cu fața în jos DOAR dacă:

1. `card.Value != ""` (cartea încă există)
2. `card.FaceUp == true` (e cu fața în sus)
3. `card.Controller == ""` (nimeni nu o controlează)

**Flow-ul funcției:**

```
CleanupPreviousPlay()
    │
    ├─> Verifică HasSecond ?
    │   │
    │   ├─> Verifică Matched ?
    │   │   │
    │   │   ├─> Da (Regula 3-A):
    │   │   │   card.Value = ""
    │   │   │   card.FaceUp = false
    │   │   │   card.Controller = ""
    │   │   │
    │   │   └─> Nu (Regula 3-B):
    │   │       Dacă card.Value != "" && card.FaceUp && card.Controller == "":
    │   │           card.FaceUp = false
    │   │
    │   └─> Dacă doar HasFirst:
    │       Aplică Regula 3-B
    │
    └─> Resetează playerState
```

---

### 4. LoadBoardFromFile - Încarcă Tabla din Fișier

```go
func LoadBoardFromFile(filename string) (*Board, error)
```

**Ce face:**

- Citește fișierul cu configurația tablei
- Parsează dimensiunile (ex: "3x3")
- Creează matricea de cărți
- Returnează un obiect `Board` valid

**Format fișier:**

```
3x3
A
B
C
A
B
C
A
B
C
```

---

### 5. ReplaceCards - Înlocuiește Cărți

```go
func ReplaceCards(board *Board, playerID, fromCard, toCard string) bool
```

**Ce face:**

- Înlocuiește toate cărțile controlate de jucător
- Returnează `true` dacă cel puțin o carte a fost înlocuită

---

## Regulile Jocului

### Regula 1: Prima Carte (First Card)

#### Regula 1-A: Spațiu Gol

**Situație:** Jucătorul încearcă să întoarcă o carte la o poziție unde nu există carte.

```
Înainte:  [ ] (Value="", FaceUp=false, Controller="")

Jucător flip: (0,0)

Rezultat: Operația eșuează
```

**Cod:**

```go
if card.Value == "" {
    log.Printf("Rule 1-A: No card at (%d, %d)", row, col)
    return false
}
```

---

#### Regula 1-B: Carte cu Fața în Jos

**Situație:** Jucătorul întoarce o carte care e cu fața în jos.

```
Înainte:  [?] (Value="A", FaceUp=false, Controller="")

Jucător flip: (0,0)

După:     [A] (Value="A", FaceUp=true, Controller="player1")
```

**Cod:**

```go
if !card.FaceUp {
    card.FaceUp = true
    card.Controller = playerID
    playerState.HasFirst = true
    return true
}
```

---

#### Regula 1-C: Carte Vizibilă, Necontrolată

**Situație:** Cartea e deja cu fața în sus dar nu e controlată de nimeni.

```
Înainte:  [A↑] (Value="A", FaceUp=true, Controller="")

Jucător flip: (0,0)

După:     [A↑] (Value="A", FaceUp=true, Controller="player1")
```

**Cod:**

```go
if card.Controller == "" {
    card.Controller = playerID
    playerState.HasFirst = true
    return true
}
```

---

#### Regula 1-D: Carte Controlată de Altcineva

**Situație:** Cartea e controlată de alt jucător.

```
Înainte:  [A↑] (Value="A", FaceUp=true, Controller="player2")

Jucător1 flip: (0,0)

Rezultat: Operația eșuează
```

---

### Regula 2: A Doua Carte (Second Card)

#### Regula 2-A: Spațiu Gol

```
Jucător are: [A↑] la (0,0), Controller="player1"

Jucător flip: (0,1) unde nu există carte

Rezultat: Operația eșuează
Efect: [A] la (0,0) → Controller=""
```

**Cod:**

```go
if card.Value == "" {
    firstCard.Controller = ""
    playerState.HasFirst = false
    return false
}
```

---

#### Regula 2-B: Carte Controlată

```
Jucător1 are: [A↑] la (0,0), Controller="player1"

Jucător1 flip: (0,1) unde [B↑] Controller="player2"

Rezultat: Operația eșuează
Efect: [A] la (0,0) → Controller=""
```

**Cod:**

```go
if card.FaceUp && card.Controller != "" {
    firstCard.Controller = ""
    playerState.HasFirst = false
    return false
}
```

---

#### Regula 2-C: Întoarce Cartea

```go
if !card.FaceUp {
    card.FaceUp = true
}
```

---

#### Regula 2-D: Match

```
Înainte:
  [A↑] la (0,0), Controller="player1"
  [?]  la (0,1)

Jucător flip: (0,1) → [A]

După:
  [A↑] la (0,0), Controller="player1"
  [A↑] la (0,1), Controller="player1"

Status: MATCH
```

**Cod:**

```go
if firstCard.Value == card.Value {
    card.Controller = playerID
    playerState.Matched = true
    return true
}
```

---

#### Regula 2-E: No Match

```
Înainte:
  [A↑] la (0,0), Controller="player1"
  [?]  la (0,1)

Jucător flip: (0,1) → [B]

După:
  [A↑] la (0,0), Controller=""
  [B↑] la (0,1), Controller=""

Status: NU match
```

**Cod:**

```go
firstCard.Controller = ""
card.Controller = ""
playerState.Matched = false
return true
```

---

### Regula 3: Cleanup (Curățare)

#### Regula 3-A: Elimină Cărțile Potrivite

```
Tura anterioară: [A↑][A↑] MATCH

Jucător începe tură nouă: flip (1,0)

Cleanup: Elimină ambele cărți

Tabla: [ ][ ][?]
       [?][?][?]
       [?][?][?]
```

**Cod:**

```go
if playerState.Matched {
    if card1.Controller == playerID {
        card1.Value = ""
        card1.FaceUp = false
        card1.Controller = ""
    }
}
```

---

#### Regula 3-B: Întoarce Cărțile Nepotrivite

```
Tura anterioară: [A↑][B↑] NO MATCH
Ambele Controller=""

Jucător începe tură nouă: flip (1,0)

Cleanup: Întoarce ambele cu fața în jos

Tabla: [?][?][?]
       [?][?][?]
       [?][?][?]
```

**IMPORTANT:** Cartea NU se întoarce dacă:

- Este controlată de alt jucător
- A fost eliminată

**Cod:**

```go
if card.Value != "" && card.FaceUp && card.Controller == "" {
    card.FaceUp = false
}
```

---

## Thread Safety și Race Conditions

### Problema: Multiple Jucători Concurenți

**Exemplu de race condition FĂRĂ mutex:**

```
Thread 1 (player1):              Thread 2 (player2):
├─> Citește card[0][0]          ├─> Citește card[0][0]
├─> FaceUp = false              ├─> FaceUp = false
├─> Setează FaceUp = true       ├─> Setează FaceUp = true
└─> Setează Controller="p1"     └─> Setează Controller="p2"

Rezultat: Ambii jucători cred că controlează cartea!
```

---

### Soluția: 3 Mutexuri

#### 1. board.mu (RWMutex) - Protejează Cards și version

```go
type Board struct {
    Cards   [][]Card
    version int
    mu      sync.RWMutex
}
```

**RWMutex permite:**

- Multiple citiri simultane (RLock)
- O singură scriere exclusivă (Lock)

**Exemplu citire:**

```go
func (b *Board) FormatBoard(playerID string) string {
    b.mu.RLock()
    defer b.mu.RUnlock()
  
    // Mai multe thread-uri pot citi simultan
    for i := 0; i < b.Rows; i++ {
        card := b.Cards[i][j]
    }
}
```

**Exemplu scriere:**

```go
func handleFlip(...) {
    board.mu.Lock()
    defer board.mu.Unlock()
  
    // Doar acest thread poate modifica
    card := &board.Cards[row][col]
    card.FaceUp = true
    board.version++
}
```

---

#### 2. board.listenersMu (Mutex) - Protejează listeners

```go
func (b *Board) NotifyListeners() {
    b.listenersMu.Lock()
    defer b.listenersMu.Unlock()
  
    for ch := range b.listeners {
        select {
        case ch <- struct{}{}:
        default:
        }
    }
}
```

---

#### 3. board.playerStatesMu (Mutex) - Protejează playerStates

```go
func (b *Board) GetPlayerState(playerID string) *PlayerState {
    b.playerStatesMu.Lock()
    defer b.playerStatesMu.Unlock()
  
    if _, exists := b.playerStates[playerID]; !exists {
        b.playerStates[playerID] = NewPlayerState()
    }
    return b.playerStates[playerID]
}
```

---

### Pattern-ul Defer pentru Unlock

```go
func handleFlip(...) {
    board.mu.Lock()
    defer board.mu.Unlock()  // Se execută ÎNTOTDEAUNA
  
    if error1 {
        return  // Unlock automat
    }
  
    if error2 {
        return  // Unlock automat
    }
  
    return  // Unlock automat
}
```

**Fără defer (periculos):**

```go
func handleFlip(...) {
    board.mu.Lock()
  
    if error {
        return  // UITĂM Unlock → Deadlock
    }
  
    board.mu.Unlock()
}
```

---

### Previne Race Conditions: Exemplu Complet

**Cu mutex (CORECT):**

```go
func handleFlip(...) {
    // Thread 1 obține lock-ul primul
    board.mu.Lock()
  
    card := &board.Cards[row][col]
  
    if !card.FaceUp {
        card.FaceUp = true
        card.Controller = "player1"
    }
  
    board.mu.Unlock()
    // Thread 1 eliberează lock-ul
  
    // ACUM Thread 2 poate obține lock-ul
    board.mu.Lock()
  
    card := &board.Cards[row][col]
  
    // Thread 2 vede că FaceUp == true
    if !card.FaceUp {  // False
        // Nu se execută
    }
  
    board.mu.Unlock()
}
```

---

### Deadlock Prevention

**Ce este un deadlock?**
Două thread-uri așteaptă unul după altul la infinit.

**Cum prevenim?**

- Întotdeauna lock-uim în aceeași ordine
- Folosim `defer` pentru unlock automat
- Ținem lock-urile cât mai puțin timp posibil

**Ordinea corectă:**

```go
1. board.mu (pentru Cards)
2. board.listenersMu (pentru listeners)
3. board.playerStatesMu (pentru playerStates)
```

---

### Testarea Thread Safety: simulate.go

```go
func main() {
    players := []string{"player1", "player2", "player3", "player4"}
  
    // Pornește 4 goroutine-uri simultan
    for _, player := range players {
        go simulatePlayer(player, 100)  // Fiecare face 100 mișcări
    }
}
```

**Ce testează:**

- 4 jucători fac mișcări simultan
- Timeout-uri aleatorii (0.1ms - 2ms)
- 400 request-uri totale
- Verifică că nu crashuiește

---

## API Endpoints

### 1. GET /look/ {playerID}

**Descriere:** Returnează starea tablei.

**Response:**

```
3x3
down
my A
up B
none
```

---

### 2. GET /flip/ {playerID}/ {row}, {col}

**Descriere:** Întoarce o carte.

**Example:** `/flip/player1/0,1`

**Response Success (200):**

```
3x3
my A
my B
```

**Response Failure (409):**

```
Cannot flip that card
```

---

### 3. GET /watch/ {playerID}

**Descriere:** Long polling - așteaptă până se schimbă tabla.

**Comportament:**

- Blochează până când tabla se schimbă SAU
- Timeout 30 secunde SAU
- Client se deconectează

---

### 4. GET /replace/ {playerID}/ {from}/ {to}

**Descriere:** Înlocuiește cărți.

**Example:** `/replace/player1/A/B`

---

## Representation Invariants

### Card Invariants

```go
type Card struct {
    Value      string
    FaceUp     bool
    Controller string
}
```

**Invarianți:**

1. `Value == ""` ⟹ `FaceUp == false` ∧ `Controller == ""`
2. `Controller != ""` ⟹ `FaceUp == true`

```go
func (c *Card) checkRep() {
    if c.Value == "" && (c.FaceUp || c.Controller != "") {
        panic("Eliminated card cannot be face-up or controlled")
    }
    if c.Controller != "" && !c.FaceUp {
        panic("Controlled card must be face-up")
    }
}
```

---

### Board Invariants

**Invarianți:**

1. `Rows > 0` ∧ `Cols > 0`
2. `len(Cards) == Rows`
3. `∀i: len(Cards[i]) == Cols`
4. `version >= 0`
5. Toate cărțile satisfac Card.checkRep()

```go
func (b *Board) checkRep() {
    if b.Rows <= 0 || b.Cols <= 0 {
        panic("Board dimensions must be positive")
    }
    if len(b.Cards) != b.Rows {
        panic("Cards length doesn't match Rows")
    }
    for i := 0; i < b.Rows; i++ {
        if len(b.Cards[i]) != b.Cols {
            panic("Row length doesn't match Cols")
        }
        for j := 0; j < b.Cols; j++ {
            b.Cards[i][j].checkRep()
        }
    }
}
```

---

### PlayerState Invariants

**Invariant:**

- `HasSecond == true` ⟹ `HasFirst == false`

```go
func (p *PlayerState) checkRep() {
    if p.HasSecond && p.HasFirst {
        panic("Player cannot have both HasFirst and HasSecond true")
    }
}
```

---

### Safety from Rep Exposure

**Card - Safe**

- Toate câmpurile sunt primitive (string, bool)
- Nu există referințe mutabile

**Board - Safe**

- Nu returnăm `Cards` direct
- Returnăm doar reprezentare string via `FormatBoard()`
- Toate modificările prin funcții controlate

**PlayerState - Safe**

- Toate câmpurile sunt primitive (int, bool)
- Pointer-ul e gestionat intern de Board

---

## Rezultate Teste

### Output Complet

```
=== RUN   TestFlipEmptySpace
2025/11/13 23:33:07 Rule 1-A: No card at (0, 1)
--- PASS: TestFlipEmptySpace (0.00s)

=== RUN   TestFlipFaceDownCard
2025/11/13 23:33:07 Rule 1-B: Turning up card at (0, 0)
--- PASS: TestFlipFaceDownCard (0.00s)

=== RUN   TestTakeControlOfFaceUpCard
2025/11/13 23:33:07 Rule 1-C: Taking control of card at (0, 0)
--- PASS: TestTakeControlOfFaceUpCard (0.00s)

=== RUN   TestCannotFlipControlledCard
2025/11/13 23:33:07 Rule 1-D: Card at (0, 0) is controlled by player2
--- PASS: TestCannotFlipControlledCard (0.00s)

=== RUN   TestMatchingCards
2025/11/13 23:33:07 Rule 2-C: Turning up card at (0, 1)
2025/11/13 23:33:07 Rule 2-D: Match! A == A
--- PASS: TestMatchingCards (0.00s)

=== RUN   TestNonMatchingCards
2025/11/13 23:33:07 Rule 2-C: Turning up card at (0, 1)
2025/11/13 23:33:07 Rule 2-E: No match. A != B
--- PASS: TestNonMatchingCards (0.00s)

=== RUN   TestRemoveMatchedCards
2025/11/13 23:33:07 Cleaning up previous play for player player1 (HasFirst: false, HasSecond: true, Matched: true)
2025/11/13 23:33:07 Removing matched card at (0, 0)
2025/11/13 23:33:07 Removing matched card at (0, 1)
--- PASS: TestRemoveMatchedCards (0.00s)

=== RUN   TestTurnDownNonMatched
2025/11/13 23:33:07 Cleaning up previous play for player player1 (HasFirst: false, HasSecond: true, Matched: false)
2025/11/13 23:33:07 Turning down card at (0, 0)
2025/11/13 23:33:07 Turning down card at (0, 1)
--- PASS: TestTurnDownNonMatched (0.00s)

=== RUN   TestFlipSecondCardEmptySpace
2025/11/13 23:33:07 Rule 2-A: No card at (0, 1), relinquishing first card
--- PASS: TestFlipSecondCardEmptySpace (0.00s)

=== RUN   TestFlipSecondCardControlled
2025/11/13 23:33:07 Rule 2-B: Card at (0, 1) is controlled, relinquishing first card
--- PASS: TestFlipSecondCardControlled (0.00s)

=== RUN   TestLoadBoardFromFile
--- PASS: TestLoadBoardFromFile (0.00s)

PASS
ok      command-line-arguments  0.002s
```

---

### Analiză Rezultate

Toate cele 11 teste trec cu succes, acoperind complet cele 9 reguli din specificație. Testele verifică atât cazurile de success cât și cazurile de failure pentru fiecare regulă, asigurând corectitudinea implementării. Timpul de execuție de 0.002s demonstrează eficiența implementării.


| Test                         | Regulă | Status |
| ---------------------------- | ------- | ------ |
| TestFlipEmptySpace           | 1-A     | PASS   |
| TestFlipFaceDownCard         | 1-B     | PASS   |
| TestTakeControlOfFaceUpCard  | 1-C     | PASS   |
| TestCannotFlipControlledCard | 1-D     | PASS   |
| TestMatchingCards            | 2-D     | PASS   |
| TestNonMatchingCards         | 2-E     | PASS   |
| TestFlipSecondCardEmptySpace | 2-A     | PASS   |
| TestFlipSecondCardControlled | 2-B     | PASS   |
| TestRemoveMatchedCards       | 3-A     | PASS   |
| TestTurnDownNonMatched       | 3-B     | PASS   |
| TestLoadBoardFromFile        | Loading | PASS   |

---

## Concluzie

Acest proiect implementează un joc multiplayer Memory Scramble complet funcțional în Go, demonstrând principii solide de software construction. Implementarea separă clar logica jocului de infrastructura HTTP, făcând codul ușor de testat și de întreținut. Thread safety-ul este asigurat prin utilizarea a trei mutexuri specializate, iar folosirea pattern-ului defer garantează eliberarea corectă a lock-urilor în toate situațiile.

Toate cele 9 reguli din specificație sunt implementate corect și verificate prin 11 teste unit comprehensive care trec cu succes. Representation invariants sunt definite clar pentru fiecare tip de date și verificate prin funcții checkRep(), asigurând integritatea datelor în timpul execuției. Proiectul demonstrează înțelegere profundă a concurrency în Go, design patterns ADT, și best practices pentru testing.

**Author:** Gurschi Gheorghe

**Data:** 14 Noiembrie 2025

## Comenzi de Rulare

### Inițializare (doar prima dată)

```bash
# Navighează în directorul proiectului
cd memory-scramble

# Inițializează modulul Go
go mod init memory-scramble
```

### Rularea Testelor

```bash
# Rulează toate testele unit cu output detaliat
go test -v board_test.go board.go card.go player.go commands.go

# Sau pentru a vedea doar rezultatul final
go test board_test.go board.go card.go player.go commands.go
```

### Pornirea Serverului

```bash
# Pornește serverul HTTP pe portul 8080
go run board.go card.go player.go commands.go server.go
```

Serverul va porni pe `http://localhost:8080`

```###

```bash
# În alt terminal, după ce serverul pornește
go run simulate.go
```

---

## Structura Proiectului

```
memory-scramble/
├── board.go          # Board ADT - tabla de joc cu toate operațiile
├── card.go           # Card type - structura și validarea unei cărți
├── player.go         # PlayerState - starea unui jucător în joc
├── commands.go       # Logica regulilor jocului (flip, cleanup, replace)
├── server.go         # HTTP server și handler-ele pentru endpoints
├── board_test.go     # Unit tests pentru toate regulile
├── simulate.go       # Script de simulare multi-player
├── index.html        # Client web (interfața jocului)
├── perfect.txt       # Fișierul cu configurația tablei de joc
└── go.mod            # Definiția modulului Go
```

---

## Funcții Principale

### 1. FlipFirstCard - Întoarce Prima Carte

```go
func FlipFirstCard(board *Board, card *Card, row, col int, 
                   playerID string, playerState *PlayerState) bool
```

**Ce face:**

- Implementează **Regulile 1-A, 1-B, 1-C, 1-D** din specificație
- Încearcă să întoarcă prima carte pentru un jucător
- Returnează `true` dacă operația reușește, `false` altfel

**Reguli implementate:**


| Regulă | Situație                        | Acțiune                              | Return  |
| ------- | -------------------------------- | ------------------------------------- | ------- |
| 1-A     | Nu există carte (`Value == ""`) | Eșuează                             | `false` |
| 1-B     | Carte cu fața în jos           | Întoarce cartea, setează controller | `true`  |
| 1-C     | Carte vizibilă, necontrolată   | Preia controlul                       | `true`  |
| 1-D     | Carte controlată de altcineva   | Eșuează (ar aștepta)               | `false` |

**Exemplu de utilizare:**

```go
card := &board.Cards[0][0]
playerState := board.GetPlayerState("player1")

success := FlipFirstCard(board, card, 0, 0, "player1", playerState)

if success {
    fmt.Println("Prima carte întorsă cu succes!")
}
```

**Flow-ul funcției:**

```
FlipFirstCard()
    │
    ├─> Verifică Value == "" ? → return false (Regula 1-A)
    │
    ├─> Verifică !FaceUp ? 
    │   └─> Da: card.FaceUp = true
    │           card.Controller = playerID
    │           return true (Regula 1-B)
    │
    ├─> Verifică Controller == "" ?
    │   └─> Da: card.Controller = playerID
    │           return true (Regula 1-C)
    │
    └─> Altfel: return false (Regula 1-D)
```

---

### 2. FlipSecondCard - Întoarce A Doua Carte

```go
func FlipSecondCard(board *Board, card *Card, row, col int, 
                    playerID string, playerState *PlayerState) bool
```

**Ce face:**

- Implementează **Regulile 2-A, 2-B, 2-C, 2-D, 2-E** din specificație
- Încearcă să întoarcă a doua carte
- Verifică dacă cele două cărți se potrivesc
- Gestionează controlul cărților în funcție de rezultat

**Reguli implementate:**


| Regulă | Situație                   | Acțiune                               | Return  |
| ------- | --------------------------- | -------------------------------------- | ------- |
| 2-A     | Nu există carte            | Renunță la prima carte               | `false` |
| 2-B     | Carte controlată           | Renunță la prima carte               | `false` |
| 2-C     | Carte cu fața în jos      | Întoarce cartea                       | -       |
| 2-D     | Cărți identice (MATCH)    | Păstrează controlul ambelor          | `true`  |
| 2-E     | Cărți diferite (NO MATCH) | Renunță la control, rămân vizibile | `true`  |

**Exemplu de utilizare:**

```go
card := &board.Cards[0][1]

success := FlipSecondCard(board, card, 0, 1, "player1", playerState)

if success && playerState.Matched {
    fmt.Println("MATCH!")
}
```

**Flow-ul funcției:**

```
FlipSecondCard()
    │
    ├─> Verifică Value == "" ?
    │   └─> Da: firstCard.Controller = ""
    │           return false (Regula 2-A)
    │
    ├─> Verifică FaceUp && Controller != "" ?
    │   └─> Da: firstCard.Controller = ""
    │           return false (Regula 2-B)
    │
    ├─> Verifică !FaceUp ?
    │   └─> Da: card.FaceUp = true (Regula 2-C)
    │
    ├─> Compară firstCard.Value == card.Value ?
    │   │
    │   ├─> Da (MATCH):
    │   │   card.Controller = playerID
    │   │   playerState.Matched = true
    │   │   return true (Regula 2-D)
    │   │
    │   └─> Nu (NO MATCH):
    │       firstCard.Controller = ""
    │       card.Controller = ""
    │       playerState.Matched = false
    │       return true (Regula 2-E)
```

---

### 3. CleanupPreviousPlay - Curăță Jocul Anterior

```go
func CleanupPreviousPlay(board *Board, playerState *PlayerState, playerID string)
```

**Ce face:**

- Implementează **Regulile 3-A și 3-B** din specificație
- Se apelează **automat** când jucătorul începe o nouă tură
- Elimină cărțile potrivite sau întoarce cu fața în jos cărțile nepotrivite

**Reguli implementate:**


| Regulă | Situație                           | Acțiune                                           |
| ------- | ----------------------------------- | -------------------------------------------------- |
| 3-A     | Cărți potrivite (Matched=true)    | Elimină ambele cărți (`Value=""`)               |
| 3-B     | Cărți nepotrivite (Matched=false) | Întoarce cu fața în jos (dacă`Controller==""`) |

**Condiții pentru Regula 3-B:**
O carte se întoarce cu fața în jos DOAR dacă:

1. `card.Value != ""` (cartea încă există)
2. `card.FaceUp == true` (e cu fața în sus)
3. `card.Controller == ""` (nimeni nu o controlează)

**Flow-ul funcției:**

```
CleanupPreviousPlay()
    │
    ├─> Verifică HasSecond ?
    │   │
    │   ├─> Verifică Matched ?
    │   │   │
    │   │   ├─> Da (Regula 3-A):
    │   │   │   card.Value = ""
    │   │   │   card.FaceUp = false
    │   │   │   card.Controller = ""
    │   │   │
    │   │   └─> Nu (Regula 3-B):
    │   │       Dacă card.Value != "" && card.FaceUp && card.Controller == "":
    │   │           card.FaceUp = false
    │   │
    │   └─> Dacă doar HasFirst:
    │       Aplică Regula 3-B
    │
    └─> Resetează playerState
```

---

### 4. LoadBoardFromFile - Încarcă Tabla din Fișier

```go
func LoadBoardFromFile(filename string) (*Board, error)
```

**Ce face:**

- Citește fișierul cu configurația tablei
- Parsează dimensiunile (ex: "3x3")
- Creează matricea de cărți
- Returnează un obiect `Board` valid

**Format fișier:**

```
3x3
A
B
C
A
B
C
A
B
C
```

---

### 5. ReplaceCards - Înlocuiește Cărți

```go
func ReplaceCards(board *Board, playerID, fromCard, toCard string) bool
```

**Ce face:**

- Înlocuiește toate cărțile controlate de jucător
- Returnează `true` dacă cel puțin o carte a fost înlocuită

---

## Regulile Jocului

### Regula 1: Prima Carte (First Card)

#### Regula 1-A: Spațiu Gol

**Situație:** Jucătorul încearcă să întoarcă o carte la o poziție unde nu există carte.

```
Înainte:  [ ] (Value="", FaceUp=false, Controller="")

Jucător flip: (0,0)

Rezultat: Operația eșuează
```

**Cod:**

```go
if card.Value == "" {
    log.Printf("Rule 1-A: No card at (%d, %d)", row, col)
    return false
}
```

---

#### Regula 1-B: Carte cu Fața în Jos

**Situație:** Jucătorul întoarce o carte care e cu fața în jos.

```
Înainte:  [?] (Value="A", FaceUp=false, Controller="")

Jucător flip: (0,0)

După:     [A] (Value="A", FaceUp=true, Controller="player1")
```

**Cod:**

```go
if !card.FaceUp {
    card.FaceUp = true
    card.Controller = playerID
    playerState.HasFirst = true
    return true
}
```

---

#### Regula 1-C: Carte Vizibilă, Necontrolată

**Situație:** Cartea e deja cu fața în sus dar nu e controlată de nimeni.

```
Înainte:  [A↑] (Value="A", FaceUp=true, Controller="")

Jucător flip: (0,0)

După:     [A↑] (Value="A", FaceUp=true, Controller="player1")
```

**Cod:**

```go
if card.Controller == "" {
    card.Controller = playerID
    playerState.HasFirst = true
    return true
}
```

---

#### Regula 1-D: Carte Controlată de Altcineva

**Situație:** Cartea e controlată de alt jucător.

```
Înainte:  [A↑] (Value="A", FaceUp=true, Controller="player2")

Jucător1 flip: (0,0)

Rezultat: Operația eșuează
```

---

### Regula 2: A Doua Carte (Second Card)

#### Regula 2-A: Spațiu Gol

```
Jucător are: [A↑] la (0,0), Controller="player1"

Jucător flip: (0,1) unde nu există carte

Rezultat: Operația eșuează
Efect: [A] la (0,0) → Controller=""
```

**Cod:**

```go
if card.Value == "" {
    firstCard.Controller = ""
    playerState.HasFirst = false
    return false
}
```

---

#### Regula 2-B: Carte Controlată

```
Jucător1 are: [A↑] la (0,0), Controller="player1"

Jucător1 flip: (0,1) unde [B↑] Controller="player2"

Rezultat: Operația eșuează
Efect: [A] la (0,0) → Controller=""
```

**Cod:**

```go
if card.FaceUp && card.Controller != "" {
    firstCard.Controller = ""
    playerState.HasFirst = false
    return false
}
```

---

#### Regula 2-C: Întoarce Cartea

```go
if !card.FaceUp {
    card.FaceUp = true
}
```

---

#### Regula 2-D: Match

```
Înainte:
  [A↑] la (0,0), Controller="player1"
  [?]  la (0,1)

Jucător flip: (0,1) → [A]

După:
  [A↑] la (0,0), Controller="player1"
  [A↑] la (0,1), Controller="player1"

Status: MATCH
```

**Cod:**

```go
if firstCard.Value == card.Value {
    card.Controller = playerID
    playerState.Matched = true
    return true
}
```

---

#### Regula 2-E: No Match

```
Înainte:
  [A↑] la (0,0), Controller="player1"
  [?]  la (0,1)

Jucător flip: (0,1) → [B]

După:
  [A↑] la (0,0), Controller=""
  [B↑] la (0,1), Controller=""

Status: NU match
```

**Cod:**

```go
firstCard.Controller = ""
card.Controller = ""
playerState.Matched = false
return true
```

---

### Regula 3: Cleanup (Curățare)

#### Regula 3-A: Elimină Cărțile Potrivite

```
Tura anterioară: [A↑][A↑] MATCH

Jucător începe tură nouă: flip (1,0)

Cleanup: Elimină ambele cărți

Tabla: [ ][ ][?]
       [?][?][?]
       [?][?][?]
```

**Cod:**

```go
if playerState.Matched {
    if card1.Controller == playerID {
        card1.Value = ""
        card1.FaceUp = false
        card1.Controller = ""
    }
}
```

---

#### Regula 3-B: Întoarce Cărțile Nepotrivite

```
Tura anterioară: [A↑][B↑] NO MATCH
Ambele Controller=""

Jucător începe tură nouă: flip (1,0)

Cleanup: Întoarce ambele cu fața în jos

Tabla: [?][?][?]
       [?][?][?]
       [?][?][?]
```

**IMPORTANT:** Cartea NU se întoarce dacă:

- Este controlată de alt jucător
- A fost eliminată

**Cod:**

```go
if card.Value != "" && card.FaceUp && card.Controller == "" {
    card.FaceUp = false
}
```

---

## Thread Safety și Race Conditions

### Problema: Multiple Jucători Concurenți

**Exemplu de race condition FĂRĂ mutex:**

```
Thread 1 (player1):              Thread 2 (player2):
├─> Citește card[0][0]          ├─> Citește card[0][0]
├─> FaceUp = false              ├─> FaceUp = false
├─> Setează FaceUp = true       ├─> Setează FaceUp = true
└─> Setează Controller="p1"     └─> Setează Controller="p2"

Rezultat: Ambii jucători cred că controlează cartea!
```

---

### Soluția: 3 Mutexuri

#### 1. board.mu (RWMutex) - Protejează Cards și version

```go
type Board struct {
    Cards   [][]Card
    version int
    mu      sync.RWMutex
}
```

**RWMutex permite:**

- Multiple citiri simultane (RLock)
- O singură scriere exclusivă (Lock)

**Exemplu citire:**

```go
func (b *Board) FormatBoard(playerID string) string {
    b.mu.RLock()
    defer b.mu.RUnlock()
  
    // Mai multe thread-uri pot citi simultan
    for i := 0; i < b.Rows; i++ {
        card := b.Cards[i][j]
    }
}
```

**Exemplu scriere:**

```go
func handleFlip(...) {
    board.mu.Lock()
    defer board.mu.Unlock()
  
    // Doar acest thread poate modifica
    card := &board.Cards[row][col]
    card.FaceUp = true
    board.version++
}
```

---

#### 2. board.listenersMu (Mutex) - Protejează listeners

```go
func (b *Board) NotifyListeners() {
    b.listenersMu.Lock()
    defer b.listenersMu.Unlock()
  
    for ch := range b.listeners {
        select {
        case ch <- struct{}{}:
        default:
        }
    }
}
```

---

#### 3. board.playerStatesMu (Mutex) - Protejează playerStates

```go
func (b *Board) GetPlayerState(playerID string) *PlayerState {
    b.playerStatesMu.Lock()
    defer b.playerStatesMu.Unlock()
  
    if _, exists := b.playerStates[playerID]; !exists {
        b.playerStates[playerID] = NewPlayerState()
    }
    return b.playerStates[playerID]
}
```

---

### Pattern-ul Defer pentru Unlock

```go
func handleFlip(...) {
    board.mu.Lock()
    defer board.mu.Unlock()  // Se execută ÎNTOTDEAUNA
  
    if error1 {
        return  // Unlock automat
    }
  
    if error2 {
        return  // Unlock automat
    }
  
    return  // Unlock automat
}
```

**Fără defer (periculos):**

```go
func handleFlip(...) {
    board.mu.Lock()
  
    if error {
        return  // UITĂM Unlock → Deadlock
    }
  
    board.mu.Unlock()
}
```

---

### Previne Race Conditions: Exemplu Complet

**Cu mutex (CORECT):**

```go
func handleFlip(...) {
    // Thread 1 obține lock-ul primul
    board.mu.Lock()
  
    card := &board.Cards[row][col]
  
    if !card.FaceUp {
        card.FaceUp = true
        card.Controller = "player1"
    }
  
    board.mu.Unlock()
    // Thread 1 eliberează lock-ul
  
    // ACUM Thread 2 poate obține lock-ul
    board.mu.Lock()
  
    card := &board.Cards[row][col]
  
    // Thread 2 vede că FaceUp == true
    if !card.FaceUp {  // False
        // Nu se execută
    }
  
    board.mu.Unlock()
}
```

---

### Deadlock Prevention

**Ce este un deadlock?**
Două thread-uri așteaptă unul după altul la infinit.

**Cum prevenim?**

- Întotdeauna lock-uim în aceeași ordine
- Folosim `defer` pentru unlock automat
- Ținem lock-urile cât mai puțin timp posibil

**Ordinea corectă:**

```go
1. board.mu (pentru Cards)
2. board.listenersMu (pentru listeners)
3. board.playerStatesMu (pentru playerStates)
```

---

### Testarea Thread Safety: simulate.go

```go
func main() {
    players := []string{"player1", "player2", "player3", "player4"}
  
    // Pornește 4 goroutine-uri simultan
    for _, player := range players {
        go simulatePlayer(player, 100)  // Fiecare face 100 mișcări
    }
}
```

**Ce testează:**

- 4 jucători fac mișcări simultan
- Timeout-uri aleatorii (0.1ms - 2ms)
- 400 request-uri totale
- Verifică că nu crashuiește

---

## API Endpoints

### 1. GET /look/ {playerID}

**Descriere:** Returnează starea tablei.

**Response:**

```
3x3
down
my A
up B
none
```

---

### 2. GET /flip/ {playerID}/ {row}, {col}

**Descriere:** Întoarce o carte.

**Example:** `/flip/player1/0,1`

**Response Success (200):**

```
3x3
my A
my B
```

**Response Failure (409):**

```
Cannot flip that card
```

---

### 3. GET /watch/ {playerID}

**Descriere:** Long polling - așteaptă până se schimbă tabla.

**Comportament:**

- Blochează până când tabla se schimbă SAU
- Timeout 30 secunde SAU
- Client se deconectează

---

### 4. GET /replace/ {playerID}/ {from}/ {to}

**Descriere:** Înlocuiește cărți.

**Example:** `/replace/player1/A/B`

---

## Representation Invariants

### Card Invariants

```go
type Card struct {
    Value      string
    FaceUp     bool
    Controller string
}
```

**Invarianți:**

1. `Value == ""` ⟹ `FaceUp == false` ∧ `Controller == ""`
2. `Controller != ""` ⟹ `FaceUp == true`

```go
func (c *Card) checkRep() {
    if c.Value == "" && (c.FaceUp || c.Controller != "") {
        panic("Eliminated card cannot be face-up or controlled")
    }
    if c.Controller != "" && !c.FaceUp {
        panic("Controlled card must be face-up")
    }
}
```

---

### Board Invariants

**Invarianți:**

1. `Rows > 0` ∧ `Cols > 0`
2. `len(Cards) == Rows`
3. `∀i: len(Cards[i]) == Cols`
4. `version >= 0`
5. Toate cărțile satisfac Card.checkRep()

```go
func (b *Board) checkRep() {
    if b.Rows <= 0 || b.Cols <= 0 {
        panic("Board dimensions must be positive")
    }
    if len(b.Cards) != b.Rows {
        panic("Cards length doesn't match Rows")
    }
    for i := 0; i < b.Rows; i++ {
        if len(b.Cards[i]) != b.Cols {
            panic("Row length doesn't match Cols")
        }
        for j := 0; j < b.Cols; j++ {
            b.Cards[i][j].checkRep()
        }
    }
}
```

---

### PlayerState Invariants

**Invariant:**

- `HasSecond == true` ⟹ `HasFirst == false`

```go
func (p *PlayerState) checkRep() {
    if p.HasSecond && p.HasFirst {
        panic("Player cannot have both HasFirst and HasSecond true")
    }
}
```

---

### Safety from Rep Exposure

**Card - Safe**

- Toate câmpurile sunt primitive (string, bool)
- Nu există referințe mutabile

**Board - Safe**

- Nu returnăm `Cards` direct
- Returnăm doar reprezentare string via `FormatBoard()`
- Toate modificările prin funcții controlate

**PlayerState - Safe**

- Toate câmpurile sunt primitive (int, bool)
- Pointer-ul e gestionat intern de Board

---

## Rezultate Teste

### Output Complet

```
=== RUN   TestFlipEmptySpace
2025/11/13 23:33:07 Rule 1-A: No card at (0, 1)
--- PASS: TestFlipEmptySpace (0.00s)

=== RUN   TestFlipFaceDownCard
2025/11/13 23:33:07 Rule 1-B: Turning up card at (0, 0)
--- PASS: TestFlipFaceDownCard (0.00s)

=== RUN   TestTakeControlOfFaceUpCard
2025/11/13 23:33:07 Rule 1-C: Taking control of card at (0, 0)
--- PASS: TestTakeControlOfFaceUpCard (0.00s)

=== RUN   TestCannotFlipControlledCard
2025/11/13 23:33:07 Rule 1-D: Card at (0, 0) is controlled by player2
--- PASS: TestCannotFlipControlledCard (0.00s)

=== RUN   TestMatchingCards
2025/11/13 23:33:07 Rule 2-C: Turning up card at (0, 1)
2025/11/13 23:33:07 Rule 2-D: Match! A == A
--- PASS: TestMatchingCards (0.00s)

=== RUN   TestNonMatchingCards
2025/11/13 23:33:07 Rule 2-C: Turning up card at (0, 1)
2025/11/13 23:33:07 Rule 2-E: No match. A != B
--- PASS: TestNonMatchingCards (0.00s)

=== RUN   TestRemoveMatchedCards
2025/11/13 23:33:07 Cleaning up previous play for player player1 (HasFirst: false, HasSecond: true, Matched: true)
2025/11/13 23:33:07 Removing matched card at (0, 0)
2025/11/13 23:33:07 Removing matched card at (0, 1)
--- PASS: TestRemoveMatchedCards (0.00s)

=== RUN   TestTurnDownNonMatched
2025/11/13 23:33:07 Cleaning up previous play for player player1 (HasFirst: false, HasSecond: true, Matched: false)
2025/11/13 23:33:07 Turning down card at (0, 0)
2025/11/13 23:33:07 Turning down card at (0, 1)
--- PASS: TestTurnDownNonMatched (0.00s)

=== RUN   TestFlipSecondCardEmptySpace
2025/11/13 23:33:07 Rule 2-A: No card at (0, 1), relinquishing first card
--- PASS: TestFlipSecondCardEmptySpace (0.00s)

=== RUN   TestFlipSecondCardControlled
2025/11/13 23:33:07 Rule 2-B: Card at (0, 1) is controlled, relinquishing first card
--- PASS: TestFlipSecondCardControlled (0.00s)

=== RUN   TestLoadBoardFromFile
--- PASS: TestLoadBoardFromFile (0.00s)

PASS
ok      command-line-arguments  0.002s
```

---

### Analiză Rezultate

Toate cele 11 teste trec cu succes, acoperind complet cele 9 reguli din specificație. Testele verifică atât cazurile de success cât și cazurile de failure pentru fiecare regulă, asigurând corectitudinea implementării. Timpul de execuție de 0.002s demonstrează eficiența implementării.


| Test                         | Regulă | Status |
| ---------------------------- | ------- | ------ |
| TestFlipEmptySpace           | 1-A     | PASS   |
| TestFlipFaceDownCard         | 1-B     | PASS   |
| TestTakeControlOfFaceUpCard  | 1-C     | PASS   |
| TestCannotFlipControlledCard | 1-D     | PASS   |
| TestMatchingCards            | 2-D     | PASS   |
| TestNonMatchingCards         | 2-E     | PASS   |
| TestFlipSecondCardEmptySpace | 2-A     | PASS   |
| TestFlipSecondCardControlled | 2-B     | PASS   |
| TestRemoveMatchedCards       | 3-A     | PASS   |
| TestTurnDownNonMatched       | 3-B     | PASS   |
| TestLoadBoardFromFile        | Loading | PASS   |

---

## Concluzie

Acest proiect implementează un joc multiplayer Memory Scramble complet funcțional în Go, demonstrând principii solide de software construction. Implementarea separă clar logica jocului de infrastructura HTTP, făcând codul ușor de testat și de întreținut. Thread safety-ul este asigurat prin utilizarea a trei mutexuri specializate, iar folosirea pattern-ului defer garantează eliberarea corectă a lock-urilor în toate situațiile.

Toate cele 9 reguli din specificație sunt implementate corect și verificate prin 11 teste unit comprehensive care trec cu succes. Representation invariants sunt definite clar pentru fiecare tip de date și verificate prin funcții checkRep(), asigurând integritatea datelor în timpul execuției. Proiectul demonstrează înțelegere profundă a concurrency în Go, design patterns ADT, și best practices pentru testing.
