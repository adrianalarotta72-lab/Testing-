package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"
)

type Candidate struct {
	Key   string
	Plain string
	Score float64
}

// -------------------------------
// 1) Vigenère decrypt (A=1, B=2, ..., Z=26)
// -------------------------------

// keyShifts convierte "abc" -> [1,2,3] (según A=1, B=2)
func keyShifts(key string) []int {
	key = strings.ToLower(key)
	shifts := make([]int, 0, len(key))
	for _, r := range key {
		if r >= 'a' && r <= 'z' {
			// A=1..Z=26
			shifts = append(shifts, int(r-'a')+1)
		}
	}
	return shifts
}

func decryptVigenere(cipher string, key string) string {
	shifts := keyShifts(key)
	if len(shifts) == 0 {
		return ""
	}

	var out strings.Builder
	out.Grow(len(cipher))

	k := 0 // índice de la clave (solo avanza cuando procesamos letras)

	for _, r := range cipher {
		// Solo desciframos letras. El resto queda igual.
		if unicode.IsLetter(r) {
			low := unicode.ToLower(r)
			if low >= 'a' && low <= 'z' {
				shift := shifts[k%len(shifts)]
				k++

				// Convertimos a 0..25
				x := int(low - 'a')
				// Descifrar: restar shift (pero shift es 1..26)
				// Ajuste modular para quedar en 0..25:
				y := (x - shift) % 26
				if y < 0 {
					y += 26
				}
				out.WriteRune(rune('a' + y))
				continue
			}
		}

		// Si no es letra a-z, la copiamos tal cual
		out.WriteRune(r)
	}

	return out.String()
}

// -------------------------------
// 2) Scoring “parece inglés”
// -------------------------------

var commonWords = []string{
	" the ", " and ", " to ", " of ", " in ", " is ", " that ", " for ", " you ", " it ",
}

var commonBigrams = []string{
	"th", "he", "in", "er", "an", "re", "on", "at", "en", "nd",
}

// scoreEnglish es un scoring sencillo pero efectivo:
// - recompensa palabras comunes
// - recompensa bigramas frecuentes
// - penaliza “basura” (muchos símbolos raros)
func scoreEnglish(s string) float64 {
	txt := " " + strings.ToLower(s) + " "

	score := 0.0

	// 1) Palabras comunes (fuerte señal)
	for _, w := range commonWords {
		score += float64(strings.Count(txt, w)) * 5.0
	}

	// 2) Bigramas comunes (señal media)
	onlyLetters := make([]rune, 0, len(txt))
	for _, r := range txt {
		if r >= 'a' && r <= 'z' {
			onlyLetters = append(onlyLetters, r)
		}
	}
	letterStr := string(onlyLetters)
	for _, bg := range commonBigrams {
		score += float64(strings.Count(letterStr, bg)) * 0.5
	}

	// 3) Penalización por caracteres raros (señal anti-basura)
	weird := 0
	total := 0
	for _, r := range txt {
		total++
		if (r >= 'a' && r <= 'z') || r == ' ' || r == '.' || r == ',' || r == ':' || r == ';' || r == '\'' || r == '-' {
			continue
		}
		// números u otros símbolos
		weird++
	}
	if total > 0 {
		score -= float64(weird) / float64(total) * 20.0
	}

	return score
}

// -------------------------------
// 3) Leer diccionario y resolver Stage 1
// -------------------------------

func loadDictionary(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	words := make([]string, 0, 50000)
	for sc.Scan() {
		w := strings.TrimSpace(sc.Text())
		if w == "" {
			continue
		}
		// Normalizamos: solo letras (para que la clave tenga shifts)
		w = strings.ToLower(w)
		ok := true
		for _, r := range w {
			if r < 'a' || r > 'z' {
				ok = false
				break
			}
		}
		if ok {
			words = append(words, w)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return words, nil
}

func solveStage1(cipher string, dict []string, topK int) []Candidate {
	cands := make([]Candidate, 0, topK)

	// Guardamos “mejores K” ordenados por score ascendente internamente (min-heap manual sencillo)
	pushCandidate := func(c Candidate) {
		if len(cands) < topK {
			cands = append(cands, c)
			sort.Slice(cands, func(i, j int) bool { return cands[i].Score < cands[j].Score })
			return
		}
		// si el nuevo es mejor que el peor (cands[0] es el menor score porque orden asc)
		if c.Score <= cands[0].Score {
			return
		}
		cands[0] = c
		sort.Slice(cands, func(i, j int) bool { return cands[i].Score < cands[j].Score })
	}

	for _, key := range dict {
		plain := decryptVigenere(cipher, key)
		if plain == "" {
			continue
		}
		s := scoreEnglish(plain)
		pushCandidate(Candidate{Key: key, Plain: plain, Score: s})
	}

	// devolvemos en orden descendente (mejor primero)
	sort.Slice(cands, func(i, j int) bool { return cands[i].Score > cands[j].Score })
	return cands
}

func main() {
	dictPath := flag.String("dict", "dictionary.txt", "ruta a dictionary.txt")
	msg := flag.String("msg", "", "mensaje cifrado (entre comillas)")
	topK := flag.Int("top", 3, "cuántos candidatos mostrar")
	flag.Parse()

	if strings.TrimSpace(*msg) == "" {
		fmt.Println("Pasa el mensaje cifrado con -msg=\"...\"")
		os.Exit(1)
	}

	dict, err := loadDictionary(*dictPath)
	if err != nil {
		panic(err)
	}

	best := solveStage1(*msg, dict, *topK)
	if len(best) == 0 {
		fmt.Println("No se encontraron candidatos.")
		return
	}

	for i, c := range best {
		fmt.Printf("\n#%d  key=%%q  score=%%.2f\n", i+1, c.Key, c.Score)
		fmt.Println(c.Plain)
	}
}