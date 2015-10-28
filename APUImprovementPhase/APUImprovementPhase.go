package main

import "fmt"
import "os"
import "bufio"
import "sort"

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

type Grid [][]*Node

type Link struct {
	count *int //should never exceed 2
	cross []*Link
	a, b  *Node
}

type Connexion struct {
	link  *Link
	other *Node
}
type Connexions []Connexion

type Node struct {
	connexions     Connexions
	optimalNbLinks int
	r, c           int //coords
}

func (connexions Connexions) Len() int { return len(connexions) }
func (connexions Connexions) Swap(i, j int) {
	connexions[i], connexions[j] = connexions[j], connexions[i]
}
func (connexions Connexions) Less(i, j int) bool {
	return connexions[i].AssignableSlots() > connexions[j].AssignableSlots()
}

func (con Connexion) RemainingSlots() int {
	return 2 - con.TakenSlots()
}
func (con Connexion) AssignableSlots() int {
	for _, cross := range con.link.cross {
		if *(cross.count) > 0 {
			return 0
		}
	}
	return Min(con.RemainingSlots(), con.other.NbLinksToAssign())
}
func (con Connexion) TakenSlots() int {
	return *(con.link.count)
}
func (con Connexion) SetTakenSlots(nb int) {
	*(con.link.count) = nb
}
func (con Connexion) String() string {
	s := fmt.Sprintf("%v", con.TakenSlots())
	if len(con.link.cross) > 0 {
		s = fmt.Sprintf("%vx%v", s, len(con.link.cross))
	}
	return s
}

func (node Node) String() string {
	return fmt.Sprintf("(%v,%v) Links:%v/%v Remaining:%v/%v - %v Connexions: %v",
		node.c,
		node.r,
		node.NbLinksAssigned(),
		node.optimalNbLinks,
		node.NbLinksToAssign(),
		node.NbLinksCanBeAssigned(),
		len(node.connexions),
		node.connexions)
}
func (node Node) NbLinksAssigned() int {
	res := 0
	for _, link := range node.connexions {
		res += link.TakenSlots()
	}
	return res
}
func (node Node) NbLinksToAssign() int {
	return node.optimalNbLinks - node.NbLinksAssigned()
}
func (node Node) NbLinksCanBeAssigned() int {
	res := 0
	for _, link := range node.connexions {
		res += link.AssignableSlots()
	}
	return res
}

var state []int
var links []Link

func saveState() []int {
	res := make([]int, len(state))
	copy(res, state)
	return res
}
func restoreState(backup []int) {
	copy(state, backup)
}

func connectNodes(a *Node, b *Node) {
	state = append(state, 0)
	l := &state[len(state)-1]
	links = append(links, Link{l, nil, a, b})
	link := &links[len(links)-1]
	a.connexions = append(a.connexions, Connexion{link, b})
	b.connexions = append(b.connexions, Connexion{link, a})
}

type Nodes []*Node

func (nodes Nodes) setTrivialLinks() {
	pass := 0
	/* trivial links because mandatory */
	for somethingDone := true; somethingDone; {
		pass++
		//fmt.Fprintf(os.Stderr, "pass %v:\n", pass)
		somethingDone = false
		for _, node := range nodes {
			if node.NbLinksToAssign() > 0 &&
				node.NbLinksToAssign() == node.NbLinksCanBeAssigned() {
				for _, con := range node.connexions {
					if con.AssignableSlots() > 0 {
						con.SetTakenSlots(Min(2, con.other.NbLinksToAssign()))
					}
				}
				somethingDone = true
			}

			if node.NbLinksToAssign() == 1 {
				var uniqueConnexion Connexion
				nonFullConnexions := 0
				for _, con := range node.connexions {
					if con.AssignableSlots() > 0 {
						nonFullConnexions++
						if nonFullConnexions > 1 {
							break
						}
						uniqueConnexion = con
					}
				}
				if nonFullConnexions == 1 {
					uniqueConnexion.SetTakenSlots(1 + uniqueConnexion.TakenSlots())
					somethingDone = true
				}
			}
		}
	}
}

func (nodes Nodes) Print() {
	fmt.Fprintln(os.Stderr, state)
	for _, n := range nodes {
		fmt.Fprintln(os.Stderr, n)
	}
}

func (nodes Nodes) totalRemaining() int {
	sum := 0
	for _, n := range nodes {
		sum += n.NbLinksToAssign()
	}
	return sum
}

func diffless(a, b int) (diff, less bool) {
	return a != b, a < b
}

func (nodes Nodes) Len() int      { return len(nodes) }
func (nodes Nodes) Swap(i, j int) { nodes[i], nodes[j] = nodes[j], nodes[i] }
func (nodes Nodes) Less(i, j int) bool {
	if nodes[i].NbLinksToAssign() == 0 {
		return false
	}
	if nodes[j].NbLinksToAssign() == 0 {
		return true
	}
	if diff, less := diffless(nodes[i].NbLinksCanBeAssigned()-nodes[i].NbLinksToAssign(),
		nodes[j].NbLinksCanBeAssigned()-nodes[j].NbLinksToAssign()); diff {
		return less
	}
	return nodes[i].NbLinksToAssign() < nodes[j].NbLinksToAssign()
}

func (nodes Nodes) areAllConnected() bool {
	buffer := make(Nodes, 0)
	treatedNodes := make(map[*Node]bool)
	buffer = append(buffer, nodes[0])
	treatedNodes[nodes[0]] = true
	for i := 0; i < len(buffer); i++ {
		for _, con := range buffer[i].connexions {
			if con.TakenSlots() > 0 {
				if !treatedNodes[con.other] {
					buffer = append(buffer, con.other)
					treatedNodes[con.other] = true
				}
			}
		}
	}
	return len(buffer) == len(nodes)
}

func (nodes Nodes) findSolution_rec() bool {
	nodes.setTrivialLinks()

	if nodes.totalRemaining() == 0 {
		for _, n := range nodes {
			if n.NbLinksAssigned() != n.optimalNbLinks {
				return false
			}
		}
		return nodes.areAllConnected()
	} else {
		sort.Sort(nodes)
		//nodes.Print()

		currentNode := nodes[0]

		if currentNode.NbLinksToAssign() > currentNode.NbLinksCanBeAssigned() {
			return false
		}

		sort.Sort(currentNode.connexions)
		nbCon := 0
		for _, con := range currentNode.connexions {
			if con.AssignableSlots() > 0 {
				nbCon++
			} else {
				break
			}
		}

		maxes := [5]int{0, 0, 0, 0, 0}
		distrib := [5]int{0, 0, 0, 0, 0}
		nbToDistribute := currentNode.NbLinksToAssign()
		distributed := 0
		/* initial Distrib */
		for i, con := range currentNode.connexions {
			maxes[i] = con.AssignableSlots()
			assigned := Min(maxes[i], nbToDistribute-distributed)
			distrib[i] = assigned
			distributed += assigned
		}

		for {
			backup := saveState()

			//fmt.Fprintf(os.Stderr,"nb=%v distrib=%v  maxes=%v\n",nbToDistribute,distrib,maxes)

			/* apply distrib */
			for i, con := range currentNode.connexions {
				con.SetTakenSlots(distrib[i] + con.TakenSlots())
			}

			if nodes.findSolution_rec() {
				return true
			} else {
				restoreState(backup)
				/* next distrib */
				toRedispatch := 0
				var i int
				for i = nbCon - 1; i >= 0; i-- {
					/* retrieve all that cannot be shifted to the right */
					if maxes[i+1]-distrib[i+1] == 0 {
						toRedispatch += distrib[i]
					} else {
						break
					}
				}
				if toRedispatch == nbToDistribute {
					break
				}
				for distrib[i] == 0 {
					i--
				}
				distrib[i]--
				toRedispatch++
				i++
				for ;i<nbCon; i++ {
					nb := Min(maxes[i], toRedispatch)
					distrib[i] = nb
					toRedispatch -= nb
				}
			}

		}

	}

	return false
}

func main() {

	scanner := bufio.NewScanner(os.Stdin)

	// width: the number of cells on the X axis
	var width int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &width)

	// height: the number of cells on the Y axis
	var height int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &height)

	grid := make(Grid, height)
	nodes := make(Nodes, 0)
	state = make([]int, 0, width*height*2)
	links = make([]Link, 0, width*height*2)

	for i := range grid {
		scanner.Scan()
		chars := scanner.Text() // width characters, each either a number or a '.'

		grid[i] = make([]*Node, width)
		for j := range grid[i] {
			if chars[j] != '.' {
				node := new(Node)
				node.optimalNbLinks = int(chars[j] - '0')
				node.r = i
				node.c = j
				grid[i][j] = node
				nodes = append(nodes, grid[i][j])
			}
		}
	}

	//Print
	for _, line := range grid {
		for _, node := range line {
			if node == nil {
				fmt.Fprintf(os.Stderr, "%c", '.')
			} else {
				fmt.Fprint(os.Stderr, node.optimalNbLinks)
			}
		}
		fmt.Fprintln(os.Stderr)
	}

	/* connect nodes */
	for r := 0; r < height; r++ {
		var previousNode *Node = nil
		for c := 0; c < width; c++ {
			if grid[r][c] != nil {
				if previousNode != nil {
					connectNodes(previousNode, grid[r][c])
				}
				previousNode = grid[r][c]
			}
		}
	}
	for c := 0; c < width; c++ {
		var previousNode *Node = nil
		for r := 0; r < height; r++ {
			if grid[r][c] != nil {
				if previousNode != nil {
					connectNodes(previousNode, grid[r][c])
				}
				previousNode = grid[r][c]
			}
		}
	}

	/* find crossing links */
	for i := range links {
		for j := range links {
			if links[i].a.c == links[i].b.c {
				/* vertical link */
				if links[j].a.r == links[j].b.r {
					if links[j].a.r > links[i].a.r &&
						links[j].a.r < links[i].b.r &&
						links[i].a.c > links[j].a.c &&
						links[i].a.c < links[j].b.c {
						links[i].cross = append(links[i].cross, &links[j])
						//fmt.Fprintf(os.Stderr, "CROSS! %v\n", len(links[i].cross))
					}
				}
			} else {
				/* horizontal */
				if links[j].a.c == links[j].b.c {
					if links[j].a.c > links[i].a.c &&
						links[j].a.c < links[i].b.c &&
						links[i].a.r > links[j].a.r &&
						links[i].a.r < links[j].b.r {
						links[i].cross = append(links[i].cross, &links[j])
						//fmt.Fprintf(os.Stderr, "CROSS! %v\n", len(links[i].cross))
					}
				}
			}
		}
	}

	nodes.Print()

	nodes.findSolution_rec()

	//Print all connexion
	for _, node := range nodes {
		for _, con := range node.connexions {
			if con.TakenSlots() > 0 {
				fmt.Printf("%d %d %d %d %d\n",
					node.c,
					node.r,
					con.other.c,
					con.other.r,
					con.TakenSlots())
				con.SetTakenSlots(0)
			}
		}
	}
}
