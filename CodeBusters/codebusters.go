package main

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"os"
)

//import "os"

const (
	visibility      = 2200
	width           = 16001
	height          = 9001
	bustMinDistance = 900
	bustMaxDistance = 1760
	baseRadius      = 1600
	maxTurns        = 400
	stunDistance    = 1760
	stunRecharge    = 20
	stunDuration    = 10
	busterSpeed     = 800
	ghostSpeed      = 400
	mapGranularity  = 20
)

var optimalDistanceToBorder float64

type Position struct {
	x, y int
}

func (p Position) String() string {
	return fmt.Sprintf("%v %v", p.x, p.y)
}
func Vector(a, b Position) Position {
	return Position{b.x - a.x, b.y - a.y}
}
func VectorTrigo(angleInRadians float64, norm int) Position {
	/* Angle as in trigonometry: from the x-axis. 0 radians = to the right */
	return Position{-int(math.Cos(angleInRadians) * float64(norm)), int(math.Sin(angleInRadians) * float64(norm))}
}
func AngleFromVector(v Position) float64 {
	return math.Atan2(float64(-v.y), float64(v.x))
}
func (a Position) distanceTo(b Position) float64 {
	ab := Vector(a, b)
	return math.Sqrt(float64(ab.x*ab.x + ab.y*ab.y))
}
func Add(a, b Position) Position {
	return Position{a.x + b.x, a.y + b.y}
}
func Divide(v Position, d int) Position {
	return Position{v.x / d, v.y / d}
}
func Multiply(v Position, d int) Position {
	return Position{v.x * d, v.y * d}
}
func barycenter(positions []Position) Position {
	sum := Position{0, 0}
	for _, p := range positions {
		sum = Add(sum, p)
	}
	return Divide(sum, len(positions))
}
func (vector Position) Norm() float64 {
	return vector.distanceTo(Position{0, 0})
}
func Normalize(vector Position, norm float64) Position {
	oldNorm := vector.Norm()
	if oldNorm != 0 {
		return Position{int(float64(vector.x) * norm / oldNorm), int(float64(vector.y) * norm / oldNorm)}
	} else {
		return vector
	}
}
func IsSecure(p Position) bool {
	return p.x >= 0 && p.x < width && p.y >= 0 && p.y < height
}
func Secure(p Position) Position {
	secured := p
	if p.x < 0 {
		secured.x = 0
	}
	if p.x >= width {
		secured.x = width - 1
	}
	if p.y < 0 {
		secured.y = 0
	}
	if p.y >= height {
		secured.y = height - 1
	}
	return secured
}

func OverlappingAreaOfTwoCircles(a, b Position, r int) float64 {
	R := float64(r)
	d := a.distanceTo(b)
	if d >= 2*R {
		return 0
	} else {
		return 2*R*R*math.Acos(d/(2*R)) - d/2*math.Sqrt(4*R*R-d*d)
	}
}

type Entity struct {
	Position
	id         int
	entityType int
	state      int
	value      int
	turn       int16
}

func (e *Entity) Acquire() {
	fmt.Scan(&e.id, &e.x, &e.y, &e.entityType, &e.state, &e.value)
	e.turn = turn
}
func (e Entity) isInBase(id int) bool {
	return e.distanceTo(bases[id]) < baseRadius
}
func (e Entity) IsVisible() bool {
	return e.turn == turn
}
func (e Entity) SeenOnce() bool {
	return e.turn != 0
}

type MovingEntity struct {
	Entity
	history []Entity
}

func (me *MovingEntity) SetNewState(e Entity) {
	me.history = append(me.history, me.Entity)
	me.Entity = e
}

type Ghost struct {
	MovingEntity
	isTargetted bool
}

func (g Ghost) NbSuckingBusters() int {
	return g.value
}

type OrderNature int

const (
	idle OrderNature = iota
	move
	bust
	release
	stun
)

type Order struct {
	nature    OrderNature
	targetID  int
	targetPos Position
}

func (o Order) String() string {
	var orderString string
	switch o.nature {
	case move:
		orderString = fmt.Sprintf("MOVE %v", o.targetPos)
	case bust:
		orderString = fmt.Sprintf("BUST %v", o.targetID)
	case release:
		orderString = fmt.Sprintf("RELEASE")
	case stun:
		orderString = fmt.Sprintf("STUN %v", o.targetID)
	case idle: //should not happen
		orderString = fmt.Sprintf("MOVE 8000 4500")
	}
	return orderString
}

type Buster struct {
	MovingEntity
	orders [maxTurns + 1]Order
}

func (b *Buster) CarriesAGhost() bool {
	return b.state != 0
}
func (b *Buster) canBust(g *Ghost) bool {
	if b.CarriesAGhost() {
		return false
	}
	d := b.distanceTo(g.Position)
	return d >= bustMinDistance && d < bustMaxDistance
}
func (b *Buster) canStun(o *Buster) bool {
	if b.distanceTo(o.Position) < stunDistance {
		for t := turn - 1; t > 0 && t >= turn-stunRecharge; t-- {
			if b.orders[t].nature == stun {
				return false
			}
		}
		return true
	}
	return false
}
func (b *Buster) Bust(ghostID int) {
	b.orders[turn].nature = bust
	b.orders[turn].targetID = ghostID
}
func (b *Buster) MoveTo(p Position) {
	b.orders[turn].nature = move
	b.orders[turn].targetPos = p
	b.SecureTarget()
	if b.orders[turn].targetPos == b.Position {
		b.orders[turn].targetPos = Position{rand.Intn(width - 1), rand.Intn(height - 1)}
	}
}
func (b *Buster) ReleaseGhost() {
	b.orders[turn].nature = release
}
func (b *Buster) Stun(busterID int) {
	b.orders[turn].nature = stun
	b.orders[turn].targetID = busterID
}
func (b *Buster) IsIdle() bool {
	return b.orders[turn].nature == idle
}
func (b *Buster) moveAwayFrom(p Position) {
	direction := Vector(p, b.Position)
	direction = Normalize(direction, busterSpeed+1)
	b.MoveTo(Add(b.Position, direction))
}
func (b *Buster) SecureTarget() {
	t := b.orders[turn].targetPos
	if t.x < 0 || t.x >= width {
		if t.y < b.y {
			t.y = 0
		} else {
			t.y = height - 1
		}
	}
	if t.y < 0 || t.y >= height {
		if t.x < b.x {
			t.x = 0
		} else {
			t.x = width - 1
		}
	}
	b.orders[turn].targetPos = Secure(t)
}

type BusterTeam struct {
	busters      map[int]*Buster
	minID, maxID int
}

func (bt BusterTeam) SendOrders() {
	for i := bt.minID; i <= bt.maxID; i++ {
		if bt.busters[i] != nil {
			fmt.Printf("%v\n", bt.busters[i].orders[turn])
		}
	}
}

func (bt *BusterTeam) Init() {
	bt.busters = make(map[int]*Buster)
	bt.minID = 999
}
func (bt *BusterTeam) Update(e Entity) {
	b := bt.busters[e.id]
	if b == nil {
		b = new(Buster)
		bt.busters[e.id] = b
	}
	b.SetNewState(e)
	if e.id < bt.minID {
		bt.minID = e.id
	}
	if e.id > bt.maxID {
		bt.maxID = e.id
	}
}
func (bt *BusterTeam) BarycenterOfOtherBusters(busterID int) Position {
	positions := make([]Position, 0)
	for id, buster := range bt.busters {
		if id != busterID {
			positions = append(positions, buster.Position)
		}
	}
	return barycenter(positions)
}

type GameMap struct {
	mapReduced [width / mapGranularity][height / mapGranularity]int16
}

func Reduce(p Position) Position {
	return Divide(p, mapGranularity)
}
func Unreduce(p Position) Position {
	return Multiply(p, mapGranularity)
}

func SecureReduced(p Position) Position {
	secured := p
	if p.x < 0 {
		secured.x = 0
	}
	if p.x >= width/mapGranularity {
		secured.x = width/mapGranularity - 1
	}
	if p.y < 0 {
		secured.y = 0
	}
	if p.y >= height/mapGranularity {
		secured.y = height/mapGranularity - 1
	}
	return secured
}

func (gm *GameMap) updateDiscovery(p Position) {
	topLeft := SecureReduced(Reduce(Add(p, Position{-visibility, -visibility})))
	bottomRight := SecureReduced(Reduce(Add(p, Position{visibility, visibility})))

	for x := topLeft.x; x <= bottomRight.x; x++ {
		for y := topLeft.y; y <= bottomRight.y; y++ {
			curPos := Position{x, y}
			if Unreduce(curPos).distanceTo(p) < visibility {
				//fmt.Fprintf(os.Stderr, "%v = %v\n", Position{x, y}, turn)

				gm.mapReduced[x][y] = int16(turn)
			}
		}
	}
}

func (gm GameMap) String() string {
	var buffer bytes.Buffer
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			buffer.WriteString(fmt.Sprintf("%v", gm.mapReduced[x][y]%10))
		}
		buffer.WriteString(fmt.Sprintf("\n"))
	}
	return buffer.String()
}

func (gm *GameMap) lastSeen(p Position) int16 {
	r := SecureReduced(Reduce(p))
	return gm.mapReduced[r.x][r.y]
}

func (gm *GameMap) findNearbyUnknownTerritory(p Position, dir Position) Position {
	angle := AngleFromVector(dir)
	nbAngles := 100
	count := 0
	for a := 0; a < nbAngles/2; a++ {
		for coef := -1; coef <= 1; coef += 2 {
			testAngle := angle + 2*math.Pi*float64(coef*a)/float64(nbAngles)
			testVector := VectorTrigo(testAngle, visibility+2*mapGranularity)
			testPos := Add(p, testVector)
			count++
			if IsSecure(testPos) {
				if gm.lastSeen(testPos) == 0 {
					fmt.Fprintf(os.Stderr, "found angle: %v count=%v\n", testAngle*180/math.Pi, count)
					return testPos
				}
			}
		}
	}

	/* No Nearby unknown territory, try further */
	oldest := turn + 1
	oldestPos := Reduce(bases[him])
	oy := Reduce(p).y
	for j := 0; j < height/mapGranularity; j++ {
		coef := 1
		if j%2 == 1 {
			coef = -1
		}
		y := (oy + coef*j/2 + height/mapGranularity) % (height / mapGranularity)
		for x := 0; x < width/mapGranularity; x++ {
			if gm.mapReduced[x][y] == 0 {
				fmt.Fprintf(os.Stderr, "found virgin territory: %v - %v\n", Position{x, y}, Unreduce(Position{x, y}))
				return Unreduce(Position{x, y})
			}
			if gm.mapReduced[x][y] < oldest {
				oldest = gm.mapReduced[x][y]
				oldestPos = Position{x, y}
			}
		}
	}
	fmt.Fprintf(os.Stderr, "...nothing found, odest = %v - %v\n", oldestPos, Unreduce(oldestPos))
	return Unreduce(oldestPos)
}

var ghosts []Ghost
var visibleGhosts []*Ghost
var teams [2]BusterTeam

var me int
var him int

var bases [2]Position = [2]Position{Position{0, 0}, Position{width - 1, height - 1}}

var turn int16

/**
 * Send your busters out into the fog to trap ghosts and bring them home!
 **/

func main() {
	optimalDistanceToBorder = math.Sqrt(visibility*visibility - (busterSpeed/2)*(busterSpeed/2))

	if false {
		//////// OverlappingAreaOfTwoCircles UNIT-TEST////////
		fmt.Fprintf(os.Stderr, "opening file\n")
		os.Remove("./overlap.csv")
		file, _ := os.Create("./overlap.csv")
		fmt.Fprintf(os.Stderr, "file opened\n")
		for i := 0; i < 10000; i++ {
			fmt.Fprintf(file, "%v,%v\n", i, OverlappingAreaOfTwoCircles(Position{0, i}, Position{0, 5000}, 2200))
		}
		file.Close()
		fmt.Fprintf(os.Stderr, "file closed\n")

		//
		fmt.Fprintf(os.Stderr, "optimalDistanceToBorder = %v\n", optimalDistanceToBorder)

		return
	}

	// bustersPerPlayer: the amount of busters you control
	var bustersPerPlayer int
	fmt.Scan(&bustersPerPlayer)

	// ghostCount: the amount of ghosts on the map
	var ghostCount int
	fmt.Scan(&ghostCount)

	// myTeamId: if this is 0, your base is on the top left of the map, if it is one, on the bottom right
	fmt.Scan(&me)

	if me == 0 {
		him = 1
	}

	teams[0].Init()
	teams[1].Init()
	ghosts = make([]Ghost, ghostCount)

	var gameMap GameMap

	turn = 1

	for {
		// entities: the number of busters and ghosts visible to you
		var nbEntities int
		fmt.Scan(&nbEntities)

		visibleGhosts = make([]*Ghost, 0, ghostCount)

		for i := 0; i < nbEntities; i++ {
			// entityId: buster id or ghost id
			// y: position of this buster / ghost
			// entityType: the team id if it is a buster, -1 if it is a ghost.
			// state: For busters: 0=idle, 1=carrying a ghost.
			// value: For busters: Ghost id being carried. For ghosts: number of busters attempting to trap this ghost.
			var entity Entity
			entity.Acquire()
			if entity.entityType == -1 {
				ghosts[entity.id].SetNewState(entity)
				visibleGhosts = append(visibleGhosts, &ghosts[entity.id])
			} else {
				teams[entity.entityType].Update(entity)
				if entity.entityType == me {
					gameMap.updateDiscovery(entity.Position)
				}
			}
		}

		//		if turn == 1 {
		//			for i := range teams[me].busters {
		//				teams[me].busters[i].target = Position{rand.Intn(width - 1), rand.Intn(height - 1)}
		//				fmt.Fprintf(os.Stderr, "target %v: %v\n", i, teams[me].busters[i].target)
		//			}
		//		}

		/* Stun */
		for _, buster := range teams[me].busters {
			if buster.IsIdle() {
				for _, opponent := range teams[him].busters {
					if buster.canStun(opponent) {
						buster.Stun(opponent.id)
					}
				}
			}
		}

		/* Bust */
		for i, ghost := range visibleGhosts {
			visibleGhosts[i].isTargetted = false
			var bestBuster *Buster = nil
			for _, buster := range teams[me].busters {
				if buster.IsIdle() && buster.canBust(ghost) {
					if bestBuster == nil ||
						buster.distanceTo(bases[me]) < bestBuster.distanceTo(bases[me]) {
						bestBuster = buster
					}
				}
			}
			if bestBuster != nil {
				bestBuster.Bust(ghost.id)
				visibleGhosts[i].isTargetted = true
			}
		}

		/* Go Bust */
		for i, ghost := range visibleGhosts {
			if !visibleGhosts[i].isTargetted {
				var bestBuster *Buster = nil
				for _, buster := range teams[me].busters {
					if buster.IsIdle() && !buster.CarriesAGhost() {
						if bestBuster == nil ||
							buster.distanceTo(ghost.Position) < bestBuster.distanceTo(ghost.Position) {
							bestBuster = buster
						}
					}
				}
				if bestBuster != nil {
					bestBuster.MoveTo(ghost.Position)
					visibleGhosts[i].isTargetted = true
				}
			}
		}

		//fmt.Fprintf(os.Stderr, "gameMap:\n%v", gameMap)

		for _, buster := range teams[me].busters {
			if buster.IsIdle() {
				if buster.CarriesAGhost() {
					if buster.isInBase(me) {
						buster.ReleaseGhost()
					} else {
						buster.MoveTo(bases[me])
					}
				} else {
					//buster.moveAwayFrom(barycenter([]Position{bases[me], teams[me].BarycenterOfOtherBusters(buster.id)}))
					//buster.moveAwayFrom(bases[me])
					fmt.Fprintf(os.Stderr, "buster %v\n", buster.id)
					buster.MoveTo(gameMap.findNearbyUnknownTerritory(buster.Position, Vector(teams[me].BarycenterOfOtherBusters(buster.id), buster.Position)))
				}
			}
			// fmt.Fprintln(os.Stderr, "Debug messages...")
		}

		teams[me].SendOrders()

		turn++
	}
}
