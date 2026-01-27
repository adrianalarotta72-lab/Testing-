package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// Stage 1: tamaños fijos
var sizes = []int{5, 16, 42, 59}

// minBoxesForN devuelve (minBoxes, ok)
// Encuentra la mínima cantidad de cajas para sumar EXACTO N usando {5,16,42,59}.
//
// Estrategia rápida (sin DP gigante):
// - Enumeramos a = #59, b = #42 (rangos pequeños porque son grandes)
// - Para el resto R = N - 59a - 42b, resolvemos con 16 y 5:
//      R - 16c debe ser divisible por 5, y d = (R-16c)/5 >= 0
func minBoxesForN(N int) (int, bool) {
	if N < 0 {
		return 0, false
	}
	if N == 0 {
		return 0, true
	}

	const INF = int(^uint(0) >> 1)
	best := INF
	found := false

	maxA := N / 59
	for a := 0; a <= maxA; a++ {
		R1 := N - 59*a
		if R1 < 0 {
			break
		}

		maxB := R1 / 42
		for b := 0; b <= maxB; b++ {
			R := R1 - 42*b
			if R < 0 {
				break
			}

			// Ahora resolver R con 16 y 5:
			// buscamos c >= 0 tal que:
			//   R - 16c >= 0 y (R - 16c) % 5 == 0
			// Para minimizar cajas, conviene maximizar c (usar más 16 y menos 5),
			// porque 16 "cubre más" por caja que 5.
			maxC := R / 16

			// En vez de iterar todos los c, buscamos el primero (más grande) que cumpla el mod.
			// Necesitamos: (R - 16c) % 5 == 0
			// 16 mod 5 = 1, entonces:
			// (R - c) % 5 == 0  =>  c % 5 == R % 5
			wantMod := R % 5

			// Tomamos c = maxC ajustado hacia abajo para que c%5 == wantMod
			c := maxC - ((maxC-wantMod+5)%5)
			// Nota: la fórmula de arriba asegura que c%5 == wantMod, si existe >=0.

			if c < 0 {
				continue
			}

			rest := R - 16*c
			if rest < 0 || rest%5 != 0 {
				// por seguridad (normalmente no entra)
				continue
			}
			d := rest / 5

			totalBoxes := a + b + c + d
			if totalBoxes < best {
				best = totalBoxes
				found = true
			}
		}
	}

	return best, found
}

type job struct {
	lineNo int
	N      int
}

type result struct {
	lineNo   int
	minBoxes int
	ok       bool
}

// worker: consume jobs, compute minBoxesForN, send result
func worker(jobs <-chan job, results chan<- result) {
	for j := range jobs {
		minB, ok := minBoxesForN(j.N)
		results <- result{lineNo: j.lineNo, minBoxes: minB, ok: ok}
	}
}

func main() {
	// Uso:
	//   go run main.go input1.txt
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run main.go input1.txt")
		os.Exit(1)
	}
	path := os.Args[1]

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Worker pool
	numWorkers := runtime.NumCPU() // buena base
	jobs := make(chan job, 4096)
	results := make(chan result, 4096)

	for w := 0; w < numWorkers; w++ {
		go worker(jobs, results)
	}

	// 1) Productor: leer archivo y mandar jobs
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	lineNo := 0
	sent := 0

	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}

		n64, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			fmt.Printf("Línea %d inválida: %q\n", lineNo, line)
			os.Exit(1)
		}
		if n64 < 0 || n64 > int64(^uint(0)>>1) {
			fmt.Printf("Línea %d fuera de rango int: %d\n", lineNo, n64)
			os.Exit(1)
		}
		N := int(n64)

		jobs <- job{lineNo: lineNo, N: N}
		sent++
	}
	if err := sc.Err(); err != nil {
		panic(err)
	}
	close(jobs)

	// 2) Consumidor: recibir exactamente "sent" resultados y sumar
	totalBoxes := 0
	for i := 0; i < sent; i++ {
		r := <-results
		if !r.ok {
			// Stage 1: si no se puede empacar exactamente, lo reportamos.
			// Si quieres ignorar, cambia esto a: continue
			fmt.Printf("No se puede empacar exactamente (línea %d)\n", r.lineNo)
			os.Exit(1)
		}
		totalBoxes += r.minBoxes
	}

	fmt.Printf("Total boxes: %d\n", totalBoxes)

	_ = sizes // solo para dejar claro que los tamaños están fijos arriba
}