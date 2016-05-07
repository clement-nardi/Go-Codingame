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
var minAddScoreToWin int

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
		grids[nbSteps] = strings.Split(fmt.Sprintf("%v %v\n", state.area.potential, state.step+1)+state.area.String(), "\n")
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

func (g *Grid) CellAt(c Coord) byte {
	return g[c.row][c.col]
}

func (g *Grid) ExplodeCellAt(c Coord) {
	/* clear cell */
	g.SetCellAt(c, empty)
	/* clear adjacent skulls */
	around := [4]Coord{Coord{1, 0}, Coord{0, 1}, Coord{-1, 0}, Coord{0, -1}}
	for i := 0; i < 4; i++ {
		c2 := c.Add(around[i])
		if c2.isInGrid() {
			if isSkull(g.CellAt(c2)) {
				g.SetCellAt(c2, empty)
			}
		}
	}
}

func (g *Grid) SetCellAt(c Coord, value byte) {
	g[c.row][c.col] = value
}

func (c Coord) Add(c2 Coord) Coord {
	return Coord{c.row + c2.row, c.col + c2.col}
}

func (c Coord) isInGrid() bool {
	return c.row >= 0 && c.row < nbRows && c.col >= 0 && c.col < nbCols
}

/* builds the list of adjacent cells of the same color */
func (grid *Grid) fourWayExplore(c Coord, treated *Grid, groupCells *[]Coord, mark byte) {
	*groupCells = append(*groupCells, c)
	treated.SetCellAt(c, mark)
	around := [4]Coord{Coord{1, 0}, Coord{0, 1}, Coord{-1, 0}, Coord{0, -1}}
	for i := 0; i < 4; i++ {
		c2 := c.Add(around[i])
		if c2.isInGrid() {
			if treated.CellAt(c2) == 0 && grid.CellAt(c2) == grid.CellAt(c) {
				grid.fourWayExplore(c2, treated, groupCells, mark)
			}
		}
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

func (grid *Grid) dropOneSkullLine() {
	for col := 0; col < nbCols; col++ {
		for row := nbRows - 1; row >= 0; row-- {
			if isEmpty(grid[row][col]) {
				grid[row][col] = skull
				break
			}
		}
	}
}

func (grid *Grid) willLooseForSure() bool {
	largestEmptyArea := 0
	maxPairsDropped := 0

	var treated Grid

	for col := 0; col < nbCols; col++ {
		if treated[0][col] == 0 && isEmpty(grid[0][col]) {
			var group []Coord = make([]Coord, 0, 6)
			grid.fourWayExplore(Coord{0, col}, &treated, &group, 'x')
			if len(group) >= largestEmptyArea {
				largestEmptyArea = len(group)
			}
			/* can only fit 1 pair on 3 empty cells */
			maxPairsDropped += len(group) / 2
		}
	}

	if largestEmptyArea < 4 {
		return true
	}
	if maxPairsDropped > nbPairsKnown {
		return false
	}

	/* TODO: determine if the player can align 4 blocks of the same color
	   with largestEmptyArea/2 pairs from the maxPairDropped next pairs */
	return false
}

func (pa *PlayerArea) resolveAdjacents(dropCoords *[2]Coord, iteration uint) {
	var treated Grid //0 = untreated

	var bigGroups [][]Coord = make([][]Coord, 0, 4)
	var smallGroups [][]Coord = make([][]Coord, 0, 10)
	var mark byte = 'a'

	if dropCoords != nil {
		for _, drop := range dropCoords {
			var group []Coord = make([]Coord, 0, 6)
			pa.grid.fourWayExplore(drop, &treated, &group, mark)
			mark++
			if len(group) >= 4 {
				bigGroups = append(bigGroups, group)
			} else {
				smallGroups = append(smallGroups, group)
			}
		}
	} else {
		/* try all cells */
		for col := 0; col < nbCols; col++ {
			for row := nbRows - 1; row >= 0 && !isEmpty(pa.grid[row][col]); row-- {
				if treated[row][col] == 0 && isColor(pa.grid[row][col]) {
					var group []Coord = make([]Coord, 0, 6)
					pa.grid.fourWayExplore(Coord{row, col}, &treated, &group, mark)
					mark++
					if len(group) >= 4 {
						bigGroups = append(bigGroups, group)
					} else {
						smallGroups = append(smallGroups, group)
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
		B := 0  /* Blocks cleared */
		CP := 0 /* Chain Power */
		CB := 0 /* Color Bonus */
		GB := 0 /* Group Bonus */
		if iteration > 0 {
			CP = 1 << (iteration + 2) // 8, 16, 32, etc. 32 not observed
		}
		var colorCleared [6]bool
		for _, group := range bigGroups {
			colorCleared[pa.grid.CellAt(group[0])-'0'] = true
			for _, coord := range group {
				pa.grid.ExplodeCellAt(coord)
			}
			B += len(group) /* number of blocks cleared, without skulls */
			if len(group) >= 11 {
				GB += 8
			} else {
				GB += len(group) - 4
			}
		}
		nbColorCleared := 0
		for i := 1; i < 6; i++ { //0 is not a color
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
					if col > 0 {
						if currentCell == pa.grid[row][col-1] ||
							currentCell == pa.grid[row+1][col-1] {
							/*  11       1  */
							/*   1  or  11  */
							potential += 2
						}
					}
					if col < nbCols-1 {
						if currentCell == pa.grid[row][col+1] ||
							currentCell == pa.grid[row+1][col+1] {
							/*  11      1   */
							/*  1   or  11  */
							potential += 2
						}
					}
				}
			}
		}

		/* 1 point per vertical groups of same color separated by only one color
		 * (ignore skulls for now) */
		//		var colorStack [nbRows]byte
		//		nbColors := 0
		//		for row := nbRows - 1; row >= 0; row-- {
		//			currentCell := pa.grid[row][col]
		//			if isEmpty(currentCell) {
		//				break
		//			}
		//			if isColor(currentCell) {
		//				if nbColors == 0 || currentCell != colorStack[nbColors-1] {
		//					colorStack[nbColors] = currentCell
		//					nbColors++
		//					if nbColors >= 2 && currentCell == colorStack[nbColors-2] {
		//						potential++
		//					}
		//				}
		//			}
		//		}

		/* 1 point if two groups of the same color are in the same column */
		var colorGroupCount [6]int
		for row := nbRows - 1; row >= 0; row-- {
			currentCell := pa.grid[row][col]
			if isEmpty(currentCell) {
				break
			}
			if isColor(currentCell) {
				if row == nbRows-1 || currentCell != pa.grid[row+1][col] {
					colorGroupCount[currentCell-'0']++
					if colorGroupCount[currentCell-'0'] >= 2 {
						potential += 1
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
				rightCell := pa.grid[row][col+1]
				if currentCell == rightCell {
					potential += 2
					if col < nbCols-2 && currentCell == pa.grid[row][col+2] {
						/* 2 additional points for 3 horizontal cells */
						potential += 2
					}
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
				if currentCell == pa.grid[row+1][col-1] &&
					currentCell != pa.grid[row][col-1] &&
					currentCell != pa.grid[row+1][col] {
					/* x1 */
					/* 1x */
					potential++
				}
				if currentCell == pa.grid[row+1][col+1] &&
					currentCell != pa.grid[row+1][col] &&
					currentCell != pa.grid[row][col+1] {
					/* 1x */
					/* x1 */
					potential++
				}
			}
		}
	}

	/* points if empty columns on the left */
	for col := 0; col < nbCols; col++ {
		row := nbRows - 1
		for row >= 0 && isSkull(pa.grid[row][col]) {
			row--
		}
		if row < 0 || isEmpty(pa.grid[row][col]) {
			potential++
		} else {
			break
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

	next.area.resolveAdjacents(&coords, 0) // will compute the potential

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

func Min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func (s *State) isBetterThan(other *State) bool {
	score1 := s.area.score
	score2 := other.area.score
	if score1 >= minAddScoreToWin && score2 >= minAddScoreToWin {
		if s.step != other.step {
			return s.step < other.step
		} else {
			return score1 > score2
		}
	} else {
		if score1 > 0 && score2 > 0 {
			crit1 := Min(score1, minAddScoreToWin) - 70*nbCols*s.step
			crit2 := Min(score2, minAddScoreToWin) - 70*nbCols*other.step
			if crit1 != crit2 {
				return crit1 > crit2
			} else {
				return s.area.potential > other.area.potential
			}
		} else if score1 != score2 {
			return score1 > score2
		} else {
			/* both scores are 0 */
			return s.area.potential > other.area.potential
		}
	}
}

func (s *State) hasToBeTreatedBefore(other *State) bool {
	return s.area.potential > other.area.potential
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

		ennemyGrid := currentGameArea().playerArea[him].grid
		nbSkullLines := 0

		for !ennemyGrid.willLooseForSure() {
			nbSkullLines++
			ennemyGrid.dropOneSkullLine()
		}

		additionalNuisanceNeededToWin := nbSkullLines*nbCols - currentGameArea().nuisanceAfterSkullDropFrom(me)
		minAddScoreToWin = additionalNuisanceNeededToWin * 70

		fmt.Fprintf(os.Stderr, "minAddScoreToWin=%v\n", minAddScoreToWin)
		//fmt.Fprintln(os.Stderr, currentGameArea())

		var initialState *State = new(State)
		initialState.area = currentGameArea().playerArea[me]
		initialState.area.score = 0

		//fmt.Fprintln(os.Stderr, state)

		var queues [nbPairsKnown + 1]genHeap

		for i := 0; i < nbPairsKnown+1; i++ {
			queues[i] = make(genHeap, 0, 1000)
		}

		queues[0].Push(initialState)

		for i := 0; i < nbPairsKnown+1; i++ {
			heap.Init(&queues[i])
		}

		var workingState *State
		var bestState *State = initialState

		minStep := 0

		for minStep < nbPairsKnown+1 && !timeout {
			for i := minStep; i < nbPairsKnown+1 && !timeout; i++ {
				if len(queues[i]) == 0 {
					if i == minStep {
						minStep++
					}
					continue
				}

				workingState = heap.Pop(&queues[i]).(*State)
				nextStates := workingState.nextStates()
				for _, nextState := range nextStates {
					if i < nbPairsKnown {
						heap.Push(&queues[i+1], nextState)
					}
					if nextState.isBetterThan(bestState) {
						bestState = nextState
					}
				}
			}
		}

		elapsed := time.Since(begin)
		fmt.Fprintf(os.Stderr, "elapsed: %v countNextState=%v\n", elapsed, countNextState)

		fmt.Fprintln(os.Stderr, bestState.Path())
		// fmt.Fprintln(os.Stderr, "Debug messages...")

		/*fmt.Fprintln(os.Stderr, bestState.getNthState(1).areas[me])*/

		nextState := bestState.getNthState(2)
		solutionCol := nextState.area.dropCol
		solutionRot := nextState.area.dropRotation

		debug = true
		if debug && solutionCol >= 0 {
			bestState.getNthState(1).nextState(solutionCol, solutionRot)
		}

		if solutionCol < 0 || solutionCol >= nbCols {
			solutionCol = 0
		}
		if solutionRot < 0 || solutionRot >= 4 {
			solutionRot = 0
		}

		addScore = nextState.area.score - initialState.area.score
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
