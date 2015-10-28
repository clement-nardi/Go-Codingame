package main

import "fmt"
import "os"

type Coord struct {
	r, c int
}

type Node struct {
	coord            Coord
	previousCoordIdx int
	depth            int
}

type Grid [][]int

func (grid Grid) At(coord Coord) *int {
	return &grid[coord.r][coord.c]
}

func (grid_ Grid) PathTo(start Coord, target int) (direction string, distance int) {

	grid := make(Grid, len(grid_))
	for i := range grid {
		grid[i] = make([]int,len(grid_[i]))
		copy(grid[i], grid_[i])
	}
	queue := make([]Node, 0, 128)
	queue = append(queue, Node{start, -1, 0})

	targetIdx := -1

	for i := 0; i < len(queue) && targetIdx == -1; i++ {
		//fmt.Fprintf(os.Stderr, "%v", queue[i])
		currentCell := grid.At(queue[i].coord)
		/*if *currentCell >= 200 {
			fmt.Fprintf(os.Stderr, "%v", (*currentCell-200)%10)
		} else {
			fmt.Fprintf(os.Stderr, "%c", uint8(*currentCell))
		}*/

		if *currentCell < 200 {
			switch *currentCell {
			case target:
				targetIdx = i
			case '#':
			case '?':
				// do nothing
			default:
				*currentCell = queue[i].depth + 200
				nextCells := [4]Node{queue[i], queue[i], queue[i], queue[i]}
				nextCells[0].coord.r--
				nextCells[1].coord.r++
				nextCells[2].coord.c--
				nextCells[3].coord.c++
				for _, cell := range nextCells {
					if cell.coord.r >= 0 &&
						cell.coord.c >= 0 &&
						cell.coord.r < len(grid) &&
						cell.coord.c < len(grid[0]) {
						cell.previousCoordIdx = i
						cell.depth = queue[i].depth + 1
						queue = append(queue, cell)
					}
				}
			}
		}
	}

	if targetIdx == -1 {
		return "", -1
	} else {
		nextNode := queue[targetIdx]
		for nextNode.previousCoordIdx != 0 {
			nextNode = queue[nextNode.previousCoordIdx]
		}
		var dir string
		nextCoord := nextNode.coord
		if nextCoord.c > start.c {
			dir = "RIGHT"
		} else if nextCoord.c < start.c {
			dir = "LEFT"
		} else if nextCoord.r > start.r {
			dir = "DOWN"
		} else {
			dir = "UP"
		}
		return dir, queue[targetIdx].depth
	}

}

func main() {
	// R: number of rows.
	// C: number of columns.
	// A: number of rounds between the time the alarm countdown is activated and the time the alarm goes off.
	var R, C, A int
	fmt.Scan(&R, &C, &A)

	grid := make(Grid, R)
	for i := range grid {
		grid[i] = make([]int, C)
	}

	controlRoomFound := false

	for {
		// KR: row where Kirk is located.
		// KC: column where Kirk is located.
		var kirk, controlRoom Coord
		fmt.Scan(&kirk.r, &kirk.c)

		for i := 0; i < R; i++ {
			// ROW: C of the characters in '#.TC?' (i.e. one line of the ASCII maze).
			var ROW string
			fmt.Scan(&ROW)
			for j, char := range ROW {
				grid[i][j] = int(char)
				if char == 'C' {
					controlRoom = Coord{i, j}
				}
			}
		}

		if controlRoomFound {
			A--
		}

		if *grid.At(kirk) == int('C') {
			controlRoomFound = true
		}

		finalDir := ""

		if !controlRoomFound {
			_, distControlOrigin := grid.PathTo(controlRoom, 'T')
			fmt.Fprintf(os.Stderr, "distControlOrigin=%v\n", distControlOrigin)
			if distControlOrigin == -1 || distControlOrigin > A {
				finalDir, _ = grid.PathTo(kirk, '?')
				fmt.Fprintf(os.Stderr, "-> ? finalDir=%v\n", finalDir)
			} else {
				finalDir, _ = grid.PathTo(kirk, 'C')
				fmt.Fprintf(os.Stderr, "-> C finalDir=%v\n", finalDir)
			}
		} else {
			finalDir, _ = grid.PathTo(kirk, 'T')
			fmt.Fprintf(os.Stderr, "-> T finalDir=%v\n", finalDir)
		}

		if finalDir == "" { //should not happen
			finalDir, _ = grid.PathTo(kirk, '?')
			fmt.Fprintf(os.Stderr, "!! -> ? finalDir=%v\n", finalDir)
		}

		//Print
		for r, line := range grid {
			for c, cell := range line {
				if r == kirk.r && c == kirk.c {
					fmt.Fprintf(os.Stderr, "%c", 'K')
				} else if cell >= 200 {
					fmt.Fprintf(os.Stderr, "%v", (cell-200)%10)
				} else {
					fmt.Fprintf(os.Stderr, "%c", uint8(cell))
				}
			}
			fmt.Fprintln(os.Stderr)
		}

		// fmt.Fprintln(os.Stderr, "Debug messages...")

		fmt.Println(finalDir) // Kirk's next move (UP DOWN LEFT or RIGHT).
	}
}
