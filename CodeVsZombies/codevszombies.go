package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
)

/**
 * Save humans, destroy zombies!
 **/

const width = 16000
const height = 9000
const sampling = 10
const killRadius = 2000

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

func deg2rad(deg float64) float64 {
	return deg * (math.Pi / 180)
}

type Coord struct {
	x, y int
}

func (c Coord) String() string {
	return fmt.Sprintf("(%v,%v)", c.x, c.y)
}

func vector(a, b Coord) Coord {
	return Coord{b.x - a.x, b.y - a.y}
}

func (c Coord) coordAt(dist, angle float64) Coord {
	res := c
	res.x += int(math.Cos(deg2rad(angle)) * dist)
	res.y += int(math.Sin(deg2rad(angle)) * dist)
	return res
}
func (in Coord) secure() (c Coord){
	c = in
	if c.x < 0 {
		c.x = 0
	} else if c.x >= width {
		c.x = width - 1
	}
	if c.y < 0 {
		c.y = 0
	} else if c.y >= height {
		c.y = height - 1
	}
	return
}

func distance(a, b Coord) float64 {
	dx := b.x - a.x
	dy := b.y - a.y
	return math.Sqrt(float64(dx*dx + dy*dy))
}

func (c Coord) stepTo(destination Coord, stepSize int) Coord {
	dist := distance(c, destination)
	if dist <= float64(stepSize) {
		return destination
	} else {
		result := c
		result.x += int(math.Floor(float64((destination.x-c.x)*stepSize) / dist))
		result.y += int(math.Floor(float64((destination.y-c.y)*stepSize) / dist))
		return result
	}
}

func (c Coord) shrink() Coord {
	return Coord{c.x / sampling, c.y / sampling}
}
func (c Coord) unshrink() Coord {
	return Coord{c.x * sampling, c.y * sampling}
}

type Biped struct {
	id  int
	pos Coord
}

func (b *Biped) stepTo(destination Coord, stepSize int) {
	b.pos = b.pos.stepTo(destination, stepSize)
}
func (b Biped) String() string {
	return fmt.Sprintf("%v:%v", b.id, b.pos)
}

type Bipeds []Biped

type State struct {
	aliveHumans, deadHumans, zombies Bipeds
	turn                             int
	score                            int
}

func (s State) String() string {
	return fmt.Sprintf("H=%v Z=%v s=%v", len(s.aliveHumans), len(s.zombies), s.score)
}

func (s State) hasWon() bool {
	return len(s.zombies) == 0 && len(s.aliveHumans) >= 2
}
func (s State) hasLost() bool {
	return len(s.aliveHumans) < 2
}
func (s State) isFinished() bool {
	return s.hasLost() || s.hasWon()
}

func (s State) copyState() State {
	var next State

	next.aliveHumans = make(Bipeds, len(s.aliveHumans))
	next.deadHumans = make(Bipeds, len(s.deadHumans))
	next.zombies = make(Bipeds, len(s.zombies))

	copy(next.aliveHumans, s.aliveHumans)
	copy(next.deadHumans, s.deadHumans)
	copy(next.zombies, s.zombies)
	next.turn = s.turn
	next.score = s.score

	return next
}

func (s State) nextState(myTarget Coord) (next State) {

	next = s.copyState()

	//Zombies move
	for i := range next.zombies {
		z := &next.zombies[i]
		z.stepTo(next.aliveHumans.closestFrom(z.pos).pos, 400)
	}

	//I move
	me := &next.aliveHumans[0]
	me.stepTo(myTarget, 1000)

	//I kill Zombies
	nbZ := len(next.zombies)
	nbOtherHumans := len(next.aliveHumans) - 1
	humanFactor := nbOtherHumans * nbOtherHumans * 10
	fibo1, fibo2 := 1, 1
	for i := 0; i < nbZ; i++ {
		z := &next.zombies[i]
		if distance(z.pos, me.pos) < 2000.0 {
			next.zombies[i] = next.zombies[nbZ-1]
			nbZ--
			i--
			//TODO compute score
			next.score += humanFactor * fibo2
			fibo1, fibo2 = fibo2, fibo1+fibo2
		}
	}
	next.zombies = next.zombies[:nbZ]

	//Zombies kill Humans
	for zi := range next.zombies {
		z := &next.zombies[zi]
		nbH := len(next.aliveHumans)
		for i := 0; i < nbH; i++ {
			h := &next.aliveHumans[i]
			if h.pos == z.pos {
				next.deadHumans = append(next.deadHumans, next.aliveHumans[i])
				next.aliveHumans[i] = next.aliveHumans[nbH-1]
				nbH--
				i--
			}
		}
		next.aliveHumans = next.aliveHumans[:nbH]
	}

	next.turn++

	return
}

func (bipeds Bipeds) closestFrom(position Coord) *Biped {
	minDist := 30000.0
	var closest *Biped = nil
	for i := range bipeds {
		b := &bipeds[i]
		dist := distance(position, b.pos)
		if dist < minDist {
			minDist = dist
			closest = b
		}
	}
	return closest
}

type Field [width / sampling][height / sampling]int8

var circle2000 [killRadius * 2][killRadius * 2]bool

func applyCircle2000Around(field *Field, c_ Coord) (maxZ int8, maxPos Coord) {
	c := c_.shrink()
	maxZ = int8(-1)
	for i := Max(0, c.x-killRadius/sampling); i < Min(width/sampling, c.x+killRadius/sampling); i++ {
		for j := Max(0, c.y-killRadius/sampling); j < Min(height/sampling, c.y+killRadius/sampling); j++ {
			if circle2000[(i-c.x)*sampling+killRadius][(j-c.y)*sampling+killRadius] {
				field[i][j]++
				if field[i][j] > maxZ {
					maxZ = field[i][j]
					maxPos = Coord{i, j}
				}
			}
		}
	}
	return
}
func resetCircle2000Around(field *Field, c_ Coord) {
	c := c_.shrink()
	for i := Max(0, c.x-killRadius/sampling); i < Min(width/sampling, c.x+killRadius/sampling); i++ {
		for j := Max(0, c.y-killRadius/sampling); j < Min(height/sampling, c.y+killRadius/sampling); j++ {
			if circle2000[(i-c.x)*sampling+killRadius][(j-c.y)*sampling+killRadius] {
				field[i][j] = 0
			}
		}
	}
}

func (field *Field) String() string {
	const targetWidth = 60
	var buffer bytes.Buffer

	for j := 0; j < height*targetWidth/width; j++ {
		for i := 0; i < targetWidth; i++ {
			buffer.WriteString(fmt.Sprint(field[i*width/targetWidth/sampling][j*width/targetWidth/sampling]))
		}
		buffer.WriteString("\n")
	}
	return buffer.String()
}

var field *Field = new(Field)

func (s State) evaluateBest() (dest Coord, maxScore int) {
	maxScore = 0
	for _, h := range s.aliveHumans {
		//fmt.Fprintf(os.Stderr,"Try go to H%v\n", h.id)

		state := s.copyState()
		for !state.isFinished() {
			state = state.nextState(h.pos)
			//fmt.Fprintln(os.Stderr,state)
		}
		if state.hasWon() {
			if state.score > maxScore {
				maxScore = state.score
				dest = h.pos
			}
		}
	}
	return
}

func main() {

	for i := 0; i < killRadius*2; i++ {
		for j := 0; j < killRadius*2; j++ {
			if distance(Coord{i, j}, Coord{killRadius, killRadius}) < float64(killRadius-sampling) {
				circle2000[i][j] = true
			}
		}
	}

	var currentState State
	currentState.turn = 0
	currentState.score = 0

	for {
		currentState.aliveHumans = make([]Biped, 0)
		currentState.deadHumans = make([]Biped, 0)
		currentState.zombies = make([]Biped, 0)

		var x, y int
		fmt.Scan(&x, &y)

		me := Biped{-1, Coord{x, y}}

		currentState.aliveHumans = append(currentState.aliveHumans, me)

		var humanCount int
		fmt.Scan(&humanCount)

		for i := 0; i < humanCount; i++ {
			var humanId, humanX, humanY int
			fmt.Scan(&humanId, &humanX, &humanY)
			currentState.aliveHumans = append(currentState.aliveHumans, Biped{humanId, Coord{humanX, humanY}})
		}
		var zombieCount int
		fmt.Scan(&zombieCount)

		for i := 0; i < zombieCount; i++ {
			var zombieId, zombieX, zombieY, zombieXNext, zombieYNext int
			fmt.Scan(&zombieId, &zombieX, &zombieY, &zombieXNext, &zombieYNext)

			zombie := Biped{zombieId, Coord{zombieX, zombieY}}
			currentState.zombies = append(currentState.zombies, zombie)

			next := Coord{zombieXNext, zombieYNext}
			computedNext := zombie.pos.stepTo(currentState.aliveHumans.closestFrom(zombie.pos).pos, 400)
			if next != computedNext {
				fmt.Fprintf(os.Stderr, "%v: %v != %v\n", zombieId, next, computedNext)
			}
		}


		// fmt.Fprintln(os.Stderr, "Debug messages...")

		/*
			for _, z := range zombies {
				resetCircle2000Around(field, z.pos.stepTo(humans.closestFrom(z.pos).pos, 400))
			}

			max := int8(-1)
			for _, z := range zombies {
				maxz, maxPos := applyCircle2000Around(field, z.pos.stepTo(humans.closestFrom(z.pos).pos, 400))
				if maxz > max {
					max = maxz
					dest = maxPos.unshrink()
				}
			}

			//fmt.Fprint(os.Stderr, field)

			//dest := zombies.closestFrom(me.pos).pos
		*/

		dest, maxScore := currentState.evaluateBest()
		fmt.Fprintln(os.Stderr, currentState)
		
		

		for _, dist := range []float64{1002.0, 800.0, 600.0, 400.0} {
			for angle := 0.0; angle < 360.0; angle += 360.0 / 16.0 {
				nextPos := me.pos.coordAt(dist, angle).secure()
				state := currentState.nextState(nextPos)
				_,score := state.evaluateBest()
				if score > maxScore {
					maxScore = score
					dest = nextPos
				}
			}
		}

		fmt.Fprintf(os.Stderr, "currentScore = %v\n", currentState.score)
		fmt.Fprintf(os.Stderr, "maxScore     = %v\n", maxScore)

		fmt.Printf("%v %v\n", dest.x, dest.y) // Your destination coordinates

		currentState = currentState.nextState(dest) //updates score and turn

	}
}
