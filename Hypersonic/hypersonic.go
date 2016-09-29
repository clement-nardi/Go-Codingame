package main

import (
	"bytes"
	"fmt"
	"os"
)

//import "os"
//import "bufio"

//import "strings"
//import "strconv"

const (
	playerEntity         = 0
	bombEntity           = 1
	itemEntity           = 2
	nbCols               = 13
	nbRows               = 11
	box                  = '0'
	empty                = '.'
	nbCardinalDirections = 4
	bombTimer            = 8
)

var cardinalVectors [nbCardinalDirections]Position = [nbCardinalDirections]Position{Position{0, 1}, Position{0, -1}, Position{1, 0}, Position{-1, 0}}

func Min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
func Max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

type Position struct {
	x, y int
}

func (p Position) String() string {
	return fmt.Sprintf("%v %v", p.x, p.y)
}
func (p Position) IsOnGrid() bool {
	return p.x >= 0 && p.x < nbCols && p.y >= 0 && p.y < nbRows
}
func Add(a, b Position) Position {
	return Position{a.x + b.x, a.y + b.y}
}

type Player struct {
	Position
	remainingBombs int
	bombRange      int
}

type Bomb struct {
	Position
	ownerID             int
	nbRoundsToExplosion int
	explosionRange      int
}

type Item struct {
	Position
	itemType int
}

type Line [nbCols]byte
type Grid [nbRows]Line

func (g *Grid) acquire() {
	for i := 0; i < nbRows; i++ {
		var s string
		fmt.Scan(&s)
		for j := 0; j < nbCols; j++ {
			g[i][j] = s[j]
		}
	}
}

func isBox(b byte) bool {
	return b == box
}
func isEmpty(b byte) bool {
	return b == empty
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

func (g Grid) CellAt(p Position) byte {
	return g[p.y][p.x]
}
func (g *Grid) SetCellAt(p Position, b byte) {
	g[p.y][p.x] = b
}

type GameArea struct {
	grid    Grid
	bombs   []Bomb
	players [2]Player
	items   []Item
}

func (ga *GameArea) acquire() {
	ga.grid.acquire()

	var nbEntities int
	fmt.Scan(&nbEntities)

	ga.bombs = make([]Bomb, 0)
	ga.items = make([]Item, 0)

	for i := 0; i < nbEntities; i++ {
		var entityType, owner, x, y, param1, param2 int
		fmt.Scan(&entityType, &owner, &x, &y, &param1, &param2)
		switch entityType {
		case playerEntity:
			ga.players[owner] = Player{Position{x, y}, param1, param2}
		case bombEntity:
			var bomb = Bomb{Position{x, y}, owner, param1, param2}
			ga.bombs = append(ga.bombs, bomb)
		case itemEntity:
			ga.items = append(ga.items, Item{Position{x, y}, param1})
		}
	}
}

func (ga GameArea) HasBombAt(p Position) bool {
	for _, bomb := range ga.bombs {
		if bomb.Position == p {
			return true
		}
	}
	return false
}

func (ga *GameArea) ExplodeBomb(bombIdx int) (brokenBoxes int) {
	bomb := ga.bombs[bombIdx]
	for _, boxPos := range ga.GetBoxesInRangeOf(bomb.Position, bomb.ownerID) {
		brokenBoxes++
		ga.grid.SetCellAt(boxPos, empty)
	}
	ga.bombs = append(ga.bombs[:bombIdx], ga.bombs[bombIdx+1:]...)
	return
}

func (ga GameArea) NbBoxesInRangeOf(p Position, playerIdx int) int {
	return len(ga.GetBoxesInRangeOf(p, playerIdx))
}

func (ga GameArea) GetBoxesInRangeOf(p Position, playerIdx int) (boxes []Position) {
	boxes = make([]Position, 0)
	for _, v := range cardinalVectors {
		pos := p
		for i := 0; i < ga.players[playerIdx].bombRange-1; i++ {
			pos = Add(pos, v)
			if pos.IsOnGrid() {
				if isBox(ga.grid.CellAt(pos)) {
					boxes = append(boxes, pos)
					break
				}
			} else {
				break
			}
		}
	}
	return
}

func (ga GameArea) NbTurnsUntilNextAllowedDrop(playerIdx int) int {
	if ga.players[playerIdx].remainingBombs != 0 {
		return 0
	} else {
		minTimer := bombTimer
		for _, bomb := range ga.bombs {
			if bomb.ownerID == playerIdx && bomb.nbRoundsToExplosion < minTimer {
				minTimer = bomb.nbRoundsToExplosion
			}
		}
		return minTimer
	}
}

/* For instance, if you call this function with maxDistance=8, it will return a table of 9 elements:
 * element 0: p
 * element 1: cells reachable in 1 turn
 * ...
 * element 8: cells reachable in 8 turn
 */
func (ga GameArea) ReachableCellsFrom(p Position, maxDistance int) (cells [][]Position) {
	cells = make([][]Position, maxDistance+1)
	var treated Grid

	cells[0] = make([]Position, 1)
	cells[0][0] = p
	treated.SetCellAt(p, '0')

	for i := 1; i <= maxDistance; i++ {
		if len(cells[i-1]) == 0 {
			break //optim
		}
		cells[i] = make([]Position, 0)
		for _, prev := range cells[i-1] {
			for _, v := range cardinalVectors {
				pos := Add(prev, v)
				if pos.IsOnGrid() && treated.CellAt(pos) == 0 { /* was not treated before */
					treated.SetCellAt(pos, '0'+byte(i))
					if isEmpty(ga.grid.CellAt(pos)) {
						cells[i] = append(cells[i], pos)
					} else {
						//treated.SetCellAt(pos, ' ')
					}
				}
			}
		}
	}
	//fmt.Fprint(os.Stderr, treated)
	return cells
}

func (ga GameArea) String() string {
	var g Grid = ga.grid

	for _, bomb := range ga.bombs {
		g.SetCellAt(bomb.Position, 'x')
	}

	for i := range ga.players {
		g.SetCellAt(ga.players[i].Position, 'A'+byte(i))
	}

	var buffer bytes.Buffer
	for i := 0; i < nbRows; i++ {
		buffer.WriteString(fmt.Sprintf("%v\n", g[i]))
	}
	return buffer.String()

}

var me int //index of me

func main() {
	var width, height int
	fmt.Scan(&width, &height, &me)

	for {

		var gameArea GameArea

		gameArea.acquire()

		fmt.Fprint(os.Stderr, gameArea)

		for len(gameArea.bombs) > 0 {
			gameArea.ExplodeBomb(0)
		}

		reachable := gameArea.ReachableCellsFrom(gameArea.players[me].Position, nbRows+nbCols)
		minNbTurn := gameArea.NbTurnsUntilNextAllowedDrop(me)

		var bestDropPos Position
		var bestScore int

		for i := range reachable {
			for _, pos := range reachable[i] {
				cost := Max(minNbTurn, i) + 1
				score := gameArea.NbBoxesInRangeOf(pos, me) * 1000 / cost
				if score > bestScore {
					bestScore = score
					bestDropPos = pos
				}
			}
		}

		// fmt.Fprintln(os.Stderr, "Debug messages...")
		fmt.Printf("BOMB %v %v\n", bestDropPos.x, bestDropPos.y) // Write action to stdout
	}
}
