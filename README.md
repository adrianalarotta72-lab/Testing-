package main

import (
	"bufio"
	"container/heap"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Odds struct {
	home float64
	draw float64
	away float64
}

// Guardamos inv=1/odd y el índice del bookie para poder exigir bookies distintos.
type InvOdd struct {
	inv float64
	idx int
}

/*
Por qué round2 así:
- Task 1 pide redondear cada margen a 2 decimales ANTES de sumar.
- Usamos “round half up” clásico: floor(x*100+0.5)/100
*/
func round2(x float64) float64 {
	return math.Floor(x*100.0+0.5) / 100.0
}

/*
Lee el archivo: cada línea es "BookieX home draw away"
Por qué parsear así:
- necesitamos las cuotas de cada bookie para cada resultado.
*/
func readOddsFile(path string) ([]Odds, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	// buffer por si hay líneas largas (no suele, pero seguro)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var out []Odds
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 4 {
			return nil, fmt.Errorf("linea inválida: %q", line)
		}
		h, e1 := strconv.ParseFloat(parts[1], 64)
		d, e2 := strconv.ParseFloat(parts[2], 64)
		a, e3 := strconv.ParseFloat(parts[3], 64)
		if e1 != nil || e2 != nil || e3 != nil {
			return nil, fmt.Errorf("parse error: %q", line)
		}
		out = append(out, Odds{home: h, draw: d, away: a})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

/*
Prepara invH, invD, invA.
Por qué:
- condición de arbitraje es invH + invD + invA < 1
- ordenamos invD e invA para poder usar two-pointer y evitar O(n^3)
- guardamos idx para filtrar “bookies distintos”
*/
func prepareInv(rows []Odds) (invH, invD, invA []InvOdd) {
	n := len(rows)
	invH = make([]InvOdd, 0, n)
	invD = make([]InvOdd, 0, n)
	invA = make([]InvOdd, 0, n)

	for i, r := range rows {
		invH = append(invH, InvOdd{inv: 1.0 / r.home, idx: i})
		invD = append(invD, InvOdd{inv: 1.0 / r.draw, idx: i})
		invA = append(invA, InvOdd{inv: 1.0 / r.away, idx: i})
	}

	// Ordenamos solo D y A; H puede ir como venga.
	sort.Slice(invD, func(i, j int) bool { return invD[i].inv < invD[j].inv })
	sort.Slice(invA, func(i, j int) bool { return invA[i].inv < invA[j].inv })
	return
}

/*
Task 2: Highest Profit
Por qué es tan simple:
- el mejor arbitraje ocurre usando la mejor cuota (odd más alto) para cada outcome
- odd más alto => inv más bajo => suma inv más pequeña => margen más grande
- NO se pide redondeo (solutions muestra float completo)
*/
func task2BestMargin(rows []Odds) float64 {
	bestH, bestD, bestA := 0.0, 0.0, 0.0
	for _, r := range rows {
		if r.home > bestH {
			bestH = r.home
		}
		if r.draw > bestD {
			bestD = r.draw
		}
		if r.away > bestA {
			bestA = r.away
		}
	}
	invSum := 1.0/bestH + 1.0/bestD + 1.0/bestA
	if invSum >= 1.0 {
		return 0.0
	}
	return 100.0 * (1.0 - invSum)
}

/*
Task 1: Identify ALL Arbitrages
Por qué two-pointer:
- invD e invA ordenados
- para cada (h,d) calculamos t = 1 - h - d
- todos los invA < t son válidos
- idx es el “corte”: invA[0:idx] < t
- idx solo disminuye, entonces es eficiente.

OJO: aquí sí contamos TODOS y sumamos margen redondeado por combinación.
*/
func task1AllArbsSum(rows []Odds) (count int64, sum float64) {
	invH, invD, invA := prepareInv(rows)
	if len(invH) == 0 || len(invD) == 0 || len(invA) == 0 {
		return 0, 0
	}

	minInvA := invA[0].inv
	nA := len(invA)

	for _, h := range invH {
		idx := nA // invA[0:idx] son < t
		for _, d := range invD {
			t := 1.0 - h.inv - d.inv
			if t <= minInvA {
				// como invD crece, con d más grande t solo baja; ya no habrá válidos
				break
			}

			// movemos idx hacia abajo hasta que invA[idx-1] < t
			for idx > 0 && invA[idx-1].inv >= t {
				idx--
			}
			if idx == 0 {
				break
			}

			// base = margen si invA fuera 0, luego restamos 100*invA[k]
			base := 100.0 * (1.0 - h.inv - d.inv)

			for k := 0; k < idx; k++ {
				a := invA[k]

				// regla del judge: 3 bookies distintos
				if h.idx == d.idx || h.idx == a.idx || d.idx == a.idx {
					continue
				}

				m := base - 100.0*a.inv
				if m > 0 {
					sum += round2(m)
					count++
				}
			}
		}
	}

	return count, sum
}

/* -------- Heap para Top-K (Task 3) -------- */

type MinHeap []float64

func (h MinHeap) Len() int            { return len(h) }
func (h MinHeap) Less(i, j int) bool  { return h[i] < h[j] } // min-heap
func (h MinHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *MinHeap) Push(x interface{}) { *h = append(*h, x.(float64)) }
func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

/*
Task 3: Top K arbitrages (márgenes) y sumar.
Por qué heap:
- En big puede haber millones de arbitrajes.
- Guardar todos y ordenar consume memoria enorme.
- Heap tamaño K mantiene solo los K mejores en streaming.
*/
func task3TopKSum(rows []Odds, K int) float64 {
	if K <= 0 {
		return 0
	}

	invH, invD, invA := prepareInv(rows)
	if len(invH) == 0 || len(invD) == 0 || len(invA) == 0 {
		return 0
	}

	minInvA := invA[0].inv
	nA := len(invA)

	hh := &MinHeap{}
	heap.Init(hh)

	for _, h := range invH {
		idx := nA
		for _, d := range invD {
			t := 1.0 - h.inv - d.inv
			if t <= minInvA {
				break
			}
			for idx > 0 && invA[idx-1].inv >= t {
				idx--
			}
			if idx == 0 {
				break
			}

			base := 100.0 * (1.0 - h.inv - d.inv)

			for k := 0; k < idx; k++ {
				a := invA[k]

				if h.idx == d.idx || h.idx == a.idx || d.idx == a.idx {
					continue
				}

				m := base - 100.0*a.inv
				if m <= 0 {
					continue
				}
				m = round2(m)

				if hh.Len() < K {
					heap.Push(hh, m)
				} else if m > (*hh)[0] {
					(*hh)[0] = m
					heap.Fix(hh, 0)
				}
			}
		}
	}

	sum := 0.0
	for hh.Len() > 0 {
		sum += heap.Pop(hh).(float64)
	}
	return sum
}

func main() {
	task := flag.Int("task", 1, "task number: 1, 2, or 3")
	k := flag.Int("k", 0, "K for task 3")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage:")
		fmt.Println("  go run main.go -task=1 task1_small.txt")
		fmt.Println("  go run main.go -task=2 task2_small.txt")
		fmt.Println("  go run main.go -task=3 -k=3 task3_small.txt")
		os.Exit(1)
	}

	file := flag.Arg(0)
	rows, err := readOddsFile(file)
	if err != nil {
		panic(err)
	}

	switch *task {
	case 1:
		count, sum := task1AllArbsSum(rows)
		fmt.Printf("Length: %d\n", count)
		fmt.Printf("Profit Percentage: %.15g\n", sum)

	case 2:
		best := task2BestMargin(rows)
		fmt.Printf("Profit Percentage: %.17g\n", best)

	case 3:
		if *k <= 0 {
			fmt.Println("Task 3 requires -k, e.g. -k=3 or -k=100")
			os.Exit(1)
		}
		sum := task3TopKSum(rows, *k)
		fmt.Printf("Profit Percentage: %.15g\n", sum)

	default:
		fmt.Println("Invalid task. Use 1, 2, or 3.")
	}
}
