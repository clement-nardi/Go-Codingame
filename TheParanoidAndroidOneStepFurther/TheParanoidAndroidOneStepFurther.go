package main

import "fmt"
import "os"

/**
 * Auto-generated code below aims at helping you parse
 * the standard input according to the problem statement.
 **/

const elevator = 1
const exit = 2

type State struct {
	floor, pos      int
	dirIsRight      bool
	pathLength      int
	nbElevatorsLeft int
	previous        *State
}

func (s State) String() string {
	dir := "LEFT "
	if s.dirIsRight {
		dir = "RIGHT"
	}
	return fmt.Sprintf("[%v,%v] %v %v %v", s.floor, s.pos, dir, s.nbElevatorsLeft, s.pathLength)
}

func main() {
	// nbFloors: number of floors
	// width: width of the area
	// nbRounds: maximum number of rounds
	// exitFloor: floor on which the exit is found
	// exitPos: position of the exit on its floor
	// nbTotalClones: number of generated clones
	// nbAdditionalElevators: number of additional elevators that you can build
	// nbElevators: number of elevators
	var nbFloors, width, nbRounds, exitFloor, exitPos, nbTotalClones, nbAdditionalElevators, nbElevators int
	fmt.Scan(&nbFloors, &width, &nbRounds, &exitFloor, &exitPos, &nbTotalClones, &nbAdditionalElevators, &nbElevators)

	var cells [15][100]int
	cells[exitFloor][exitPos] = exit

	for i := 0; i < nbElevators; i++ {
		// elevatorFloor: floor on which this elevator is found
		// elevatorPos: position of the elevator on its floor
		var elevatorFloor, elevatorPos int
		fmt.Scan(&elevatorFloor, &elevatorPos)
		cells[elevatorFloor][elevatorPos] = elevator
	}

	var cloneFloor, clonePos int
	var direction string
	fmt.Scan(&cloneFloor, &clonePos, &direction)

	initialState := State{cloneFloor, clonePos, direction[0] == 'R', 0, nbAdditionalElevators, nil}

	states := make([]State, 0, 1024)
	exitStates := make([]State, 0, 1024)
	states = append(states, initialState)

	for i := 0; i < len(states); i++ {
		state := states[i]
		//fmt.Fprintln(os.Stderr,state)
		//expand current state
		if state.floor < nbFloors {
			if cells[state.floor][state.pos] == elevator {
				state.floor++
				state.pathLength++
				state.previous = &states[i]
				states = append(states, state)
			} else if cells[state.floor][state.pos] == exit {
				exitStates = append(exitStates, state)
			} else {
				if state.nbElevatorsLeft > 0 {
					/* try build elevator right away */
					next := state
					next.floor++
					next.nbElevatorsLeft--
					next.pathLength += 1 + 3
					next.previous = &states[i]
					states = append(states, next)
				}
				costRight := 0
				costLeft := 0

				if state.dirIsRight {
					costLeft = 3
				} else {
					costRight = 3
				}
				state.dirIsRight = true

				for x := state.pos + 1; x < width; x++ {
					switch cells[state.floor][x] {
					case elevator:
						next := state
						next.floor++
						next.pos = x
						next.pathLength += x - state.pos + 1 + costRight
						next.previous = &states[i]
						states = append(states, next)
						x = width //impossible to explore further on the right on this floor
					case exit:
						next := state
						next.pos = x
						next.pathLength += x - state.pos + costRight
						next.previous = &states[i]
						exitStates = append(exitStates, next)
						x = width //impossible to explore further on the right on this floor
					case 0:
						if state.floor+1 < nbFloors {
							if state.nbElevatorsLeft > 0 &&
								cells[state.floor+1][x-1] == elevator {
								next := state
								next.floor++
								next.pos = x
								next.nbElevatorsLeft--
								next.pathLength += x - state.pos + 1 + 3 + costRight
								next.previous = &states[i]
								states = append(states, next)
							}
							if state.nbElevatorsLeft > 0 &&
								cells[state.floor+1][x] == exit {
								next := state
								next.floor++
								next.pos = x
								next.nbElevatorsLeft--
								next.pathLength += x - state.pos + 1 + 3 + costRight
								next.previous = &states[i]
								exitStates = append(exitStates, next)
							}
						}
					}
				}

				state.dirIsRight = false

				for x := state.pos - 1; x >= 0; x-- {
					switch cells[state.floor][x] {
					case elevator:
						next := state
						next.floor++
						next.pos = x
						next.pathLength += state.pos - x + 1 + costLeft
						next.previous = &states[i]
						states = append(states, next)
						x = -1 //impossible to explore further on the right on this floor
					case exit:
						next := state
						next.pos = x
						next.pathLength += state.pos - x + costLeft
						next.previous = &states[i]
						exitStates = append(exitStates, next)
						x = -1
						next.pos = x //impossible to explore further on the right on this floor
					case 0:
						if state.floor+1 < nbFloors {
							if state.nbElevatorsLeft > 0 &&
								cells[state.floor+1][x+1] == elevator {
								next := state
								next.floor++
								next.pos = x
								next.nbElevatorsLeft--
								next.pathLength += state.pos - x + 1 + 3 + costLeft
								next.previous = &states[i]
								states = append(states, next)
							}
							if state.nbElevatorsLeft > 0 &&
								cells[state.floor+1][x] == exit {
								next := state
								next.floor++
								next.pos = x
								next.nbElevatorsLeft--
								next.pathLength += state.pos - x + 1 + 3 + costLeft
								next.previous = &states[i]
								exitStates = append(exitStates, next)
							}
						}
					}
				}

			}
		}
	}

	fmt.Fprintf(os.Stderr, "nbSol=%v nbStates=%v\n", len(exitStates), len(states))

	bestState := exitStates[0]

	for _, state := range exitStates {
		//fmt.Fprintln(os.Stderr,state)
		if state.pathLength < bestState.pathLength {
			bestState = state
		}
	}

	bestPath := make([]State, 0)

	for bestState.previous != nil {
		bestPath = append(bestPath, bestState)
		bestState = *(bestState.previous)
	}

	previousState := initialState
	for i := len(bestPath) - 1; i >= 0; i-- {
		currentState := bestPath[i]
		fmt.Fprintln(os.Stderr, currentState)
		//change direction?
		if currentState.dirIsRight != previousState.dirIsRight {
			fmt.Println("BLOCK")
			fmt.Println("WAIT")
			fmt.Println("WAIT")
		}
		dist := currentState.pos - previousState.pos
		if dist < 0 {
			dist = -dist
		}
		for ; dist > 0; dist-- {
			fmt.Println("WAIT")
		}
		//build an elevator?
		if cells[currentState.floor-1][currentState.pos] != elevator {
			fmt.Println("ELEVATOR")
			fmt.Println("WAIT")
			fmt.Println("WAIT")
		}
		//Go up
		fmt.Println("WAIT")

		previousState = bestPath[i]
	}

	for {
		// cloneFloor: floor of the leading clone
		// clonePos: position of the leading clone on its floor
		// direction: direction of the leading clone: LEFT or RIGHT

		// fmt.Fprintln(os.Stderr, "Debug messages...")

		fmt.Println("WAIT") // action: WAIT or BLOCK

		fmt.Scan(&cloneFloor, &clonePos, &direction)
	}
}
