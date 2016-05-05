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
	nbRows       = 12
	nbCols       = 6
	nbPairsKnown = 8
	me           = 0
	him          = 1
	empty        = '.'
	skull        = '0'
	nilColumn    = -1
	maxRound     = 200
)

type Line [nbCols]byte
type Grid [nbRows]Line

type PlayerArea struct {
	grid         Grid
	score        int
	dropCol      int //column at which the pair was just dropped
	dropRotation int //rotation at which the pair was just dropped
	potential    int
}

type Pair [2]byte
type PairPipe [nbPairsKnown]Pair

type GameArea struct {
	playerArea [2]PlayerArea
	nextPairs  PairPipe
	previous   *GameArea
}

type Coord struct {
	row, col int
}
type BigGroup struct {
	coord Coord
	count int
}

/* structure used in the search algorithm */
type State struct {
	area     PlayerArea
	step     int //from 0 to 7
	previous *State
}

/***** Global Variables *****/
var gameHistory [maxRound]GameArea
var currentRound int // starts at 0
var countNextState int
var timeout bool
var begin time.Time
var debug bool

func currentGameArea() *GameArea {
	return &gameHistory[currentRound]
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

func (p *PairPipe) acquire() {
	for i := 0; i < nbPairsKnown; i++ {
		var s string
		fmt.Scan(&s)
		p[i][0] = s[0]
		fmt.Scan(&s)
		p[i][1] = s[0]
	}
}

func (ga *GameArea) acquire() {
	ga.nextPairs.acquire()
	ga.playerArea[me].grid.acquire()
	ga.playerArea[him].grid.acquire()
	ga.playerArea[me].dropCol = nilColumn
	if currentRound > 0 {
		ga.previous = &gameHistory[currentRound-1]
	}
}

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

func (pa PlayerArea) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%07d\n", pa.score))
	buffer.WriteString(fmt.Sprintf("%v", pa.grid))
	return buffer.String()
}

func (s *State) Path() string {
	var buffer bytes.Buffer
	var grids [nbPairsKnown + 1][]string
	state := s
	nbSteps := 0
	for state != nil {
		grids[nbSteps] = strings.Split(fmt.Sprintf("%v/%v=%v\n", state.area.potential, state.step+1, state.potentialPerStep())+state.area.String(), "\n")
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

func (ga *GameArea) nuisanceAfterSkullDropFrom(player int) int {
	return (ga.playerArea[player].score % (nbCols * 70)) / 70
}

func (ga *GameArea) nuisanceBeforeSkullDropFrom(player int) int {
	if ga.previous != nil {
		previousScore := ga.previous.playerArea[player].score
		score := ga.playerArea[player].score
		return (score - previousScore + previousScore%(nbCols*70)) / 70
	} else {
		return 0
	}
}

func (ga GameArea) String() string {
	var buffer bytes.Buffer

	myNuisance := ga.nuisanceBeforeSkullDropFrom(him)
	for i := 0; i < nbCols; i++ {
		if nbCols-i-1 < myNuisance {
			buffer.WriteByte(skull)
		} else {
			buffer.WriteByte(' ')
		}
	}
	hisNuisance := ga.nuisanceBeforeSkullDropFrom(me)
	buffer.WriteString(fmt.Sprintf(" %03d             %03d ", myNuisance, hisNuisance))
	for i := 0; i < nbCols; i++ {
		if i < hisNuisance {
			buffer.WriteByte(skull)
		} else {
			buffer.WriteByte(' ')
		}
	}

	buffer.WriteString(fmt.Sprintf("\n%v OLAF69       BADGUY %v\n", ga.playerArea[me].grid[0], ga.playerArea[him].grid[0]))
	buffer.WriteString(fmt.Sprintf("%v %07d  %c  %07d %v\n", ga.playerArea[me].grid[1], ga.playerArea[me].score, ga.nextPairs[0][0], ga.playerArea[him].score, ga.playerArea[him].grid[1]))
	buffer.WriteString(fmt.Sprintf("%v          %c          %v\n", ga.playerArea[me].grid[2], ga.nextPairs[0][1], ga.playerArea[him].grid[2]))

	for i := 0; i < 2; i++ {
		buffer.WriteString(fmt.Sprintf("%v       ", ga.playerArea[me].grid[i+3]))
		for j := 0; j < 7; j++ {
			buffer.WriteString(fmt.Sprintf("%c", ga.nextPairs[j+1][i]))
		}
		buffer.WriteString(fmt.Sprintf("       %v\n", ga.playerArea[him].grid[i+3]))
	}
	for i := 0; i < nbRows-5; i++ {
		buffer.WriteString(fmt.Sprintf("%v                     %v\n", ga.playerArea[me].grid[i+5], ga.playerArea[him].grid[i+5]))
	}
	return buffer.String()
}

func isSkull(c byte) bool {
	return c == skull
}
func isEmpty(c byte) bool {
	return c == empty
}
func isColor(c byte) bool {
	return c >= '1' && c <= '5'
}

/* This function assumes there is an empty cell in the given column */
func (g *Grid) lowestEmpty(col int) int {
	dropDepth := nbRows - 1
	for !isEmpty(g[dropDepth][col]) {
		dropDepth--
	}
	return dropDepth
}

/* if count is nil: clear the adjacent cells of the same color
 * if count is not nil: count the adjacent cells of the same color */
func (grid *Grid) fourWayExplore(row, col int, treated *Grid, count *int, mark byte) {
	treated[row][col] = mark
	around := [4][2]int{{1, 0}, {0, 1}, {-1, 0}, {0, -1}}
	for i := 0; i < 4; i++ {
		row2, col2 := row+around[i][0], col+around[i][1]
		if row2 >= 0 && col2 >= 0 && row2 < nbRows && col2 < nbCols {
			if treated[row2][col2] == 0 && grid[row2][col2] == grid[row][col] {
				grid.fourWayExplore(row2, col2, treated, count, mark)
			}
			if count == nil && isSkull(grid[row2][col2]) {
				grid[row2][col2] = empty
			}
		}
	}
	if count == nil {
		grid[row][col] = empty
	} else {
		(*count)++
	}
}

func (grid *Grid) isIdenticalExceptTopSkullsTo(other *Grid) bool {
	for col := 0; col < nbCols; col++ {
		for row := nbRows - 1; row >= 0; row-- {
			currentCell := grid[row][col]
			otherCell := other[row][col]
			if isEmpty(currentCell) && isEmpty(otherCell) {
				break
			}
			if isSkull(currentCell) && isEmpty(otherCell) ||
				isEmpty(currentCell) && isSkull(otherCell) {
				//ok
			} else if currentCell != otherCell {
				return false
			}
		}
	}
	return true
}

func (pa *PlayerArea) resolveAdjacents(dropHint *[2]Coord, iteration uint) {
	var treated Grid //0 = untreated

	var bigGroups []BigGroup = make([]BigGroup, 0) //row,col,count
	var mark byte = 'a'

	if dropHint != nil {
		for _, coord := range *dropHint {
			count := 0
			pa.grid.fourWayExplore(coord.row, coord.col, &treated, &count, mark)
			mark++
			if count >= 4 {
				bigGroups = append(bigGroups, BigGroup{Coord{coord.row, coord.col}, count})
			}
		}
	} else {
		/* try all cells */
		for col := 0; col < nbCols; col++ {
			for row := nbRows - 1; row >= 0 && !isEmpty(pa.grid[row][col]); row-- {
				if treated[row][col] == 0 && isColor(pa.grid[row][col]) {
					count := 0
					pa.grid.fourWayExplore(row, col, &treated, &count, mark)
					mark++
					if count >= 4 {
						bigGroups = append(bigGroups, BigGroup{Coord{row, col}, count})
					}
				}
			}
		}
	}

	if debug {
		for _, line := range treated {
			for _, cell := range line {
				if cell == 0 {
					fmt.Fprintf(os.Stderr, " ")
				} else {
					fmt.Fprintf(os.Stderr, "%c", cell)
				}
			}
			fmt.Fprintf(os.Stderr, "\n")
		}

		fmt.Fprintf(os.Stderr, "before clear:\n%v", pa)
	}

	if len(bigGroups) > 0 {
		/* now clear the cells from big groups */
		var treated2 Grid //0 = untreated
		B := 0            /* Blocks cleared */
		CP := 0           /* Chain Power */
		CB := 0           /* Color Bonus */
		GB := 0           /* Group Bonus */
		if iteration > 0 {
			CP = 1 << (iteration + 2) // 8, 16, 32, etc. 32 not observed
		}
		var colorCleared [6]bool
		for i := 0; i < len(bigGroups); i++ {
			colorCleared[pa.grid[bigGroups[i].coord.row][bigGroups[i].coord.col]-'0'] = true
			pa.grid.fourWayExplore(bigGroups[i].coord.row, bigGroups[i].coord.col, &treated2, nil, 1)
			B += bigGroups[i].count /* number of blocks cleared, without skulls */
			if bigGroups[i].count >= 11 {
				GB += 8
			} else {
				GB += bigGroups[i].count - 4
			}
		}
		nbColorCleared := 0
		for i := 1; i < 6; i++ {
			if colorCleared[i] {
				nbColorCleared++
			}
		}
		colorBonus := [5]int{0, 2, 4, 8, 16}
		CB = colorBonus[nbColorCleared-1]
		coef := (CP + CB + GB)
		if coef < 1 {
			coef = 1
		} else if coef > 999 {
			coef = 999
		}
		pa.score += (10 * B) * coef

		if debug {
			fmt.Fprintf(os.Stderr, "B=%v CP=%v CB=%v GB=%v coef=%v scoreAdd=%v\n",
				B, CP, CB, GB, coef, (10*B)*coef)
			fmt.Fprintf(os.Stderr, "bg=%v it=%v\n", len(bigGroups), iteration)
			fmt.Fprintf(os.Stderr, "before Drop:\n%v", pa)
		}

		/* let above cells drop */
		for col := 0; col < nbCols; col++ {
			nonEmptyRow := -1
			for row := nbRows - 1; row >= 0; row-- {
				if isEmpty(pa.grid[row][col]) {
					if nonEmptyRow == -1 {
						nonEmptyRow = row - 1
					}
					for nonEmptyRow >= 0 && isEmpty(pa.grid[nonEmptyRow][col]) {
						nonEmptyRow--
					}
					if nonEmptyRow >= 0 {
						pa.grid[row][col] = pa.grid[nonEmptyRow][col]
						pa.grid[nonEmptyRow][col] = empty
					} else {
						break
					}
				}
			}
		}

		if debug {
			fmt.Fprintf(os.Stderr, "after Drop:\n%v", pa)
		}
		/* recursively test for new adjacent colors */
		pa.resolveAdjacents(nil, iteration+1)
	}
}

func (pa *PlayerArea) computePotential() {
	potential := 0

	for col := 0; col < nbCols; col++ {
		/* 2 points per vertically adjacent cells with same color */
		for row := nbRows - 2; row >= 0; row-- {
			currentCell := pa.grid[row][col]
			if isEmpty(currentCell) {
				break
			}
			if isColor(currentCell) {
				belowCell := pa.grid[row+1][col]
				if currentCell == belowCell {
					potential += 2
					if row < nbRows-2 {
						if currentCell == pa.grid[row+2][col] {
							/* vertical alignment of 3 cells of same color: +2 (total = 6) */
							potential += 2
						}
					}
				}
			}
		}

		/* 1 point per vertical groups of same color separated by only one color
		 * (ignore skulls for now) */
		var colorStack [nbRows]byte
		nbColors := 0
		for row := nbRows - 1; row >= 0; row-- {
			currentCell := pa.grid[row][col]
			if isEmpty(currentCell) {
				break
			}
			if isColor(currentCell) {
				if nbColors == 0 || currentCell != colorStack[nbColors-1] {
					colorStack[nbColors] = currentCell
					nbColors++
					if nbColors >= 2 && currentCell == colorStack[nbColors-2] {
						potential++
					}
				}
			}
		}
	}

	/* 2 points per horizontally adjacent cells with same color */
	for col := 0; col < nbCols-1; col++ {
		for row := nbRows - 1; row >= 0; row-- {
			currentCell := pa.grid[row][col]
			if isEmpty(currentCell) {
				break
			}
			if isColor(currentCell) {
				belowCell := pa.grid[row][col+1]
				if currentCell == belowCell {
					potential += 2
				}
			}
		}
	}
	/* 1 points per diagonals (except against edges) */
	for col := 1; col < nbCols-1; col++ {
		for row := nbRows - 2; row >= 0; row-- {
			currentCell := pa.grid[row][col]
			if isEmpty(currentCell) {
				break
			}
			if isColor(currentCell) {
				if currentCell == pa.grid[row+1][col-1] {
					potential++
				}
				if currentCell == pa.grid[row+1][col+1] {
					potential++
				}
			}
		}
	}

	pa.potential = potential
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

func (s *State) nextStates() []*State {
	var nextStates []*State = make([]*State, 0, nbCols)
	if s.step < nbPairsKnown {
		for rot := 0; rot < 4; rot++ {
			firstCol := 0
			lastCol := nbCols
			if rot == 2 {
				firstCol = 1
			}
			if rot == 0 {
				lastCol = nbCols - 1
			}
			for col := firstCol; col < lastCol; col++ {
				/* do they fit? */
				theyFit := true
				switch rot {
				case 0:
					theyFit = isEmpty(s.area.grid[0][col]) && isEmpty(s.area.grid[0][col+1])
				case 2:
					theyFit = isEmpty(s.area.grid[0][col]) && isEmpty(s.area.grid[0][col-1])
				case 1:
					fallthrough
				case 3:
					theyFit = isEmpty(s.area.grid[1][col])
				}

				if theyFit {
					nextStates = append(nextStates, s.nextState(col, rot))
				}
			}
		}
	}
	return nextStates
}

func (s *State) nextState(col, rot int) *State {
	countNextState++
	var next *State = new(State)
	*next = *s
	var coords [2]Coord

	if rot == 1 || rot == 3 {
		dropDepth := next.area.grid.lowestEmpty(col)
		if rot == 1 {
			coords[0] = Coord{dropDepth, col}
			coords[1] = Coord{dropDepth - 1, col}
		} else {
			coords[0] = Coord{dropDepth - 1, col}
			coords[1] = Coord{dropDepth, col}
		}
	} else if rot == 0 {
		coords[0] = Coord{next.area.grid.lowestEmpty(col), col}
		coords[1] = Coord{next.area.grid.lowestEmpty(col + 1), col + 1}
	} else {
		coords[0] = Coord{next.area.grid.lowestEmpty(col), col}
		coords[1] = Coord{next.area.grid.lowestEmpty(col - 1), col - 1}
	}

	/* Drop the two blocks */
	next.area.grid[coords[0].row][coords[0].col] = currentGameArea().nextPairs[next.step][0]
	next.area.grid[coords[1].row][coords[1].col] = currentGameArea().nextPairs[next.step][1]

	next.area.resolveAdjacents(&coords, 0)

	next.step++
	next.area.dropCol = col
	next.area.dropRotation = rot
	next.previous = s
	next.area.computePotential()

	if countNextState%500 == 0 {
		elapsed := time.Since(begin)
		//fmt.Fprintf(os.Stderr, "elapsed: %v countNextState=%v \n", elapsed, countNextState)
		if elapsed > 90*time.Millisecond {
			timeout = true
		}
	}

	return next
}

func (s *State) potentialPerStep() int {
	return s.area.potential / (s.step + 1)
}

func (s *State) isBetterThan(other *State) bool {
	score1 := s.area.score / (s.step + 1)
	score2 := other.area.score / (other.step + 1)
	if score1 != score2 {
		return score1 > score2
	}
	return s.hasMorePotentialPerStepThan(other)
}

func (s *State) hasToBeTreatedBefore(other *State) bool {
	if s.step <= 2 { /* explore all 2-step combinations before going further */
		return true
	}
	return s.hasMorePotentialPerStepThan(other)
}

func (s *State) hasMorePotentialPerStepThan(other *State) bool {
	return s.potentialPerStep() > other.potentialPerStep()
}

type genHeap []*State

func (h genHeap) Len() int            { return len(h) }
func (h genHeap) Less(i, j int) bool  { return h[i].hasToBeTreatedBefore(h[j]) }
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
	currentRound = 0
	addScore := 0
	for {
		timeout = false
		debug = false
		countNextState = 0

		currentGameArea().acquire()
		begin = time.Now()

		if currentGameArea().previous != nil {
			currentGameArea().playerArea[me].score = currentGameArea().previous.playerArea[me].score
			currentGameArea().playerArea[me].score += addScore
			var previousEnnemyState State
			previousEnnemyState.area = currentGameArea().previous.playerArea[him]
			currentRound--
			currenPossibleEnnemyStates := previousEnnemyState.nextStates()
			currentRound++
			for _, currenPossibleEnnemyState := range currenPossibleEnnemyStates {
				//fmt.Fprintf(os.Stderr, "grid #%v\n%v", i, currenPossibleEnnemyState.area.grid)
				if currentGameArea().playerArea[him].grid.isIdenticalExceptTopSkullsTo(&currenPossibleEnnemyState.area.grid) {
					currentGameArea().playerArea[him].score = currenPossibleEnnemyState.area.score
					//fmt.Fprintf(os.Stderr, "Yes! score=%v\n", currenPossibleEnnemyState.area.score)
					break
				}
			}
		}

		//fmt.Fprintln(os.Stderr, currentGameArea())

		var state *State = new(State)

		state.area = currentGameArea().playerArea[me]
		state.area.score = 0

		//fmt.Fprintln(os.Stderr, state)

		states := make(genHeap, 0, 1000)
		states.Push(state)
		heap.Init(&states)

		var workingState *State
		var bestState *State = state

		for len(states) > 0 {
			//fmt.Fprintf(os.Stderr, "Pop %v\n", countPop)
			workingState = heap.Pop(&states).(*State)
			//fmt.Fprintln(os.Stderr, workingState)
			nextStates := workingState.nextStates()
			//fmt.Fprintf(os.Stderr, "next: %v\n", len(nextStates))
			for _, nextState := range nextStates {
				heap.Push(&states, nextState)
				//fmt.Fprintf(os.Stderr, "Push\n")
				if nextState.isBetterThan(bestState) {
					bestState = nextState
				}
			}

			if timeout {
				elapsed := time.Since(begin)
				fmt.Fprintf(os.Stderr, "elapsed: %v countNextState=%v\n", elapsed, countNextState)
				break
			}
		}

		fmt.Fprintln(os.Stderr, bestState.Path())
		// fmt.Fprintln(os.Stderr, "Debug messages...")

		/*fmt.Fprintln(os.Stderr, bestState.getNthState(1).areas[me])*/

		nextState := bestState.getNthState(2)
		solutionCol := nextState.area.dropCol
		solutionRot := nextState.area.dropRotation

		debug = false
		if debug && solutionCol >= 0 {
			bestState.getNthState(1).nextState(solutionCol, solutionRot)
		}

		if solutionCol < 0 || solutionCol >= nbCols {
			solutionCol = 0
		}
		if solutionRot < 0 || solutionRot >= 4 {
			solutionRot = 0
		}

		addScore = nextState.area.score - state.area.score
		var text string
		if addScore > 2000 {
			text = " vvvvvvvvvvvv ARMAGEDDON ^^^^^^^^^^^^"
		} else if addScore > 1000 {
			text = " BULLSEYE!!!!!"
		} else if addScore > 300 {
			text = " Take that!!"
		} else if addScore > 0 {
			text = " Pan!"
		} else {
			if bestState.getNthState(3).area.score > nextState.area.score {
				text = " Wait for it..."
			}
		}

		fmt.Printf("%v %v%s\n", solutionCol, solutionRot, text) // "x": the column in which to drop your blocks
		currentRound++
	}
}
