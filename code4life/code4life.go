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
	fmt.Fprintf(os.Stderr, "GOTO %v\n", moduleNames[m])
	fmt.Printf("GOTO %v\n", moduleNames[m])
}

func Connect(sampleId int) {
	fmt.Fprintf(os.Stderr, "CONNECT %v\n", sampleId)
	fmt.Printf("CONNECT %v\n", sampleId)
}

func Gather(moleculeIdx int) {
	fmt.Fprintf(os.Stderr, "CONNECT %c\n", moleculeType(moleculeIdx))
	fmt.Printf("CONNECT %c\n", moleculeType(moleculeIdx))
}

type State struct {
	players           [2]Player
	available         Molecules
	currentSamples    Samples
	availableProjects []Molecules
	previous          *State
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
		/* try to put as many samples as possible per step (max 3) */
		for lastSampleIdx = firstSampleIdx; lastSampleIdx < len(sampleSet) && lastSampleIdx%3 >= firstSampleIdx%3; lastSampleIdx++ {
			sample := sampleSet[lastSampleIdx]
			newCost := Add(cost, Max(zero, Subtract(sample.cost, p.expertise)))
			newNeeded := Max(zero, Subtract(newCost, p.storage))
			if newNeeded.Sum()+p.storage.Sum() > maxHeldMolecules ||
				!LowerOrEqual(newNeeded, s.available) {
				//fmt.Fprintf(os.Stderr, "Cannot add sample %v to step %v\n", lastSampleIdx, len(steps))
				break
			} else {
				//fmt.Fprintf(os.Stderr, "add sample %v to step %v\n", lastSampleIdx, len(steps))
				needed = newNeeded
				cost = newCost
				p.expertise[sample.expertiseGain]++
			}
		}
		//fmt.Fprintf(os.Stderr, "firstSampleIdx=%v lastSampleIdx=%v\n", firstSampleIdx, lastSampleIdx)
		if lastSampleIdx > firstSampleIdx {
			//fmt.Fprintf(os.Stderr, "create step %v\n", len(steps))
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

	for _, samp := range p.heldSamples {
		if !samp.isInSteps(steps) {
			/* dump sample */
			cost += 1
		}
	}

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
			nbSamples++
			if sample.carriedBy != playerIdx {
				/* download sample */
				cost += 1
			}
		}
		cost += step.needed.Sum()
		if stepIdx > 0 {
			if nbSamples > 3 {
				nbSamples -= 3
				cost += distance(laboratory, diagnosis) + distance(diagnosis, molecules) + 3
			} else {
				/* round-trip laboratory<->molecules */
				cost += 2 * distance(laboratory, molecules)
			}
		}
		if stepIdx == 0 &&
			p.target == laboratory &&
			step.needed.Sum() > 0 {
			cost += p.eta + distance(laboratory, molecules)
		}
	}

	cost += distance(laboratory, samples)
	cost += distance(samples, diagnosis)
	cost += distance(diagnosis, molecules)

	return float64(gain) / float64(cost)
}

func (steps Steps) nbSamples() (nb int) {
	for _, step := range steps {
		nb += len(step.completed)
	}
	return
}

func (sampleSet Samples) filterUndiagnosed() (newSampleSet Samples) {
	newSampleSet = make(Samples, 0, len(sampleSet))
	for _, samp := range sampleSet {
		if samp.IsDiagnosed() {
			newSampleSet = append(newSampleSet, samp)
		}
	}
	return
}

func (s State) bestComplete(playerIdx int, sampleSet Samples) (bestSteps Steps) {
	/* safety */
	sampleSet = sampleSet.filterUndiagnosed()
	permut := makePermutation(len(sampleSet))
	bestValue := 0.0
	for {
		samplePermut := permut.Reordered(sampleSet)

		steps := s.bestCompleteInThisOrder(playerIdx, samplePermut)
		for stepIdx := range steps {
			subSteps := steps[:stepIdx+1]
			value := s.StepsValue(playerIdx, subSteps)
			if value > bestValue {
				bestSteps = subSteps
				bestValue = value
				fmt.Fprintf(os.Stderr, "permut:%v\n", permut)
				fmt.Fprintf(os.Stderr, "new best:\n%vvalue=%v\n", subSteps, value)
			}
		}
		if !permut.ForceNext(steps.nbSamples()) {
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
func (p Permutation) ForceNext(i int) bool {
	n := len(p)

	minIdx := -1
	for j := i + 1; j < n; j++ {
		if p[j] > p[i] &&
			(minIdx == -1 || p[j] < p[minIdx]) {
			minIdx = j
		}
	}
	if minIdx == -1 {
		if i > 0 {
			return p.ForceNext(i - 1)
		}
	} else {
		p.Swap(i, minIdx)
		sort.Sort(p[i+1:])
		return true
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

func (sample *Sample) isInSteps(steps Steps) bool {
	for _, step := range steps {
		for _, s := range step.completed {
			if s.id == sample.id {
				return true
			}
		}
	}
	return false
}

func (state State) PlayerIsWaiting(playerIdx int) bool {
	if state.previous != nil {
		p := state.players[playerIdx]
		pp := state.previous.players[playerIdx]
		if pp.target == p.target &&
			pp.eta == p.eta &&
			len(pp.heldSamples) == len(p.heldSamples) &&
			pp.storage.Sum() == p.storage.Sum() {
			return true
		}
	}
	return false
}

func (s State) AvailableAtLabo(playerIdx int) Samples {
	var availableSamples Samples
	for _, samp := range s.currentSamples {
		if samp.IsDiagnosed() &&
			(samp.carriedBy == playerIdx || samp.carriedBy < 0) {
			availableSamples = append(availableSamples, samp)
		}
	}
	return availableSamples
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

	var previousState *State = nil

	for {

		var currentState *State = new(State)
		currentState.Acquire()
		currentState.previous = previousState

		var me Player = currentState.Me()

		fmt.Fprintf(os.Stderr, "value for me : %v\n", currentState.moleculeValuesForProjects(me))
		fmt.Fprintf(os.Stderr, "value for him: %v\n", currentState.moleculeValuesForProjects(currentState.Him()))
		fmt.Fprintln(os.Stderr, currentState)

		switch me.target {
		case startingPosition:
			GoTo(samples)
		case samples:
			fmt.Fprintf(os.Stderr, "samples\n")
			nbHeld := len(me.heldSamples)
			if nbHeld < maxHeldSamples {
				rank := 1
				totalExpertiseBelowFour := Min(me.expertise, three).Sum()
				if totalExpertiseBelowFour > 3 && me.nbRankHeld(2) == 0 ||
					totalExpertiseBelowFour > 6 && me.nbRankHeld(2) <= 1 ||
					totalExpertiseBelowFour > 7 {
					rank = 2
				}
				if totalExpertiseBelowFour > 7 && me.nbRankHeld(3) == 0 ||
					totalExpertiseBelowFour > 10 && me.nbRankHeld(3) <= 1 ||
					totalExpertiseBelowFour > 12 {
					rank = 3
				}
				Connect(rank)
			} else {
				GoTo(diagnosis)
			}
		case diagnosis:
			fmt.Fprintf(os.Stderr, "diagnosis\n")
			actionDone := false
			for _, sample := range me.heldSamples {
				if !sample.IsDiagnosed() {
					fmt.Fprintf(os.Stderr, "diagnose %v\n", sample.id)
					Connect(sample.id)
					actionDone = true
					break
				}
			}
			if !actionDone {
				fmt.Fprintf(os.Stderr, "selection\n")
				future := *currentState
				if currentState.Him().target == laboratory ||
					currentState.Him().target == molecules {
					fmt.Fprintf(os.Stderr, "opponent evaluation\n")

					steps := currentState.bestComplete(1, currentState.Him().heldSamples)

					fmt.Fprintf(os.Stderr, "opponent steps:\n%v", steps)

					for _, step := range steps {
						future.available = Subtract(future.available, step.needed)
						future.available = Add(future.available, future.CostInThisOrder(1, step.completed))
					}

					fmt.Fprintf(os.Stderr, "future available=%v\n", future.available)
				}
				availableSamples := currentState.AvailableAtLabo(0)

				fmt.Fprintf(os.Stderr, "sample set:\n%v", availableSamples)

				steps := future.bestComplete(0, availableSamples)

				fmt.Fprintf(os.Stderr, "my steps:\n%v", steps)

				for _, sample := range me.heldSamples {
					if !sample.isInSteps(steps) {
						/* dump samples */
						Connect(sample.id)
						actionDone = true
						break
					}
				}
				if !actionDone && len(me.heldSamples) < maxHeldSamples {
					for _, step := range steps {
						for _, samp := range step.completed {
							if samp.carriedBy != 0 {
								Connect(samp.id)
								actionDone = true
								break
							}
						}
						if actionDone {
							break
						}
					}
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
			fmt.Fprintf(os.Stderr, "molecules\n")
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
					} else if me.storage.Sum() < maxHeldMolecules {
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
			if !actionDone && currentState.Him().target == laboratory && me.storage.Sum() < maxHeldMolecules {
				opponentBestSteps := currentState.bestComplete(1, currentState.Him().heldSamples)
				if len(opponentBestSteps) > 0 && opponentBestSteps[0].needed.Sum() == 0 {
					/* opponent is going to the laboratory with one or more complete samples */
					future := *currentState
					future.available = Add(future.available, currentState.CostInThisOrder(1, opponentBestSteps[0].completed))
					futureSteps := future.bestComplete(0, me.heldSamples)

					var cumulNeeded Molecules

					for i, step := range futureSteps {
						cumulNeeded = Add(cumulNeeded, step.needed)
						fmt.Fprintln(os.Stderr, currentState.available)
						fmt.Fprintln(os.Stderr, cumulNeeded)
						m, _ := currentState.moleculeToPickFirst(cumulNeeded)
						if m >= 0 {
							fmt.Fprintf(os.Stderr, "future: gather molecule for step %v\n", i)
							Gather(m)
							actionDone = true
						} else {
							fmt.Fprintf(os.Stderr, "future: CAN'T gather molecule for step %v\n", i)
						}
						break
					}
					if len(futureSteps) > 0 && !actionDone {
						if !currentState.PlayerIsWaiting(1) {
							fmt.Println("WAIT")
							actionDone = true
						}
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
			fmt.Fprintf(os.Stderr, "laboratory\n")
			actionDone := false

			steps := currentState.bestComplete(0, me.heldSamples)

			if len(steps) > 0 {
				if steps[0].needed.Sum() == 0 {
					Connect(steps[0].completed[0].id)
				} else {
					GoTo(molecules)
				}
				actionDone = true
			}

			if !actionDone {
				future := *currentState
				if currentState.Him().target == laboratory ||
					currentState.Him().target == molecules {
					fmt.Fprintf(os.Stderr, "evaluate him\n")
					steps := currentState.bestComplete(1, currentState.Him().heldSamples)
					for _, step := range steps {
						future.available = Subtract(future.available, step.needed)
						future.available = Add(future.available, future.CostInThisOrder(1, step.completed))
					}
				}

				steps := future.bestComplete(0, me.heldSamples)

				if len(steps) > 0 {
					if steps[0].needed.Sum() == 0 {
						Connect(steps[0].completed[0].id)
					} else {
						GoTo(molecules)
					}
					actionDone = true
				} else {
					availableSamples := currentState.AvailableAtLabo(0)

					steps := future.bestComplete(0, availableSamples)

					if steps.nbSamples() >= 2 && len(availableSamples) > 3 ||
						future.StepsValue(0, steps) > 5.0 {
						/* solution with at least 2 samples among 4 */
						GoTo(diagnosis)
					} else {
						GoTo(samples)
					}
				}
			}
		}
		previousState = currentState
	}
}
