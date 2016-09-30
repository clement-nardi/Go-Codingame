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
	boxWithExtraRange    = '1'
	boxWithExtraBomb     = '2'
	wall                 = 'X'
	empty                = '.'
	nbCardinalDirections = 4
	bombTimer            = 8
	itemNone             = 0
	itemExtraRange       = 1
	itemExtraBomb        = 2
	maxPlayers           = 4
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
	remainingBombs int
	bombRange      int
	isDead         bool
}

type Bomb struct {
	Position
	ownerID int
}

type Cell int16

// prevent player from walking:
// bit : bomb from player 0
// bit : bomb from player 1
// bit : bomb from player 2
// bit : bomb from player 3
// bit : wall
// bit : box without item
// bit : box with item type 1 extra range
// bit : box with item type 2 extra bomb
//
// don't prevent players from walking:
// bit : player 0
// bit : player 1
// bit : player 2
// bit : player 3
// bit : item type 1 extra range
// bit : item type 2 extra bomb
const (
	bombPlayer0Bit    = 01
	bombPlayer1Bit    = 02
	bombPlayer2Bit    = 04
	bombPlayer3Bit    = 010
	wallBit           = 020
	boxWithoutItemBit = 040
	boxWithItem1Bit   = 0100
	boxWithItem2Bit   = 0200
	player0Bit        = 0400
	player1Bit        = 01000
	player2Bit        = 02000
	player3Bit        = 04000
	item1Bit          = 010000
	item2Bit          = 020000

	inaccessibleBits = 0377
	boxBits          = 0340
	bombBits         = 017
	playerBits       = 07400
	itemBits         = 030000
)

func (c *Cell) SetBomb(playerId int) {
	switch playerId {
	case 0:
		*c |= bombPlayer0Bit
	case 1:
		*c |= bombPlayer1Bit
	case 2:
		*c |= bombPlayer2Bit
	case 3:
		*c |= bombPlayer3Bit
	}
}
func (c *Cell) ResetBomb(playerId int) {
	switch playerId {
	case 0:
		*c &^= bombPlayer0Bit
	case 1:
		*c &^= bombPlayer1Bit
	case 2:
		*c &^= bombPlayer2Bit
	case 3:
		*c &^= bombPlayer3Bit
	}
}
func (c *Cell) SetPlayer(playerId int) {
	switch playerId {
	case 0:
		*c |= player0Bit
	case 1:
		*c |= player1Bit
	case 2:
		*c |= player2Bit
	case 3:
		*c |= player3Bit
	}
}
func (c *Cell) ResetPlayer(playerId int) {
	switch playerId {
	case 0:
		*c &^= player0Bit
	case 1:
		*c &^= player1Bit
	case 2:
		*c &^= player2Bit
	case 3:
		*c &^= player3Bit
	}
}
func (c *Cell) SetWall() {
	*c |= wallBit
}
func (c *Cell) ResetWall() {
	*c &^= wallBit
}
func (c *Cell) SetBox(itemType int) {
	switch itemType {
	case itemNone:
		*c |= boxWithoutItemBit
	case itemExtraRange:
		*c |= boxWithItem1Bit
	case itemExtraBomb:
		*c |= boxWithItem2Bit
	}
}
func (c *Cell) ResetBox() {
	*c &^= boxBits
}

func (c *Cell) SetItem(itemType int) {
	switch itemType {
	case itemExtraRange:
		*c |= item1Bit
	case itemExtraBomb:
		*c |= item2Bit
	}
}
func (c *Cell) ResetItem() {
	*c &^= itemBits
}

func (c Cell) isEmpty() bool {
	return c == 0
}
func (c Cell) CanReceiveBombNow() bool {
	return c.isEmpty()
}
func (c Cell) isAccessible() bool {
	return (c & inaccessibleBits) == 0
}
func (c Cell) isWall() bool {
	return (c & wallBit) > 0
}
func (c Cell) isBox() bool {
	return (c & boxBits) > 0
}
func (c Cell) isBomb() bool {
	return (c & bombBits) > 0
}
func (c Cell) isPlayer() bool {
	return (c & playerBits) > 0
}
func (c Cell) isItem() bool {
	return (c & itemBits) > 0
}
func (c Cell) getItemType() int {
	if c&(item1Bit|boxWithItem1Bit) > 0 {
		return itemExtraRange
	} else if c&(item2Bit|boxWithItem2Bit) > 0 {
		return itemExtraBomb
	}
	return itemNone
}
func (c Cell) getPlayerIds() (nbPlayers int, playerIds [maxPlayers]int) {
	if (c & player0Bit) > 0 {
		playerIds[nbPlayers] = 0
		nbPlayers++
	}
	if (c & player1Bit) > 0 {
		playerIds[nbPlayers] = 1
		nbPlayers++
	}
	if (c & player2Bit) > 0 {
		playerIds[nbPlayers] = 2
		nbPlayers++
	}
	if (c & player3Bit) > 0 {
		playerIds[nbPlayers] = 3
		nbPlayers++
	}
	//fmt.Fprintf(os.Stderr, "nbPlayers=%v\n", nbPlayers)
	return
}
func (c Cell) getBombPlayerIds() (nbPlayers int, playerIds [maxPlayers]int) {
	if (c & bombPlayer0Bit) > 0 {
		playerIds[nbPlayers] = 0
		nbPlayers++
	}
	if (c & bombPlayer1Bit) > 0 {
		playerIds[nbPlayers] = 1
		nbPlayers++
	}
	if (c & bombPlayer2Bit) > 0 {
		playerIds[nbPlayers] = 2
		nbPlayers++
	}
	if (c & bombPlayer3Bit) > 0 {
		playerIds[nbPlayers] = 3
		nbPlayers++
	}
	//fmt.Fprintf(os.Stderr, "nbPlayers=%v\n", nbPlayers)
	return
}

type Line [nbCols]Cell
type Grid [nbRows]Line

func (g *Grid) acquire() {
	for i := 0; i < nbRows; i++ {
		var s string
		fmt.Scan(&s)
		for j := 0; j < nbCols; j++ {
			switch s[j] {
			case box:
				g[i][j].SetBox(itemNone)
			case boxWithExtraRange:
				g[i][j].SetBox(itemExtraRange)
			case boxWithExtraBomb:
				g[i][j].SetBox(itemExtraBomb)
			case wall:
				g[i][j].SetWall()
			}
		}
	}
}

func (cell Cell) String() string {
	//fmt.Fprintf(os.Stderr, "raw: %o\n", int16(cell))

	if cell.isEmpty() {
		return " "
	} else if cell.isWall() {
		return "."
	} else if cell.isBox() {
		switch cell.getItemType() {
		case itemNone:
			return "N"
		case itemExtraRange:
			return "R"
		case itemExtraBomb:
			return "M"
		}
	} else if cell.isItem() {
		switch cell.getItemType() {
		case itemExtraRange:
			return "r"
		case itemExtraBomb:
			return "m"
		}
	} else if cell.isBomb() {
		_, ids := cell.getBombPlayerIds()
		return string('a' + byte(ids[0]))
	} else if cell.isPlayer() {
		_, ids := cell.getPlayerIds()
		return string('A' + byte(ids[0]))
	}
	return "?"
}

func (l Line) String() string {
	var buffer bytes.Buffer
	for _, cell := range l {
		buffer.WriteString(fmt.Sprintf("%v", cell))
	}
	return buffer.String()
}
func (g Grid) String() string {
	var buffer bytes.Buffer
	for i := 0; i < nbRows; i++ {
		buffer.WriteString(fmt.Sprintf("%v\n", g[i]))
	}
	return buffer.String()
}

func (g *Grid) CellAt(p Position) *Cell {
	return &(g[p.y][p.x])
}

type GameArea struct {
	grid Grid

	droppedBombs []Bomb
	players      [maxPlayers]Player

	previous *GameArea
}

func (ga *GameArea) acquire() {
	ga.grid.acquire()

	var nbEntities int
	fmt.Scan(&nbEntities)

	ga.droppedBombs = make([]Bomb, 0)

	for i := 0; i < nbEntities; i++ {
		var entityType, owner, x, y, param1, param2 int
		fmt.Scan(&entityType, &owner, &x, &y, &param1, &param2)
		pos := Position{x, y}
		switch entityType {
		case playerEntity:
			ga.players[owner] = Player{param1, param2, false}
			ga.grid.CellAt(pos).SetPlayer(owner)
		case bombEntity:
			ga.grid.CellAt(pos).SetBomb(owner)
			if ga.previous != nil {
				if !ga.previous.grid.CellAt(pos).isBomb() {
					var bomb = Bomb{pos, owner}
					ga.droppedBombs = append(ga.droppedBombs, bomb)
				}
			}
		case itemEntity:
			ga.grid.CellAt(pos).SetItem(param1)
		}
	}
}

func (ga *GameArea) NbBoxesInRangeOf(p Position, playerIdx int) int {
	return len(ga.GetBoxesInRangeOf(p, playerIdx))
}

func (ga *GameArea) GetBoxesInRangeOf(p Position, playerIdx int) (boxes []Position) {
	boxes = make([]Position, 0)
	for _, v := range cardinalVectors {
		pos := p
		for i := 0; i < ga.players[playerIdx].bombRange-1; i++ {
			pos = Add(pos, v)
			if pos.IsOnGrid() {
				cell := ga.grid.CellAt(pos)
				if !cell.isEmpty() {
					if cell.isBox() {
						boxes = append(boxes, pos)
					}
					break
				}
			} else {
				break
			}
		}
	}
	return
}

func (ga GameArea) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%v", ga.grid))
	buffer.WriteString(fmt.Sprintf("%v\n", ga.droppedBombs))
	buffer.WriteString(fmt.Sprintf("%v\n", ga.players))

	return buffer.String()

}

var me int //index of me

func main() {

	var c Cell
	i := 0

	fmt.Fprintln(os.Stderr, i, c)
	i++
	c.SetWall()
	fmt.Fprintln(os.Stderr, i, c)
	i++
	c.ResetWall()
	fmt.Fprintln(os.Stderr, i, c)
	i++
	c.SetBomb(2)
	fmt.Fprintln(os.Stderr, i, c)
	i++
	c.ResetBomb(2)

	fmt.Fprintln(os.Stderr, i, c)
	i++
	c.SetBox(1)

	fmt.Fprintln(os.Stderr, i, c)
	i++
	c.ResetBox()
	fmt.Fprintln(os.Stderr, i, c)
	i++
	c.SetPlayer(1)
	fmt.Fprintln(os.Stderr, i, c)
	i++
	c.ResetPlayer(1)
	fmt.Fprintln(os.Stderr, i, c)
	i++

	var width, height int
	fmt.Scan(&width, &height, &me)
	var gameArea *GameArea = nil
	var previous *GameArea = nil

	for {

		gameArea = new(GameArea)

		gameArea.acquire()
		gameArea.previous = previous

		fmt.Fprint(os.Stderr, gameArea)

		// fmt.Fprintln(os.Stderr, "Debug messages...")
		fmt.Printf("BOMB %v %v\n", 5, 6) // Write action to stdout
		previous = gameArea
	}
}
