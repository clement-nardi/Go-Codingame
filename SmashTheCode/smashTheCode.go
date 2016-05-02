package main

import (
	"bytes"
	"container/heap"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	nbRows         = 12
	nbCols         = 6
	pairPipeLength = 8
	me             = 0
	him            = 1
	empty          = '.'
	skull          = '0'
	nilColumn      = -1
)

func isSkull(c byte) bool {
	return c == skull
}
func isEmpty(c byte) bool {
	return c == empty
}
func isColor(c byte) bool {
	return c >= '1' && c <= '5'
}

type Line [nbCols]byte
type Grid [nbRows]Line

func (l Line) String() string {
	return string(l[:nbCols])
}
func (g Grid) String() string {
	var buffer bytes.Buffer
	for i := 0; i < nbRows; i++ {
		buffer.WriteString(fmt.Sprintf("%v\n", g[i]))
	}
	return buffer.String()
}

func (g *Grid) acquire() {
	for i := 0; i < nbRows; i++ {
		var s string
		fmt.Scan(&s)
		for j := 0; j < nbCols; j++ {
			g[i][j] = s[j]
		}
	}
}

type GameArea struct {
	grid     Grid
	score    int
	nuisance int
}

func (g GameArea) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%07d\n", g.score))
	buffer.WriteString(fmt.Sprintf("%v", g.grid))
	return buffer.String()
}

type Pair [2]byte
type PairPipe [pairPipeLength]Pair

func (p *PairPipe) acquire() {
	for i := 0; i < pairPipeLength; i++ {
		var s string
		fmt.Scan(&s)
		p[i][0] = s[0]
		fmt.Scan(&s)
		p[i][1] = s[0]
	}
}

type State struct {
	areas     [2]GameArea
	nextPairs PairPipe
	step      int //from 0 to 7
	dropCol   int
	previous  *State
}

func (s *State) acquire() {
	s.nextPairs.acquire()
	s.areas[me].grid.acquire()
	s.areas[him].grid.acquire()
	s.dropCol = nilColumn
	s.previous = nil
}

/* n >= 1 */
func (s *State) getNthState(n int) *State {
	pathLen := 0
	state := s
	for state != nil {
		state = state.previous
		pathLen++
	}

	if n >= pathLen {
		return s
	}

	state = s
	count := pathLen - n
	for count > 0 {
		count--
		if state.previous != nil {
			state = state.previous
		}
	}
	return state
}

func (s *State) firstDropCol() int {
	state := s
	for state.previous != nil && state.previous.previous != nil {
		state = state.previous
	}

	return state.dropCol
}

func (s *State) Path() string {
	var buffer bytes.Buffer
	var grids [pairPipeLength + 1][]string
	state := s
	nbSteps := 0
	for state != nil {
		grids[nbSteps] = strings.Split(state.areas[me].String(), "\n")
		state = state.previous
		nbSteps++
	}

	for l := range grids[0] {
		for i := 0; i < nbSteps; i++ {
			buffer.WriteString(fmt.Sprintf("%7v|", grids[i][l]))
		}
		buffer.WriteString("\n")
	}
	return buffer.String()
}

func (s State) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%07d s=%v f=%v       %07d\n", s.areas[me].score, s.step, s.firstDropCol(), s.areas[him].score))
	for i := 0; i < 2; i++ {
		buffer.WriteString(fmt.Sprintf("%v        %c        %v\n", s.areas[me].grid[i], s.nextPairs[0][i], s.areas[him].grid[i]))
	}
	for i := 0; i < 2; i++ {
		buffer.WriteString(fmt.Sprintf("%v     ", s.areas[me].grid[i+2]))
		for j := 0; j < 7; j++ {
			buffer.WriteString(fmt.Sprintf("%c", s.nextPairs[j+1][i]))
		}
		buffer.WriteString(fmt.Sprintf("     %v\n", s.areas[him].grid[i+2]))
	}
	for i := 0; i < nbRows-4; i++ {
		buffer.WriteString(fmt.Sprintf("%v                 %v\n", s.areas[me].grid[i+4], s.areas[him].grid[i+4]))
	}
	return buffer.String()
}

func (s *State) nextStates() []*State {
	var nextStates []*State = make([]*State, 0, nbCols)
	if s.step < pairPipeLength {
		for col := 0; col < nbCols; col++ {
			if isEmpty(s.areas[me].grid[1][col]) {
				nextStates = append(nextStates, s.nextState(col))
			}
		}
	}
	return nextStates
}

func (s *State) nextState(col int) *State {
	var next *State = new(State)
	*next = *s

	dropDepth := nbRows - 1
	for ; ; dropDepth-- {
		if isEmpty(next.areas[me].grid[dropDepth][col]) {
			break
		}
	}

	next.areas[me].grid[dropDepth][col] = next.nextPairs[next.step][1]
	next.areas[me].grid[dropDepth-1][col] = next.nextPairs[next.step][0]

	next.areas[me].resolveAdjacents(dropDepth, col, 0)

	next.step++
	next.dropCol = col
	next.previous = s

	return next
}

func (a *GameArea) resolveAdjacents(rowHint, colHint int, iteration uint) {
	var treated Grid //0 = untreated

	for i, line := range treated {
		for j := range line {
			treated[i][j] = ' '
		}
	}
	var bigGroups [][3]int = make([][3]int, 0)
	var mark byte = 'a'

	for col := 0; col < nbCols; col++ {
		for row := nbRows - 1; row >= 0 && !isEmpty(a.grid[row][col]); row-- {
			if treated[row][col] == ' ' && isColor(a.grid[row][col]) {
				count := 0
				a.fourWayExplore(row, col, &treated, &count, mark)
				mark++
				if count >= 4 {
					bigGroups = append(bigGroups, [3]int{row, col, count})
				}
			}
		}
	}

	if debug {
		for _, line := range treated {
			for _, cell := range line {
				fmt.Fprintf(os.Stderr, "%c", cell)
			}
			fmt.Fprintf(os.Stderr, "\n")
		}

		fmt.Fprintf(os.Stderr, "before clear:\n%v", a)
	}

	/* now clear the cells from big groups */

	if len(bigGroups) > 0 {
		for i, line := range treated {
			for j := range line {
				treated[i][j] = ' '
			}
		}
	}

	for i := 0; i < len(bigGroups); i++ {
		a.fourWayExplore(bigGroups[i][0], bigGroups[i][1], &treated, nil, 1)
		base := 10 * bigGroups[i][2]
		coef := 1
		if iteration > 0 {
			coef <<= (iteration + 2) // 8, 16, 32, etc. 32 not observed
		}
		a.score += base * coef
		if i > 0 {
			a.score += 60
		}
	}

	/* let above cells drop */
	if len(bigGroups) > 0 {
		if debug {
			fmt.Fprintf(os.Stderr, "bg=%v it=%v\n", len(bigGroups), iteration)
			fmt.Fprintf(os.Stderr, "before Drop:\n%v", a)
		}

		for col := 0; col < nbCols; col++ {
			nonEmptyRow := -1
			for row := nbRows - 1; row >= 0; row-- {
				if isEmpty(a.grid[row][col]) {
					if nonEmptyRow == -1 {
						nonEmptyRow = row - 1
					}
					for nonEmptyRow >= 0 && isEmpty(a.grid[nonEmptyRow][col]) {
						nonEmptyRow--
					}
					if nonEmptyRow >= 0 {
						a.grid[row][col] = a.grid[nonEmptyRow][col]
						a.grid[nonEmptyRow][col] = empty
					} else {
						break
					}
				}
			}
		}

		if debug {
			fmt.Fprintf(os.Stderr, "after Drop:\n%v", a)
		}
		/* recursively test for new adjacent colors */
		a.resolveAdjacents(-1, -1, iteration+1)
	}
}

/* if count is nil: clear the adjacent cells of the same color
 * if count is not nil: count the adjacent cells of the same color */
func (a *GameArea) fourWayExplore(row, col int, treated *Grid, count *int, mark byte) {
	treated[row][col] = mark
	around := [4][2]int{{1, 0}, {0, 1}, {-1, 0}, {0, -1}}
	for i := 0; i < 4; i++ {
		row2, col2 := row+around[i][0], col+around[i][1]
		if row2 >= 0 && col2 >= 0 && row2 < nbRows && col2 < nbCols {
			if treated[row2][col2] == ' ' && a.grid[row2][col2] == a.grid[row][col] {
				a.fourWayExplore(row2, col2, treated, count, mark)
			}
			if count == nil && isSkull(a.grid[row2][col2]) {
				a.grid[row2][col2] = empty
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

var timeout bool
var begin time.Time
var debug bool

func main() {
	for {
		timeout = false
		debug = false
		begin = time.Now()

		var state *State = new(State)

		state.acquire()

		fmt.Fprintln(os.Stderr, state)

		states := make(genHeap, 0, 1000)
		states.Push(state)
		heap.Init(&states)

		var workingState *State
		var bestState *State = state

		countPop := 0

		for len(states) > 0 {
			//fmt.Fprintf(os.Stderr, "Pop %v\n", countPop)
			workingState = heap.Pop(&states).(*State)
			//fmt.Fprintln(os.Stderr, workingState)
			nextStates := workingState.nextStates()
			//fmt.Fprintf(os.Stderr, "next: %v\n", len(nextStates))
			for _, nextState := range nextStates {
				heap.Push(&states, nextState)
				//fmt.Fprintf(os.Stderr, "Push\n")
			}

			if workingState.isBetterThan(bestState) {
				bestState = workingState
			}

			countPop++
			if countPop%1000 == 0 {
				elapsed := time.Since(begin)
				//fmt.Fprintf(os.Stderr, "elapsed: %v\n", elapsed)
				if elapsed > 95*time.Millisecond {
					timeout = true
				}
			}

			if timeout {
				elapsed := time.Since(begin)
				fmt.Fprintf(os.Stderr, "elapsed: %v countPop=%v\n", elapsed, countPop)
				break
			}
		}

		fmt.Fprintln(os.Stderr, bestState.Path())
		// fmt.Fprintln(os.Stderr, "Debug messages...")

		solution := bestState.firstDropCol()

		/*fmt.Fprintln(os.Stderr, bestState.getNthState(1).areas[me])*/
		debug = true
		if bestState.getNthState(2).dropCol >= 0 {
			bestState.getNthState(1).nextState(bestState.getNthState(2).dropCol)
		}

		if solution < 0 || solution >= nbCols {
			solution = 0
		}

		fmt.Printf("%v\n", solution) // "x": the column in which to drop your blocks
	}
}
