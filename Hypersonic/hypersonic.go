package main

import (
	"bytes"
	"container/heap"
	"fmt"
	"os"
	"runtime"
	"time"
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
	move                 = 0
	dropBomb             = 1
	timeoutLimit         = 60
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
	return fmt.Sprintf("(%v,%v)", p.x, p.y)
}

func (p Position) IsOnGrid() bool {
	return p.x >= 0 && p.x < nbCols && p.y >= 0 && p.y < nbRows
}
func Add(a, b Position) Position {
	return Position{a.x + b.x, a.y + b.y}
}

type Player struct {
	Position
	remainingBombs  int
	bombRange       int
	isDead          bool
	score           int
	potential       int
	scorePerTurn    int
	potPerTurn      int
	lastScoreUpdate int
	lastPotUpdate   int
}

func (p Player) String() string {
	return fmt.Sprintf("[s=%v d=%v nbb=%v rng=%v]", p.score, p.isDead, p.remainingBombs, p.bombRange)
}

type Bomb struct {
	Position
	ownerID int
}

func (b Bomb) String() string {
	return fmt.Sprintf("%v:%v", b.ownerID, b.Position)
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
	inflamableBits   = 037757
)

func (c *Cell) Burn() {
	*c &^= inflamableBits
}

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
	//assumes that player is on the cell
	return !c.isBomb()
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
func (c Cell) getPlayerIds() (playerIds []int) {
	if (c & player0Bit) > 0 {
		playerIds = append(playerIds, 0)
	}
	if (c & player1Bit) > 0 {
		playerIds = append(playerIds, 1)
	}
	if (c & player2Bit) > 0 {
		playerIds = append(playerIds, 2)
	}
	if (c & player3Bit) > 0 {
		playerIds = append(playerIds, 3)
	}
	//fmt.Fprintf(os.Stderr, "nbPlayers=%v\n", nbPlayers)
	return
}
func (c Cell) getBombPlayerIds() (playerIds []int) {
	if (c & bombPlayer0Bit) > 0 {
		playerIds = append(playerIds, 0)
	}
	if (c & bombPlayer1Bit) > 0 {
		playerIds = append(playerIds, 1)
	}
	if (c & bombPlayer2Bit) > 0 {
		playerIds = append(playerIds, 2)
	}
	if (c & bombPlayer3Bit) > 0 {
		playerIds = append(playerIds, 3)
	}
	//fmt.Fprintf(os.Stderr, "nbPlayers=%v\n", len(playerIds))
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
		return "+"
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
		ids := cell.getBombPlayerIds()
		return string('a' + byte(ids[0]))
	} else if cell.isPlayer() {
		ids := cell.getPlayerIds()
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

type Action struct {
	action int
	pos    Position
}

func (a Action) String() string {
	var actionString string
	if a.action == dropBomb {
		actionString = "BOMB"
	} else {
		actionString = "MOVE"
	}
	return fmt.Sprintf("%v %v %v", actionString, a.pos.x, a.pos.y)
}

type GameArea struct {
	grid            Grid
	droppedBombs    []Bomb
	players         [maxPlayers]Player
	actionToGetHere Action
	turn            int
	previous        *GameArea
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
			ga.players[owner] = Player{pos, param1, param2, false, 0, 0, 0, 0, 0, 0}
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

func (ga *GameArea) Get8TurnsAgo() *GameArea {
	loopGA := ga
	for i := 0; i < bombTimer; i++ {
		loopGA = loopGA.previous
		if loopGA == nil {
			//fmt.Fprintf(os.Stderr, "Only %v turns\n", i)
			return nil
		}
	}
	return loopGA
}

func (ga *GameArea) ExplodeTimedOutBombs() (burnedCells map[Position]bool) {
	ga8Ago := ga.Get8TurnsAgo()

	if ga8Ago == nil {
		return
	}

	var bombsToExplode []Bomb = make([]Bomb, 0)

	for _, bomb := range ga8Ago.droppedBombs {
		//make sure it didn't explode prematurily
		exploded := false
		loopGA := ga
		for i := 0; i < bombTimer; i++ {
			if !loopGA.grid.CellAt(bomb.Position).isBomb() {
				exploded = true
				break
			}
			loopGA = loopGA.previous
		}
		if !exploded {
			bombsToExplode = append(bombsToExplode, bomb)
		}
	}

	//fmt.Fprintf(os.Stderr, "Bombs to explode: %v\n", bombsToExplode)

	var cellsToBurn map[Position]bool = make(map[Position]bool)
	burnedCells = make(map[Position]bool)
	savedPlayers := ga.players

	for len(bombsToExplode) > 0 {
		countBombExploded++
		if countBombExploded%100 == 0 {
			elapsed := time.Since(begin)
			//fmt.Fprintf(os.Stderr, "b-elapsed: %v\n", elapsed)
			if elapsed > timeoutLimit*time.Millisecond {
				timeout = true
				fmt.Fprintf(os.Stderr, "b-elapsed: %v\n", elapsed)
			}
		}

		bomb := bombsToExplode[0]
		bombrange := ga8Ago.players[bomb.ownerID].bombRange

		ga.grid.CellAt(bomb.Position).ResetBomb(bomb.ownerID)
		ga.players[bomb.ownerID].remainingBombs++
		bombsToExplode = bombsToExplode[1:]
		cellsToBurn[bomb.Position] = true
		burnedCells[bomb.Position] = true

		for _, direction := range cardinalVectors {
			pos := bomb.Position
			for i := 1; i < bombrange; i++ {
				pos = Add(pos, direction)
				if !pos.IsOnGrid() {
					break
				}
				burnedCells[pos] = true
				if !ga.grid.CellAt(pos).isEmpty() {
					cellsToBurn[pos] = true
					if ga.grid.CellAt(pos).isBomb() {
						for _, id := range ga.grid.CellAt(pos).getBombPlayerIds() {
							bombsToExplode = append(bombsToExplode, Bomb{pos, id})
						}
					}
					if ga.grid.CellAt(pos).isBox() {
						ga.players[bomb.ownerID].score++
					}
					break
				}
			}
		}
	}

	for pidx := range ga.players {
		if ga.players[pidx].score != savedPlayers[pidx].score {
			ga.players[pidx].scorePerTurn += (ga.players[pidx].score - savedPlayers[pidx].score) * 1000 / (ga.turn - savedPlayers[pidx].lastScoreUpdate + 1)
			ga.players[pidx].lastScoreUpdate = ga.turn
		}
	}

	for pos := range cellsToBurn {
		ga.BurnCellAt(pos)
	}

	//if len(burnedCells) > 0 {
	//	fmt.Fprintf(os.Stderr, "BurnedCells=%v\n", burnedCells)
	//}

	return
}

func (ga *GameArea) BurnCellAt(pos Position) {
	for id := range ga.grid.CellAt(pos).getPlayerIds() {
		ga.players[id].isDead = true
	}

	itemDropped := itemNone
	if ga.grid.CellAt(pos).isBox() {
		itemDropped = ga.grid.CellAt(pos).getItemType()
	}

	ga.grid.CellAt(pos).Burn()

	ga.grid.CellAt(pos).SetItem(itemDropped)
}

func (ga *GameArea) GetNextStates(playerId int) (nextStates []*GameArea) {

	nbLoops := 1
	oPos := ga.players[playerId].Position

	if ga.players[playerId].remainingBombs > 0 && ga.grid.CellAt(oPos).CanReceiveBombNow() {
		nbLoops = 2
	}

	nextStateBase := *ga
	nextStateBase.turn++
	nextStateBase.previous = ga
	nextStateBase.ExplodeTimedOutBombs()

	nextnext := nextStateBase
	nextnext.previous = &nextStateBase
	forbiddenCells := nextnext.ExplodeTimedOutBombs()

	for i := 0; i < nbLoops; i++ {
		for dir := 0; dir < nbCardinalDirections+1; dir++ {
			pos := oPos
			if dir < nbCardinalDirections {
				pos = Add(pos, cardinalVectors[dir])
			} else {
				// stay there
			}
			if pos.IsOnGrid() && nextStateBase.grid.CellAt(pos).isAccessible() && !forbiddenCells[pos] {
				nextState := new(GameArea)
				*nextState = nextStateBase

				if i > 0 {
					nextState.DropBomb(playerId)
					nextState.actionToGetHere.action = dropBomb
					nextState.players[playerId].potential += nextState.NbBoxesInRangeOf(pos, playerId)
					nextState.players[playerId].potPerTurn += nextState.NbBoxesInRangeOf(pos, playerId) * 1000 / (nextState.turn - nextState.players[playerId].lastPotUpdate + 1)
				}
				nextState.MovePlayer(playerId, pos)
				nextState.actionToGetHere.pos = pos

				nextStates = append(nextStates, nextState)

				countGeneratedStates++
				if countGeneratedStates%100 == 0 {
					elapsed := time.Since(begin)
					//fmt.Fprintf(os.Stderr, "g-elapsed: %v\n", elapsed)
					if elapsed > timeoutLimit*time.Millisecond {
						timeout = true
						fmt.Fprintf(os.Stderr, "g-elapsed: %v\n", elapsed)
					}
				}
			}
		}
	}

	return
}

func (ga *GameArea) DropBomb(playerId int) {
	pos := ga.players[playerId].Position
	ga.droppedBombs = append(ga.droppedBombs, Bomb{pos, playerId})
	ga.grid.CellAt(pos).SetBomb(playerId)
}

func (ga *GameArea) MovePlayer(playerId int, target Position) {
	ga.grid.CellAt(ga.players[playerId].Position).ResetPlayer(playerId)
	ga.players[playerId].Position = target
	ga.grid.CellAt(ga.players[playerId].Position).SetPlayer(playerId)
	if ga.grid.CellAt(target).isItem() {
		switch ga.grid.CellAt(target).getItemType() {
		case itemExtraBomb:
			ga.players[playerId].remainingBombs++
		case itemExtraRange:
			ga.players[playerId].bombRange++
		}
	}
}

func (ga GameArea) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%v", ga.grid))
	buffer.WriteString(fmt.Sprintf("%v\n", ga.droppedBombs))
	buffer.WriteString(fmt.Sprintf("%v\n", ga.players))

	return buffer.String()
}

func Diff(g1, g2 *GameArea) {
	for y, line := range g1.grid {
		for x, cell := range line {
			pos := Position{x, y}
			cell2 := *g2.grid.CellAt(pos)
			if cell != cell2 && !cell.isPlayer() && !cell2.isPlayer() {
				fmt.Fprintf(os.Stderr, "diff at %v: %v != %v\n", pos, cell, cell2)
			}
		}
	}
}

func (ga *GameArea) NbTurnsSinceReference() int {
	return ga.turn - turn
}

func (ga *GameArea) HasToBeTreatedBefore(ga2 *GameArea) bool {
	s1, s2 := ga.players[me].scorePerTurn, ga2.players[me].scorePerTurn
	if s1 != s2 {
		return s1 > s2
	}
	p1, p2 := ga.players[me].potPerTurn, ga2.players[me].potPerTurn
	if p1 != p2 {
		return p1 > p2
	}
	return ga.turn < ga2.turn
}

var nbCritDead, nbCritSpt, nbCritPpt, nbCritTurn int

func (ga *GameArea) IsBetterThan(ga2 *GameArea) bool {
	if ga.players[me].isDead != ga2.players[me].isDead {
		nbCritDead++
		return !ga.players[me].isDead
	}
	s1, s2 := ga.players[me].scorePerTurn, ga2.players[me].scorePerTurn
	if s1 != s2 {
		nbCritSpt++
		return s1 > s2
	}
	p1, p2 := ga.players[me].potPerTurn, ga2.players[me].potPerTurn
	if p1 != p2 {
		nbCritPpt++
		return p1 > p2
	}
	nbCritTurn++
	return ga.turn > ga2.turn
}

type genHeap []*GameArea

func (h genHeap) Len() int            { return len(h) }
func (h genHeap) Less(i, j int) bool  { return h[i].HasToBeTreatedBefore(h[j]) }
func (h genHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *genHeap) Push(x interface{}) { *h = append(*h, x.(*GameArea)) }
func (h *genHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

var me int //index of me
var timeout bool
var begin time.Time
var turn int
var currentGameArea *GameArea
var countBombExploded, countGeneratedStates int

func main() {

	var c Cell
	i := 0

	if false {
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
	}

	var width, height int
	fmt.Scan(&width, &height, &me)
	var previous *GameArea = nil

	for {

		runtime.GC()
		nbCritDead, nbCritSpt, nbCritPpt, nbCritTurn = 0, 0, 0, 0

		currentGameArea = new(GameArea)
		currentGameArea.previous = previous
		currentGameArea.turn = turn

		currentGameArea.acquire()

		timeout = false
		begin = time.Now()

		if previous != nil { // test explosion
			var simulatedCurrentGA GameArea = *previous
			simulatedCurrentGA.previous = previous
			cells := simulatedCurrentGA.ExplodeTimedOutBombs()
			Diff(currentGameArea, &simulatedCurrentGA)
			fmt.Fprintf(os.Stderr, "Burned cells: %v\n", cells)
			for i := 0; i < maxPlayers; i++ {
				currentGameArea.players[i].score = simulatedCurrentGA.players[i].score
			}
		}

		queue := make(genHeap, 0, 1000)
		queue.Push(currentGameArea)
		heap.Init(&queue)

		bestGameArea := currentGameArea
		turnStats := make(map[int]int)

		count := 0

		for len(queue) > 0 {
			var currentGA *GameArea = heap.Pop(&queue).(*GameArea)
			for _, state := range currentGA.GetNextStates(me) {
				if !state.players[me].isDead {
					heap.Push(&queue, state)
					if state.IsBetterThan(bestGameArea) {
						bestGameArea = state
					}
					turnStats[state.turn]++
				}
			}

			count++
			if count%100 == 0 {
				elapsed := time.Since(begin)
				//fmt.Fprintf(os.Stderr, "elapsed: %v\n", elapsed)
				if elapsed > timeoutLimit*time.Millisecond {
					timeout = true
					fmt.Fprintf(os.Stderr, "elapsed: %v\n", elapsed)
				}
			}

			if timeout {
				break
			}
		}

		pathLen := 0

		nextState := bestGameArea

		if nextState != currentGameArea {
			for nextState.previous != currentGameArea {
				nextState = nextState.previous
				pathLen++
			}
		}

		fmt.Fprint(os.Stderr, currentGameArea)

		fmt.Fprintf(os.Stderr, "PathLen=%v treated=%v remaining=%v\n", pathLen, count, len(queue))
		fmt.Fprintf(os.Stderr, "crit d=%v spt=%v ppt=%v t=%v\n", nbCritDead, nbCritSpt, nbCritPpt, nbCritTurn)
		fmt.Fprint(os.Stderr, bestGameArea)

		for i := turn + 1; turnStats[i] > 0; i++ {
			fmt.Fprintf(os.Stderr, "turn %v: %v\n", i, turnStats[i])
		}

		fmt.Printf("%v\n", nextState.actionToGetHere) // Write action to stdout

		previous = currentGameArea
		turn++
	}
}
