package main

import "fmt"
import "os"

//import "bufio"
import "bytes"

//import "strings"
//import "strconv"

type NodeType int

const (
	empty NodeType = iota
	surveillance
	passive
)

type Node struct {
	nodeType                 NodeType
	hasBomb                  bool
	surveillanceNodesInRange int
}

type Coord struct {
	i, j int
}

func (n Node) String() string {
	var typeString string
	switch n.nodeType {
	case surveillance:
		typeString = "@"
	case passive:
		typeString = "#"
	default:
		typeString = " "
	}
	return fmt.Sprintf("%v%v", typeString, n.surveillanceNodesInRange)
}

type Grid struct {
	nodes [][]Node
}

func (g *Grid) NodeAt(c Coord) *Node {
	return &g.nodes[c.i][c.j]
}

func (g *Grid) AcquireCells() {
	var width, height int
	fmt.Scan(&width, &height)

	g.nodes = make([][]Node, height)
	fmt.Fprintf(os.Stderr, "width=%v height=%v %v\n", width, height, len(g.nodes))

	for i := 0; i < height; i++ {
		var mapRow string
		fmt.Scan(&mapRow)
		g.nodes[i] = make([]Node, width)
		for j, c := range mapRow {
			switch c {
			case '#':
				g.nodes[i][j].nodeType = passive
			case '@':
				g.nodes[i][j].nodeType = surveillance
			default:
				g.nodes[i][j].nodeType = empty
			}
		}
	}
	g.ComputeRanges()
}

func (g *Grid) IsOnGrid(c Coord) bool {
	return c.i >= 0 && c.j >= 0 && c.i < len(g.nodes) && c.j < len(g.nodes[0])
}

func (g *Grid) MarkNodesAround(c Coord, mark int) {
	v := [4][2]int{{1, 0}, {0, 1}, {-1, 0}, {0, -1}}
	for k := 0; k < 4; k++ {
		for l := 1; l <= 3; l++ {
			c2 := Coord{c.i + l*v[k][0], c.j + l*v[k][1]}
			if !g.IsOnGrid(c2) || g.NodeAt(c2).nodeType == passive {
				break
			} else {
				g.NodeAt(c2).surveillanceNodesInRange += mark
			}
		}
	}
}

func (g *Grid) ComputeRanges() {

	for _, line := range g.nodes {
		for _, cell := range line {
			cell.surveillanceNodesInRange = 0
		}
	}

	for i, line := range g.nodes {
		for j, cell := range line {
			if cell.nodeType == surveillance {
				g.MarkNodesAround(Coord{i, j}, 1)
			}
		}
	}
}

func (g *Grid) DropBombAt(c Coord) {
	g.NodeAt(c).hasBomb = true
	v := [4][2]int{{1, 0}, {0, 1}, {-1, 0}, {0, -1}}
	for k := 0; k < 4; k++ {
		for l := 1; l <= 3; l++ {
			c2 := Coord{c.i + l*v[k][0], c.j + l*v[k][1]}
			if !g.IsOnGrid(c2) || g.NodeAt(c2).nodeType == passive {
				break
			} else if g.NodeAt(c2).nodeType == surveillance {
				g.MarkNodesAround(c2, -1)
			}
		}
	}

	fmt.Fprintf(os.Stderr, "Droping bomb at %v %v\n", c.i, c.j)
	fmt.Printf("%v %v\n", c.j, c.i) // Write action to stdout
}

func (g *Grid) String() string {
	var buffer bytes.Buffer

	for _, line := range g.nodes {
		for _, cell := range line {
			buffer.WriteString(fmt.Sprintf("%v ", cell))
		}
		buffer.WriteString("\n")
	}
	return buffer.String()
}

func main() {
	grid := new(Grid)

	grid.AcquireCells()

	for {
		// rounds: number of rounds left before the end of the game
		// bombs: number of bombs left
		var rounds, bombs int

		fmt.Scan(&rounds, &bombs)

		fmt.Fprintf(os.Stderr, "round=%v bombs=%v\n", rounds, bombs)
		fmt.Fprintf(os.Stderr, "%v", grid)
		bestCoord := Coord{0, 0}

		for i, line := range grid.nodes {
			for j, cell := range line {
				if !cell.hasBomb && cell.nodeType == empty {
					if cell.surveillanceNodesInRange > grid.NodeAt(bestCoord).surveillanceNodesInRange {
						bestCoord = Coord{i, j}
					}
				}
			}
		}
		fmt.Fprintf(os.Stderr, "Best coord: %v %v\n", bestCoord.i, bestCoord.j)

		grid.DropBombAt(bestCoord)
	}
}
