package main

import "fmt"

import "os"

type Link struct {
	statusIdx int
	a, b      *Node
}

func (l Link) IsSevered() bool {
	return state.linkStatus[l.statusIdx] != '_'
}
func (l Link) Sever() {
	state.linkStatus = state.linkStatus[:l.statusIdx] + "x" + state.linkStatus[l.statusIdx+1:]
}

type Connexion struct {
	link  *Link
	other *Node
}

func (con Connexion) String() string {
	return fmt.Sprintf("%c%d", state.linkStatus[con.link.statusIdx], con.other.id)
}

type Node struct {
	connexions []Connexion
	isExit     bool
	id         int
}

func (node Node) String() string {
	return fmt.Sprintf("%2d: %v", node.id, node.connexions)
}
func (node Node) nbExitArount() int {
	res := 0
	for _, con := range node.connexions {
		if !con.link.IsSevered() && con.other.isExit {
			res++
		}
	}
	return res
}
func (node Node) GetOneLinkToExit() *Link {
	for _, con := range node.connexions {
		if !con.link.IsSevered() && con.other.isExit {
			return con.link
		}
	}
	return nil
}
func (node Node) nbNonExitArount() int {
	res := 0
	for _, con := range node.connexions {
		if !con.link.IsSevered() && !con.other.isExit {
			res++
		}
	}
	return res
}

type State struct {
	linkStatus string
	skynetNode *Node
}

var nodes []Node

var state State
var links []Link

func saveState() State {
	return state
}
func restoreState(backup State) {
	state = backup
}

func connectNodes(a *Node, b *Node) {
	state.linkStatus = state.linkStatus + "_"
	links = append(links, Link{len(state.linkStatus) - 1, a, b})
	link := &links[len(links)-1]
	a.connexions = append(a.connexions, Connexion{link, b})
	b.connexions = append(b.connexions, Connexion{link, a})
}

func main() {
	// N: the total number of nodes in the level, including the gateways
	// L: the number of links
	// E: the number of exit gateways
	var N, L, E int
	fmt.Scan(&N, &L, &E)

	nodes = make([]Node, N)
	links = make([]Link, 0, L)

	for i := range nodes {
		nodes[i].id = i
	}

	for i := 0; i < L; i++ {
		// N1: N1 and N2 defines a link between these nodes
		var N1, N2 int
		fmt.Scan(&N1, &N2)
		connectNodes(&nodes[N1], &nodes[N2])
	}
	for i := 0; i < E; i++ {
		// EI: the index of a gateway node
		var EI int
		fmt.Scan(&EI)
		nodes[EI].isExit = true
	}
	for {
		// SI: The index of the node on which the Skynet agent is positioned this turn
		var SI int
		fmt.Scan(&SI)

		state.skynetNode = &nodes[SI]

		for _, node := range nodes {
			fmt.Fprintln(os.Stderr, node)
		}
		backup := saveState()

		nbSaved := make(map[*Node]int)
		stateFIFO := make([]State, 0)
		stateFIFO = append(stateFIFO, state)
		nbSaved[state.skynetNode] = 0

		var linkToSever *Link
		safety := 1000

		for i := 0; i < len(stateFIFO); i++ {
			currentState := stateFIFO[i]
			restoreState(currentState)
			currentSaved := nbSaved[currentState.skynetNode]
			if currentState.skynetNode.nbExitArount() > 0 {
				//fmt.Fprintf(os.Stderr,"%v? nbE=%v saved=",currentState.skynetNode)
				if currentSaved-currentState.skynetNode.nbExitArount() < safety {
					linkToSever = currentState.skynetNode.GetOneLinkToExit()
					safety = currentSaved-currentState.skynetNode.nbExitArount()
					if safety == 1 {
						break
					}
				}
			}

			currentSaved += 1 - currentState.skynetNode.nbExitArount()
			for _, con := range currentState.skynetNode.connexions {
				if !con.link.IsSevered() && con.other.isExit {
					con.link.Sever()
				}
			}
			for _, con := range currentState.skynetNode.connexions {
				if !con.link.IsSevered() && !con.other.isExit {
					nextNode := con.other
					if nbSaved[nextNode] == 0 || nbSaved[nextNode] > currentSaved {
						nbSaved[nextNode] = currentSaved
						nextState := saveState()
						nextState.skynetNode = nextNode
						stateFIFO = append(stateFIFO,nextState)
					}
				}
			}

		}

		restoreState(backup)

		linkToSever.Sever()

		// fmt.Fprintln(os.Stderr, "Debug messages...")

		fmt.Printf("%v %v\n",linkToSever.a.id,linkToSever.b.id) // Example: 3 4 are the indices of the nodes you wish to sever the link between
	}
}
