package main

import (
	"fmt"
	"os"
	"sort"
)

const (
	nbMolecules      = 5
	maxHeldMolecules = 10
	maxHeldSamples   = 3
	projectValue     = 50
)

type Molecules [nbMolecules]int

type Player struct {
	target             Module
	eta, score         int
	storage, expertise Molecules
	heldSamples        Samples
}

func (m *Molecules) Acquire() {
	for i := 0; i < nbMolecules; i++ {
		fmt.Scan(&m[i])
	}
}

func moleculeType(idx int) byte {
	return 'A' + byte(idx)
}

func (m Molecules) String() string {
	out := ""
	for i := 0; i < nbMolecules; i++ {
		out += fmt.Sprintf("%v ", m[i])
	}
	return out
}
func (m Molecules) Sum() int {
	sum := 0
	for i := 0; i < nbMolecules; i++ {
		sum += m[i]
	}
	return sum
}
func (m Molecules) Max() int {
	max := 0
	for i := 0; i < nbMolecules; i++ {
		if m[i] > max {
			max = m[i]
		}
	}
	return max
}

func (p *Player) Acquire() {
	var rawTarget string
	fmt.Scan(&rawTarget, &p.eta, &p.score)
	p.target = name2Module[rawTarget]
	p.storage.Acquire()
	p.expertise.Acquire()
	p.heldSamples = make(Samples, 0)
}

func (p Player) String() string {
	return fmt.Sprintf("Player: %v %v %v storage={%v} expertise={%v}",
		moduleNames[p.target], p.eta, p.score, p.storage, p.expertise)
}

type Sample struct {
	id, carriedBy, rank int
	expertiseGain       int
	health              int
	cost                Molecules
}

func (s *Sample) Acquire() {
	var rawExpertiseGain string
	fmt.Scan(&s.id, &s.carriedBy, &s.rank, &rawExpertiseGain, &s.health)
	s.expertiseGain = int(rawExpertiseGain[0] - 'A')
	s.cost.Acquire()
}

func (s Sample) String() string {
	return fmt.Sprintf("Sample %v: %v r=%v g=%c %v cost={%v}",
		s.id, s.carriedBy, s.rank, moleculeType(s.expertiseGain), s.health, s.cost)
}

func (s Sample) Value() int {
	if s.carriedBy >= 0 {
		return 0
	}
	return s.health * 1000 / (s.cost.Sum() + s.cost.Max())
}

func (s Sample) IsDiagnosed() bool {
	return s.cost[0] >= 0
}

type Module int

const (
	startingPosition = iota
	samples          = iota
	diagnosis        = iota
	molecules        = iota
	laboratory       = iota
	nbModules        = iota
)

var distanceMatrix = [nbModules][nbModules]int{
	{0, 2, 2, 2, 2},
	{2, 0, 3, 3, 3},
	{2, 3, 0, 3, 4},
	{2, 3, 3, 0, 3},
	{2, 3, 4, 3, 0}}

var moduleNames = [nbModules]string{
	"START_POS",
	"SAMPLES",
	"DIAGNOSIS",
	"MOLECULES",
	"LABORATORY"}

func makeName2Module() map[string]Module {
	res := make(map[string]Module)
	var m Module
	for m = startingPosition; m < nbModules; m++ {
		res[moduleNames[m]] = m
	}
	return res
}

var name2Module = makeName2Module()

func distance(a, b Module) int {
	return distanceMatrix[a][b]
}

type Samples []*Sample

func (a Samples) Len() int      { return len(a) }
func (a Samples) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a Samples) Less(i, j int) bool {
	return a[i].Value() > a[j].Value()
}

func GoTo(m Module) {
	fmt.Printf("GOTO %v\n", moduleNames[m])
}

func Connect(sampleId int) {
	fmt.Printf("CONNECT %v\n", sampleId)
}

func Gather(moleculeIdx int) {
	fmt.Printf("CONNECT %c\n", moleculeType(moleculeIdx))
}

type State struct {
	players           [2]Player
	available         Molecules
	currentSamples    Samples
	availableProjects []Molecules
}

func (s State) Me() Player {
	return s.players[0]
}
func (s State) Him() Player {
	return s.players[1]
}

func (s *State) Acquire() {
	for i := 0; i < 2; i++ {
		s.players[i].Acquire()
	}
	s.available.Acquire()

	var sampleCount int
	fmt.Scan(&sampleCount)

	s.currentSamples = make(Samples, sampleCount)

	for i := 0; i < sampleCount; i++ {
		sample := new(Sample)
		sample.Acquire()
		s.currentSamples[i] = sample
		carrier := s.currentSamples[i].carriedBy
		if carrier >= 0 {
			s.players[carrier].heldSamples = append(s.players[carrier].heldSamples, sample)
		}
	}
	s.availableProjects = make([]Molecules, 0)
	for _, project := range scienceProjects {
		onePlayerHasProject := false
		for _, player := range s.players {
			playerHasProject := true
			for i, mol := range player.expertise {
				if mol < project[i] {
					playerHasProject = false
					break
				}
			}
			if playerHasProject {
				onePlayerHasProject = true
				break
			}
		}
		if !onePlayerHasProject {
			s.availableProjects = append(s.availableProjects, project)
		}
	}
}

func (s State) String() string {
	var out string
	/* Print */
	for _, player := range s.players {
		out += fmt.Sprintln(player)
	}
	for _, project := range s.availableProjects {
		out += fmt.Sprintf("project: %v\n", project)
	}
	out += fmt.Sprintf("available: %v\n", s.available)
	for i := 0; i < len(s.currentSamples); i++ {
		out += fmt.Sprintln(s.currentSamples[i])
	}
	return out
}

func (p Player) MinDistanceTo(m Module) int {
	d := p.eta
	if p.target != m {
		d += distance(p.target, m)
	}
	return d
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func Subtract(a, b Molecules) Molecules {
	var res Molecules
	for i := 0; i < nbMolecules; i++ {
		res[i] = a[i] - b[i]
	}
	return res
}
func Add(a, b Molecules) Molecules {
	var res Molecules
	for i := 0; i < nbMolecules; i++ {
		res[i] = a[i] + b[i]
	}
	return res
}

func Max(a, b Molecules) Molecules {
	var res Molecules
	for i := 0; i < nbMolecules; i++ {
		res[i] = max(a[i], b[i])
	}
	return res
}
func Min(a, b Molecules) Molecules {
	var res Molecules
	for i := 0; i < nbMolecules; i++ {
		res[i] = min(a[i], b[i])
	}
	return res
}

var zero Molecules
var one Molecules = Molecules{1, 1, 1, 1, 1}
var two Molecules = Molecules{2, 2, 2, 2, 2}
var three Molecules = Molecules{3, 3, 3, 3, 3}
var four Molecules = Molecules{4, 4, 4, 4, 4}

func (s *Sample) CostForPlayer(p Player) Molecules {
	var zero Molecules
	return Max(zero, Subtract(s.cost, Add(p.expertise, p.storage)))
	return s.cost
}

func (s State) CostInThisOrder(playerIdx int, sampleSet Samples) Molecules {
	var cost Molecules
	p := s.players[playerIdx]
	for _, sample := range sampleSet {
		cost = Add(cost, Max(zero, Subtract(sample.cost, p.expertise)))
		p.expertise[sample.expertiseGain]++
	}
	return cost
}

func (s State) canCompleteInThisOrder(playerIdx int, sampleSet Samples) (bool, Molecules) {
	p := s.players[playerIdx]

	cost := s.CostInThisOrder(playerIdx, sampleSet)
	needed := Max(zero, Subtract(cost, p.storage))

	if needed.Sum()+p.storage.Sum() > maxHeldMolecules {
		return false, needed
	}

	for i := 0; i < nbMolecules; i++ {
		if needed[i] > s.available[i] {
			return false, needed
		}
	}

	//s.Him().MinDistanceTo(molecules)

	return true, needed
}

func (s Samples) String() string {
	ids := make([]int, len(s))
	for i, sample := range s {
		ids[i] = sample.id
	}
	return fmt.Sprint(ids)
}

type Step struct {
	completed Samples
	needed    Molecules
}

type Steps []Step

func (steps Steps) String() string {
	var out string
	for i, step := range steps {
		out += fmt.Sprintf("step %v: %v needed=%v\n", i, step.completed, step.needed)
	}
	return out
}

func LowerOrEqual(a, b Molecules) bool {
	for i := range a {
		if a[i] > b[i] {
			return false
		}
	}
	return true
}

func (s State) bestCompleteInThisOrder(playerIdx int, sampleSet Samples) (steps Steps) {
	p := s.players[playerIdx]
	firstSampleIdx := 0
	for {
		var needed Molecules
		var cost Molecules
		var lastSampleIdx int
		/* try to put as many samples as possible per step */
		for lastSampleIdx = firstSampleIdx; lastSampleIdx < len(sampleSet); lastSampleIdx++ {
			sample := sampleSet[lastSampleIdx]
			newCost := Add(cost, Max(zero, Subtract(sample.cost, p.expertise)))
			newNeeded := Max(zero, Subtract(newCost, p.storage))
			if newNeeded.Sum()+p.storage.Sum() > maxHeldMolecules ||
				!LowerOrEqual(newNeeded, s.available) {
				fmt.Fprintf(os.Stderr, "Cannot add sample %v to step %v\n", lastSampleIdx, len(steps))
				break
			} else {
				fmt.Fprintf(os.Stderr, "add sample %v to step %v\n", lastSampleIdx, len(steps))
				needed = newNeeded
				cost = newCost
				p.expertise[sample.expertiseGain]++
			}
		}
		fmt.Fprintf(os.Stderr, "firstSampleIdx=%v lastSampleIdx=%v\n", firstSampleIdx, lastSampleIdx)
		if lastSampleIdx > firstSampleIdx {
			fmt.Fprintf(os.Stderr, "create step %v\n", len(steps))
			steps = append(steps, Step{sampleSet[firstSampleIdx:lastSampleIdx], needed})
			firstSampleIdx = lastSampleIdx
			p.storage = Add(p.storage, needed)
			s.available = Subtract(s.available, needed)
			p.storage = Subtract(p.storage, cost)
			s.available = Add(s.available, cost)
			/* try another step */
		} else {
			break
		}
	}
	return
}

func (s State) StepsValue(playerIdx int, steps Steps) float64 {
	p := s.players[playerIdx]
	nbSamples := 0
	gain := 0
	cost := 0
	for stepIdx, step := range steps {
		for _, sample := range step.completed {
			gain += sample.health + s.moleculeValuesForProjects(p)[sample.expertiseGain]
			p.expertise[sample.expertiseGain]++
			/* expertise contributes to future purchase */
			switch p.expertise[sample.expertiseGain] {
			case 1:
				gain += 8
			case 2:
				gain += 6
			case 3:
				gain += 3
			}
		}
		cost += step.needed.Sum()
		if stepIdx > 0 {
			/* round-trip laboratory<->molecules */
			cost += 2 * distance(laboratory, molecules)
		}
	}

	cost += distance(laboratory, samples)
	cost += nbSamples
	cost += distance(samples, diagnosis)
	cost += nbSamples
	cost += distance(diagnosis, molecules)

	return float64(gain) / float64(cost)
}

func (s State) bestComplete(playerIdx int, sampleSet Samples) (bestSteps Steps) {
	permut := makePermutation(len(sampleSet))
	bestValue := 0.0
	for {
		samplePermut := permut.Reordered(sampleSet)
		steps := s.bestCompleteInThisOrder(playerIdx, samplePermut)
		for stepIdx := range steps {
			subSteps := steps[:stepIdx+1]
			value := s.StepsValue(playerIdx, subSteps)
			fmt.Fprintf(os.Stderr, "%vvalue=%v\n", subSteps, value)
			if value > bestValue {
				bestSteps = subSteps
				bestValue = value
			}
		}
		if !permut.Next() {
			break
		}
	}
	fmt.Fprintf(os.Stderr, "best steps:\n%vbestValue=%v\n", bestSteps, bestValue)
	return
}

func (m Molecules) MinValue() int {
	min := m[0]
	for i := 1; i < nbMolecules; i++ {
		if m[i] < min {
			min = i
		}
	}
	return min
}

func (s State) isBetterNeeded(neededA, neededB Molecules) bool {
	sa, sb := neededA.Sum(), neededB.Sum()
	if sa != sb {
		return sa < sb
	}

	_, wA := s.moleculeToPickFirst(neededA)
	_, wB := s.moleculeToPickFirst(neededB)
	return wA < wB
}

func (s State) canComplete(playerIdx int, sampleSet Samples) (Samples, Molecules) {
	p := makePermutation(len(sampleSet))
	bestNeeded := zero
	var bestPermut Samples = nil
	for {
		samplePermut := p.Reordered(sampleSet)
		can, needed := s.canCompleteInThisOrder(playerIdx, samplePermut)
		fmt.Fprintf(os.Stderr, "permut=%v can=%v, needed=%v\n", p, can, needed)
		if can {
			if bestPermut == nil || s.isBetterNeeded(needed, bestNeeded) {
				bestPermut = samplePermut
				bestNeeded = needed
			}

		}
		if !p.Next() {
			break
		}
	}
	if bestPermut != nil {
		return bestPermut, bestNeeded
	}
	return nil, zero
}

/* assumes than the given samples can be completed */
func (s State) moleculeToPickFirst(needed Molecules) (index int, weight int) {
	maxWeight := -1000
	maxIdx := -1
	for i := 0; i < nbMolecules; i++ {
		if needed[i] > 0 && s.available[i] > 0 {
			weight := 2*needed[i] - s.available[i]
			if weight > maxWeight {
				maxWeight = weight
				maxIdx = i
			}
		}
	}
	return maxIdx, maxWeight
}

type NPCombi struct {
	n, p  int
	combi []int
}

func makeNPCombi(n, p int) (c NPCombi) {
	c.n = n
	c.p = p
	c.combi = make([]int, n)
	for i := 0; i < n; i++ {
		c.combi[i] = i
	}
	return
}

/*
3 among 5
[0 1 2]
[0 1 3]
[0 1 4]
[0 2 3]
[0 2 4]
[0 3 4]
[1 2 3]
[1 2 4]
[1 3 4]
[2 3 4]
*/
func (c *NPCombi) Next() bool {
	for i := c.n - 1; i >= 0; i-- {
		if c.combi[i] < c.p-(c.n-i) {
			c.combi[i]++
			for j := i + 1; j < c.n; j++ {
				c.combi[j] = c.combi[j-1] + 1
			}
			return true
		}
	}
	return false
}

func (c NPCombi) subset(s Samples) (sub Samples) {
	sub = make(Samples, c.n)
	for i := range c.combi {
		sub[i] = s[c.combi[i]]
	}
	return
}

type XPCombi struct {
	NPCombi
}

func makeXPCombi(p int) (c XPCombi) {
	c.NPCombi = makeNPCombi(p, p)
	return
}

/*
[0 1 2 3 4]
[0 1 2 3]
[0 1 2 4]
[0 1 3 4]
[0 2 3 4]
[1 2 3 4]
[0 1 2]
[0 1 3]
[0 1 4]
[0 2 3]
[0 2 4]
[0 3 4]
[1 2 3]
[1 2 4]
[1 3 4]
[2 3 4]
[0 1]
[0 2]
[0 3]
[0 4]
[1 2]
[1 3]
[1 4]
[2 3]
[2 4]
[3 4]
[0]
[1]
[2]
[3]
[4]
*/
func (c *XPCombi) Next() bool {
	if c.NPCombi.Next() {
		return true
	} else if c.NPCombi.n <= 1 {
		return false
	} else {
		c.NPCombi = makeNPCombi(c.NPCombi.n-1, c.NPCombi.p)
	}
	return true
}

type Permutation []int

func (a Permutation) Len() int      { return len(a) }
func (a Permutation) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a Permutation) Less(i, j int) bool {
	return a[i] < a[j]
}

func makePermutation(n int) (p Permutation) {
	p = make(Permutation, n)
	for i := range p {
		p[i] = i
	}
	return p
}

/*
[0 1 2 3]
[0 1 3 2]
[0 2 1 3]
[0 2 3 1]
[0 3 1 2]
[0 3 2 1]
[1 0 2 3]
[1 0 3 2]
[1 2 0 3]
[1 2 3 0]
[1 3 0 2]
[1 3 2 0]
[2 0 1 3]
[2 0 3 1]
[2 1 0 3]
[2 1 3 0]
[2 3 0 1]
[2 3 1 0]
[3 0 1 2]
[3 0 2 1]
[3 1 0 2]
[3 1 2 0]
[3 2 0 1]
[3 2 1 0]

*/
func (p Permutation) Next() bool {
	n := len(p)
	for i := n - 2; i >= 0; i-- {
		if p[i] < p[i+1] {
			minIdx := i + 1
			for j := i + 2; j < n; j++ {
				if p[j] > p[i] && p[j] < p[minIdx] {
					minIdx = j
				}
			}
			p.Swap(i, minIdx)
			sort.Sort(p[i+1:])
			return true
		}
	}
	return false
}

/* assumes that len(s) >= len(p) */
func (p Permutation) Reordered(s Samples) Samples {
	res := make(Samples, len(p))
	for i := range p {
		res[i] = s[p[i]]
	}
	return res
}

func (s State) samplesThatCanBeCompleted(playerIdx int) (selection Samples, needed Molecules) {
	xpc := makeXPCombi(len(s.players[playerIdx].heldSamples))
	for {
		samples := xpc.subset(s.players[playerIdx].heldSamples)
		orderedSamples, needed := s.canComplete(playerIdx, samples)
		fmt.Fprintf(os.Stderr, "samples=%v\ncan=%v needed=%v\n", samples, orderedSamples != nil, needed)
		if orderedSamples != nil {
			if needed.Sum() > 0 {
				/* TODO: select combi with highest value?
				Useless because heldSamples are sorted by value?
				With more interesting expertise regarding projects ?*/
				return orderedSamples, needed
				break
			}
		}
		if !xpc.Next() {
			break
		}
	}
	return nil, zero
}

func (s State) completeSamples(playerIdx int) Samples {
	xpc := makeXPCombi(len(s.players[playerIdx].heldSamples))
	for {
		samples := xpc.subset(s.players[playerIdx].heldSamples)
		orderedSamples, needed := s.canComplete(playerIdx, samples)
		if orderedSamples != nil && needed.Sum() == 0 {
			return orderedSamples
		}
		if !xpc.Next() {
			break
		}
	}
	return nil
}

func (p Player) nbRankHeld(rank int) (nb int) {
	for _, sample := range p.heldSamples {
		if sample.rank == rank {
			nb++
		}
	}
	return
}

var scienceProjects []Molecules

func (s State) moleculeValuesForProject(player Player, project Molecules) (value Molecules) {
	totalExpertiseNeeded := project.Sum()
	expertiseAlreadyGained := Min(player.expertise, project)
	expertiseNeeded := Subtract(project, expertiseAlreadyGained)

	oneMoleculeValue := projectValue * (expertiseAlreadyGained.Sum() + 1) / totalExpertiseNeeded

	for i := range value {
		if expertiseNeeded[i] > 0 {
			value[i] = oneMoleculeValue
		}
	}
	return
}
func (s State) moleculeValuesForProjects(player Player) (value Molecules) {
	for _, project := range s.availableProjects {
		value = Add(value, s.moleculeValuesForProject(player, project))
	}
	return
}

/**
 * Bring data on patient samples from the diagnosis machine to the laboratory with enough molecules to produce medicine!
 **/

func main() {
	if false {
		npCombi := makeNPCombi(3, 5)
		for {
			fmt.Fprintln(os.Stderr, npCombi.combi)
			if !npCombi.Next() {
				break
			}
		}

		list := make(Samples, 5)
		for i := range list {
			list[i] = new(Sample)
			list[i].health = i
		}

		xpCombi := makeXPCombi(5)
		for {
			fmt.Fprintln(os.Stderr, xpCombi.combi)
			fmt.Fprintln(os.Stderr, xpCombi.subset(list))
			if !xpCombi.Next() {
				break
			}
		}

		p := makePermutation(3)
		for {
			fmt.Fprintln(os.Stderr, p)
			fmt.Fprintln(os.Stderr, p.Reordered(list))
			if !p.Next() {
				break
			}
		}

		return
	}

	var projectCount int
	fmt.Scan(&projectCount)

	scienceProjects = make([]Molecules, projectCount)

	for i := 0; i < projectCount; i++ {
		scienceProjects[i].Acquire()
	}
	for {

		var currentState State
		currentState.Acquire()

		var me Player = currentState.Me()

		fmt.Fprintf(os.Stderr, "value for me : %v\n", currentState.moleculeValuesForProjects(me))
		fmt.Fprintf(os.Stderr, "value for him: %v\n", currentState.moleculeValuesForProjects(currentState.Him()))
		fmt.Fprintln(os.Stderr, currentState)

		switch me.target {
		case startingPosition:
			GoTo(samples)
		case samples:
			nbHeld := len(me.heldSamples)
			if nbHeld < maxHeldSamples {
				rank := 1
				totalExpertiseBelowFour := Min(me.expertise, three).Sum()
				if totalExpertiseBelowFour > 2 && me.nbRankHeld(2) == 0 ||
					totalExpertiseBelowFour > 3 && me.nbRankHeld(2) <= 1 ||
					totalExpertiseBelowFour > 4 {
					rank = 2
				}
				if totalExpertiseBelowFour > 7 && me.nbRankHeld(3) == 0 ||
					totalExpertiseBelowFour > 9 && me.nbRankHeld(3) <= 1 ||
					totalExpertiseBelowFour > 11 {
					rank = 3
				}
				Connect(rank)
			} else {
				GoTo(diagnosis)
			}
		case diagnosis:
			actionDone := false
			for _, sample := range me.heldSamples {
				if !sample.IsDiagnosed() {
					Connect(sample.id)
					actionDone = true
					break
				}
			}
			if !actionDone && len(me.heldSamples) > 0 {
				future := currentState
				if currentState.Him().target == laboratory {
					opponentCompleteSamples := currentState.completeSamples(1)
					if len(opponentCompleteSamples) > 0 {
						future.available = Add(future.available, currentState.CostInThisOrder(1, opponentCompleteSamples))
					}
				}
				steps := currentState.bestComplete(0, me.heldSamples)
				if len(steps) == 0 {
					/* dump samples */
					Connect(me.heldSamples[0].id)
					actionDone = true
				}
			}
			if !actionDone && len(me.heldSamples) == 0 {
				GoTo(samples)
				actionDone = true
			}

			if !actionDone {
				GoTo(molecules)
			}
		case molecules:
			actionDone := false
			oneIsComplete := false
			if !actionDone {
				steps := currentState.bestComplete(0, me.heldSamples)
				var cumulNeeded Molecules
				for i, step := range steps {
					cumulNeeded = Add(cumulNeeded, step.needed)
					if cumulNeeded.Sum() == 0 {
						fmt.Fprintf(os.Stderr, "step %v is complete\n", i)
						oneIsComplete = true
					} else {
						m, _ := currentState.moleculeToPickFirst(cumulNeeded)
						if m >= 0 {
							fmt.Fprintf(os.Stderr, "gather molecule for step %v\n", i)
							Gather(m)
							actionDone = true
						} else {
							fmt.Fprintf(os.Stderr, "CAN'T gather molecule for step %v\n", i)
						}
						break
					}
				}
			}
			if !actionDone && oneIsComplete {
				GoTo(laboratory)
				actionDone = true
			}
			if !actionDone && currentState.Him().target == laboratory {
				opponentCompleteSamples := currentState.completeSamples(1)
				if len(opponentCompleteSamples) > 0 {
					/* opponent is going to the laboratory with one or more complete samples */
					future := currentState
					future.available = Add(future.available, currentState.CostInThisOrder(1, opponentCompleteSamples))
					futureSamplesToComplete, needed := future.samplesThatCanBeCompleted(0)
					if len(futureSamplesToComplete) > 0 {
						pick, _ := currentState.moleculeToPickFirst(needed)
						if pick >= 0 {
							Gather(pick)
						} else {
							fmt.Println("WAIT")
						}
						actionDone = true
					}
				}
			}
			if !actionDone {
				if len(me.heldSamples) < maxHeldSamples {
					GoTo(samples)
				} else {
					GoTo(diagnosis)
				}
			}
		case laboratory:
			actionDone := false
			if !actionDone {
				sampleSet := currentState.completeSamples(0)
				if len(sampleSet) > 0 {
					Connect(sampleSet[0].id)
					actionDone = true
				}
			}
			if !actionDone {
				sampleSet, _ := currentState.samplesThatCanBeCompleted(0)
				if len(sampleSet) > 0 {
					GoTo(molecules)
					actionDone = true
					break
				}
			}
			if !actionDone {
				GoTo(samples)
			}
		}
	}
}
