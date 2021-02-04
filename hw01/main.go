package main

import (
	"fmt"
)

func main() {
	var n, m int

	fmt.Print("n: ")
	_, _ = fmt.Scanf("%d", &n)

	fmt.Print("m: ")
	_, _ = fmt.Scanf("%d", &m)

	p := findWinner(n, m)

	fmt.Printf("p: %d\n", p)
}

func findWinner(n int, m int) int {
	peopleOut := make(map[int]bool)

	for i:= 1; i <= n; i++ {
		peopleOut[i] = false
	}

	for peopleLeft, i, skipped := len(peopleOut), 1, 0; peopleLeft > 1; {
		if skipped == m && !peopleOut[i] {
			skipped = 0
			peopleOut[i] = true

			peopleLeft--
			i++ // start from next person
		} else if !peopleOut[i] {
			skipped++
			i++
		} else if peopleOut[i] {
			i++
		}

		if i > len(peopleOut) {
			i = 1
		}
	}

	var result int

	for person, visited := range peopleOut {
		if !visited {
			result = person
		}
	}

	return result
}
