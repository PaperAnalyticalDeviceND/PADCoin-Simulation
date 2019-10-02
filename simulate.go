package main

import (
	//"log"
	"flag"
	"math/rand"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

/*
	Card Manufacturer
		- Source: Agent buys cards
		- Sink: Tax to goverment

	Drug Manufacturer
		- Source: Government
		- Sink: Pays lab for tests

	Agent
		- Source: Buy's padcoin to get cards 
		- Sink: Buying cards

	Lab
		- Source: Drug Manufacturer
		- Sink: Pays agent on successful fake detected

	Goverment
		- Source: Card Manufacturer Tax
		- Sink: Pays drug manufacturers

	Exchange
		- Source: People selling coins
		- Sink: People buying with real money
*/

const MANUFACTURER_COUNT = 10
const MANUFACTURER_COST_PER_CARD = 0.01

var manufacturers [MANUFACTURER_COUNT]Manufacturer
type Manufacturer struct {
	Wallet float64
}

func (m *Manufacturer) Step(){

}

const AGENT_COUNT = 1000
const AGENT_FAKE_FIND_RATE = 0.01
const AGENT_REWARD = 10

var agents [AGENT_COUNT]Agent
type Agent struct {
	Wallet float64
}

func (a *Agent) Step(){
	// Work if we have enough money
	if a.Wallet > MANUFACTURER_COST_PER_CARD {
		// Select random manufacturer
		selected := rand.Intn(MANUFACTURER_COUNT)

		// Purchase card
		manufacturers[selected].Wallet += MANUFACTURER_COST_PER_CARD
		a.Wallet -= MANUFACTURER_COST_PER_CARD

		// Run test
		if rand.Float64() < AGENT_FAKE_FIND_RATE {
			// Select random lab
			lab := rand.Intn(LAB_COUNT)

			// Send card
			labs[lab].ReceiveCard(a)
		}
	}
}

const LAB_COUNT = 100
const LAB_ACCURACY = 0.95
const LAB_REWARD = 50

var labs [LAB_COUNT]Lab
type Lab struct {
	Wallet float64
	Accepted int
	Rejected int
}

func (l *Lab) Step(){
	
}

func (l *Lab) ReceiveCard(a *Agent) {
	// Test if we agree
	if rand.Float64() < LAB_ACCURACY {
		l.Wallet += LAB_REWARD
		a.Wallet += AGENT_REWARD

		l.Accepted++
	}else{
		l.Rejected++
	}
}

func main(){
	iterations := flag.Int("iterations", 10, "Number of iterations to simulate")
	flag.Parse()

	// Initialize
	for i := 0; i < MANUFACTURER_COUNT; i++ {
		manufacturers[i].Wallet = 0
	}

	for i := 0; i < AGENT_COUNT; i++ {
		agents[i].Wallet = 1
	}

	for i := 0; i < LAB_COUNT; i++ {
		labs[i].Wallet = 100
	}

	// Simulate
	manAverage := make(plotter.XYs, *iterations)
	ageAverage := make(plotter.XYs, *iterations)
	labAverage := make(plotter.XYs, *iterations)

	for i := 0; i < *iterations; i++ {
		//log.Println("Iteration", i)

		for m := 0; m < MANUFACTURER_COUNT; m++ {
			manufacturers[m].Step()
		}
	
		for a := 0; a < AGENT_COUNT; a++ {
			agents[a].Step()
		}
	
		for l := 0; l < LAB_COUNT; l++ {
			labs[l].Step()
		}

		// Print Statistics
		mAverage := 0.0
		for _, m := range manufacturers {
			mAverage += m.Wallet
		}
		mAverage /= MANUFACTURER_COUNT
		manAverage[i].X = float64(i)
		manAverage[i].Y = mAverage
		//log.Println("Manufacturers", mAverage)

		aAverage := 0.0
		for _, a := range agents {
			aAverage += a.Wallet
		}
		aAverage /= AGENT_COUNT
		ageAverage[i].X = float64(i)
		ageAverage[i].Y = aAverage
		//log.Println("Agents", aAverage)

		lAverage := 0.0
		for _, l := range labs {
			lAverage += l.Wallet
		}
		lAverage /= LAB_COUNT
		labAverage[i].X = float64(i)
		labAverage[i].Y = lAverage
		//log.Println("Labs", lAverage)
		//log.Println("")
	}

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "PADCoin Simulation"
	p.X.Label.Text = "Iteration"
	p.Y.Label.Text = "Average Value"

	err = plotutil.AddLinePoints(p,
		"Manufacturer", manAverage,
		"Agent", ageAverage,
		"Lab", labAverage)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(20*vg.Inch, 20*vg.Inch, "points.png"); err != nil {
		panic(err)
	}
}