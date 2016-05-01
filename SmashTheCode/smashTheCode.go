package main

import "fmt"

//import "os"
const (
	nbRows = 12
	nbCols = 6
	me     = 0
	him    = 1
	empty  = '.'
	skull  = '0'
)

type Cell byte

func (c Cell) isSkull() bool {
	return c == skull
}
func (c Cell) isEmpty() bool {
	return c == empty
}

type Grid [nbRows][nbCols]Cell

type GameArea struct {
	grid     Grid
	score    int
	nuisance int
}

type Pair [2]Cell

type State struct {
	areas     [2]GameArea
	nextPairs [8]Pair
	step      int //from 0 to 7
	firstCol  int
}

func (s *State) nextStates() []*State {
	var nextStates []*State = make([]*State, 0, nbCols)
	for col := 0; col < nbCols; col++ {
		if !s.areas[me].grid[1][col].isEmpty() {
			nextStates = append(nextStates, s.nextState(col))
		}
	}
	return nextStates
}

func (s *State) nextState(col int) *State {
	var next *State = new(State)
	*next = *s

	dropDepth := nbRows - 1
	for ; ; dropDepth-- {
		if next.areas[me].grid[dropDepth][col].isEmpty() {
			break
		}
	}

	next.areas[me].grid[dropDepth][col] = next.nextPairs[next.step][1]
	next.areas[me].grid[dropDepth-1][col] = next.nextPairs[next.step][0]

	next.areas[me].resolveAdjacents(dropDepth, col)

	next.step++
	return next
}

func (a *GameArea) resolveAdjacents(rowHint, colHint int) {
	var treated Grid //0 = untreated

	var bigGroups [][3]int = make([][3]int, 0)

	for col := 0; col < nbCols; col++ {
		for row := nbRows - 1; !a.grid[row][col].isEmpty() && row >= 0; row-- {
			if treated[row][col] == 0 {
				count := 0
				a.fourWayExplore(row, col, &treated, &count)
				if count >= 4 {
					bigGroups = append(bigGroups, [3]int{row, col, count})
				}
			}
		}
	}

	/* now clear the cells from big groups */
	var treated2 Grid
	for i := 0; i < len(bigGroups); i++ {
		a.fourWayExplore(bigGroups[i][0], bigGroups[i][1], &treated2, nil)
		a.score += 10 * bigGroups[i][2]
	}

	/* let above cells drop */
	if len(bigGroups) > 0 {
		for col := 0; col < nbCols; col++ {
			nonEmptyRow := -1
			for row := nbRows - 1; row >= 0; row-- {
				if a.grid[row][col].isEmpty() {
					if nonEmptyRow == -1 {
						nonEmptyRow = row - 1
					}
					for nonEmptyRow >= 0 && a.grid[nonEmptyRow][col].isEmpty() {
						nonEmptyRow--
					}
					if nonEmptyRow >= 0 {
						a.grid[row][col] = a.grid[nonEmptyRow][col]
					} else {
						break
					}
				}
			}
		}

		/* recursively test for new adjacent colors */
		a.resolveAdjacents(-1, -1)
	}
}

/* if count is nil: clear the adjacent cells of the same color
 * if count is not nil: count the adjacent cells of the same color */
func (a *GameArea) fourWayExplore(row, col int, treated *Grid, count *int) {
	treated[row][col] = 1
	around := [4][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	for i := 0; i < 4; i++ {
		row2, col2 := row+around[i][0], col+around[i][1]
		if row2 > 0 && col2 > 0 && row2 < nbRows && col2 < nbCols {
			if treated[row2][col2] == 0 && a.grid[row2][col2] == a.grid[row][col] {
				a.fourWayExplore(row2, col2, treated, count)
			}
		}
	}
	if count == nil {
		a.grid[row][col] = empty
	} else {
		(*count)++
	}
}
func (s *State) isBetterThan(other *State) bool {
	return s.areas[me].score > other.areas[me].score
}

type genHeap []*State

func (h genHeap) Len() int            { return len(h) }
func (h genHeap) Less(i, j int) bool  { return h[i].isBetterThan(h[j]) }
func (h genHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *genHeap) Push(x interface{}) { *h = append(*h, x.(*State)) }
func (h *genHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func main() {
	col := 0
	for {

		for i := 0; i < 8; i++ {
			// colorA: color of the first block
			// colorB: color of the attached block
			var colorA, colorB int
			fmt.Scan(&colorA, &colorB)
		}
		for i := 0; i < 12; i++ {
			var row string
			fmt.Scan(&row)
		}
		for i := 0; i < 12; i++ {
			// row: One line of the map ('.' = empty, '0' = skull block, '1' to '5' = colored block)
			var row string
			fmt.Scan(&row)
		}

		// fmt.Fprintln(os.Stderr, "Debug messages...")
		col = (col + 1) % 6
		fmt.Printf("%v\n", col) // "x": the column in which to drop your blocks
	}
}
