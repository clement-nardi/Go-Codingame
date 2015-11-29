package main

import (
	"container/heap"
	"fmt"
	"os"
)

func Min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

type Action int

const (
	speed Action = iota
	slow
	jump
	wait
	up
	down
)

func (a Action) String() string {
	switch a {
	case speed:
		return "SPEED"
	case slow:
		return "SLOW"
	case jump:
		return "JUMP"
	case wait:
		return "WAIT"
	case up:
		return "UP"
	case down:
		return "DOWN"
	default:
		return "????"
	}
}

type Bike struct {
	lane, col int
	isActive  bool
}

func (b Bike) String() string {
	return fmt.Sprintf("(%v,%v) %v", b.lane, b.col, b.isActive)
}

type State struct {
	bikes    []Bike
	speed    int
	previous *State
	action   Action
}

func (s State) String() string {
	out := fmt.Sprintf("Speed=%v Action=%v\n", s.speed, s.action)
	for i, b := range s.bikes {
		out = out + fmt.Sprintf("%v: %v\n", i, b)
	}
	return out
}

type StateKey struct {
	dist  int
	state *State
}

func (s *State) Key() StateKey {
	return StateKey{s.distance(), s}
}

var M int

func acquireState() (s State) {
	// S: the motorbikes' speed
	fmt.Scan(&s.speed)
	for i := 0; i < M; i++ {
		// X: x coordinate of the motorbike
		// Y: y coordinate of the motorbike
		// A: indicates whether the motorbike is activated "1" or detroyed "0"
		var X, Y, A int
		fmt.Scan(&X, &Y, &A)
		s.bikes = append(s.bikes, Bike{Y, X, A == 1})
	}
	return
}

func (s *State) Next(a Action) (next *State) {

	if s.speed == 0 && a != speed {
		return nil
	}
	next = new(State)
	next.previous = s
	next.speed = s.speed
	next.action = a
	next.bikes = make([]Bike, len(s.bikes))
	copy(next.bikes, s.bikes)
	// change speed
	switch a {
	case speed:
		next.speed++
	case slow:
		next.speed--
	}
	//move right
	for i := range next.bikes {
		next.bikes[i].col += next.speed
	}
	//move up/down
	if a == up {
		for _, b := range next.bikes {
			if b.lane == 0 && b.isActive {
				//theresABikeOnTheTopLane
				return nil
			}
		}
		for i := range next.bikes {
			next.bikes[i].lane--
		}
	}
	if a == down {
		for _, b := range next.bikes {
			if b.lane == 3 && b.isActive {
				//theresABikeOnTheBottomLane
				return nil
			}
		}
		for i := range next.bikes {
			next.bikes[i].lane++
		}
	}

	//fmt.Fprintf(os.Stderr,"%v ",a)
	//check for wholes
	for i, b := range next.bikes {
		if b.isActive {
			//  destination cell
			if b.col < bridgeLength && terrain[b.lane][b.col] == '0' {
				next.bikes[i].isActive = false
				//fmt.Fprintln(os.Stderr,"0")
			}
			if a != jump {
				// previous cells
				for x := s.bikes[i].col + 1; x < Min(b.col, bridgeLength); x++ {
					//fmt.Fprintf(os.Stderr,"%v ",x)
					// on destination lane
					if terrain[b.lane][x] == '0' {
						next.bikes[i].isActive = false
						break
						//fmt.Fprintln(os.Stderr,"0")
					}
				}
				// on original lane
				if s.bikes[i].lane != b.lane {
					for x := s.bikes[i].col + 1; x < Min(b.col, bridgeLength); x++ {
						if terrain[s.bikes[i].lane][x] == '0' {
							next.bikes[i].isActive = false
							break
							//fmt.Fprintln(os.Stderr,"0")
						}
					}
				}
			}
		}
	}
	//fmt.Fprintf(os.Stderr," -> %v\n",next.bikes[0].isActive)

	return
}
func (s State) nbActiveBikes() int {
	count := 0
	for _, b := range s.bikes {
		if b.isActive {
			count++
		}
	}
	return count
}

var V int

func (s State) isValid() bool {
	return s.nbActiveBikes() >= V
}

func (s State) distance() int {
	pos := 0
	for _, b := range s.bikes {
		if b.isActive {
			pos = b.col
			break
		}
	}
	return pos
}
func (s State) isArrived() bool {
	return s.isValid() && s.distance() > len(terrain[0])
}

type genHeap []*State

func (h genHeap) Len() int { return len(h) }
func (h genHeap) Less(i, j int) bool {
	if h[i].nbActiveBikes() == h[j].nbActiveBikes() {
		return h[i].distance() > h[j].distance()
	} else {
		return h[i].nbActiveBikes() > h[j].nbActiveBikes()
	}
}
func (h genHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *genHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(*State))
}

func (h *genHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

var terrain [4]string
var bridgeLength int

func main() {
	// M: the amount of motorbikes to control

	fmt.Scan(&M)

	// V: the minimum amount of motorbikes that must survive
	fmt.Scan(&V)

	// L0: L0 to L3 are lanes of the road. A dot character . represents a safe space, a zero 0 represents a hole in the road.
	for i := 0; i < 4; i++ {
		fmt.Scan(&terrain[i])
		fmt.Fprintf(os.Stderr, "%v\n", terrain[i])
	}
	bridgeLength = len(terrain[0])

	state := acquireState()

	//testing
	//	var state State
	//	state.speed = 4
	//	state.bikes = append(state.bikes,Bike{2,10,true})

	fmt.Fprintf(os.Stderr, "%v", state)

	states := make(genHeap, 0, 1000)
	states.Push(&state)
	heap.Init(&states)

	var solution *State = nil

	for len(states) > 0 {
		currentState := heap.Pop(&states).(*State)

		//fmt.Fprintf(os.Stderr, "%v", currentState)
		if currentState.isArrived() {
			if solution == nil ||
				currentState.nbActiveBikes() > solution.nbActiveBikes() {
				solution = currentState
				fmt.Fprintf(os.Stderr, "%v", currentState)
				break
			}
		}
		if currentState.isValid() {
			for a := speed; a <= down; a++ {
				next := currentState.Next(a)
				if next != nil {
					heap.Push(&states, next)
					//fmt.Fprintf(os.Stderr, "NEXT %v", next)
				}
			}
		}
	}
	fmt.Fprintln(os.Stderr, "Done!\n")

	path := make([]*State, 0, 20)
	for solution != nil {
		path = append(path, solution)
		solution = solution.previous
	}

	for i := len(path) - 2; i >= 0; i-- {
		fmt.Fprintf(os.Stderr, "%v", path[i])
		fmt.Println(path[i].action)
	}

	for {
		state = acquireState()
	}
}
