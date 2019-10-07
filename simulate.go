package main

import (
	"log"
	"flag"
	"math/rand"
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
	
	Miner
		- Source: Minting blocks
		- Sink: Selling coins
*/

var verbose bool

type Transaction struct {
	From *Actor
	To *Actor
	Type string
	Amount float64
}

var queue []Transaction

// Actor Interface
var actors []Actor
type Actor interface {
	Step(index int64)
}

// Card Manufacturer Code
const CARD_COST = 1.25
const CARD_TAX = 0.95
const CARD_MANUFACTURERS = 10

var cManufacturer [CARD_MANUFACTURERS]CardManufacturer
type CardManufacturer struct {
	Wallet float64
	TaxCount int
}

func (m *CardManufacturer) Step(index int64){
	var self Actor = m
	if m.TaxCount > 0 && m.Wallet < CARD_TAX {
		log.Println("Card Manufacturer Exhausted")
	}
	if m.Wallet > CARD_TAX && m.TaxCount > 0 {
		var other Actor = &goverment
		queue = append(queue, Transaction{From:&self, To:&other, Type:"CARDTAX", Amount:CARD_TAX})
		m.Wallet -= CARD_TAX
		m.TaxCount--
	}
	if m.Wallet > CARD_TAX * 100 {
		var other Actor = &exchange
		queue = append(queue, Transaction{From:&self, To:&other, Type:"COINSELL", Amount:CARD_TAX*10})
		m.Wallet -= CARD_TAX*10
	}
}

// Drug Manufacturer Code
const DRUG_MANUFACTURERS = 10
const TEST_COST = 2.5

var dManufacturer [CARD_MANUFACTURERS]DrugManufacturer
type DrugManufacturer struct {
	Wallet float64
	LabRequest []*Actor
}

func (m *DrugManufacturer) Step(index int64){
	var self Actor = m
	if len(m.LabRequest) > 0 && m.Wallet < TEST_COST {
		log.Println("Drug Manufacturer Exhausted")
	}
	if m.Wallet > TEST_COST && len(m.LabRequest) > 0 {
		queue = append(queue, Transaction{From:&self, To:m.LabRequest[0], Type:"LABFEE", Amount:TEST_COST})
		m.Wallet -= TEST_COST
		m.LabRequest = m.LabRequest[1:]
	}
}

// Agent Code
const AGENTS = 100000
const FAKE_DRUG_PERCENTAGE = 0.3

const FUNDING_AMOUNT = CARD_COST
const DAYS_BETWEEN_FUNDING = 30

var agents [AGENTS]Agent
type Agent struct {
	Wallet float64
	Cards int
}

var cardsBought = 0
var fakesFound = 0
func (a *Agent) Step(index int64){
	var self Actor = a
	if a.Wallet < CARD_COST && index % DAYS_BETWEEN_FUNDING == 0{
		var other Actor = &exchange
		queue = append(queue, Transaction{From:&self, To:&other, Type:"COINBUY", Amount:FUNDING_AMOUNT})
	}

	if a.Cards == 0 && a.Wallet > CARD_COST {
		selected := rand.Intn(CARD_MANUFACTURERS)
		var manufacturer Actor = &cManufacturer[selected]

		queue = append(queue, Transaction{From:&self, To:&manufacturer, Type:"CARDBUY", Amount:CARD_COST})
		cardsBought++
		a.Wallet -= CARD_COST
	}

	if a.Cards > 0 {
		dSelected := rand.Intn(DRUG_MANUFACTURERS)
		if rand.Float64() < FAKE_DRUG_PERCENTAGE {
			selected := rand.Intn(LABS)
			var lab Actor = &labs[selected]
			queue = append(queue, Transaction{From:&self, To:&lab, Type:"FAKEFOUND", Amount:float64(dSelected)})
			fakesFound++
			Mine(&self)
		}
		a.Cards--
	}
}

// Lab Code
const LABS = 10
const AGENT_REWARD = 1.5

var labs [LABS]Lab
type Lab struct {
	Wallet float64
	Tests []*Actor
}

var fakesVerified = 0 
func (l *Lab) Step(index int64){
	var self Actor = l
	if len(l.Tests) > 0 && l.Wallet < AGENT_REWARD {
		log.Println("Lab Exhausted")
	}
	if l.Wallet > AGENT_REWARD && len(l.Tests) > 0 {
		queue = append(queue, Transaction{From:&self, To:l.Tests[0], Type:"FAKECONFIRM", Amount:AGENT_REWARD})
		fakesVerified++
		l.Wallet -= AGENT_REWARD
		l.Tests = l.Tests[1:]
	}
}

// Government Code
const DRUG_PAYMENT_DAYS = 30
const DRUG_PAYMENT_RATE = 50

var goverment Government
type Government struct {
	Wallet float64
}

func (g *Government) Step(index int64){
	if index % DRUG_PAYMENT_DAYS == 0 {
		var self Actor = g
		for i := 0; i < DRUG_MANUFACTURERS; i++ {
			if g.Wallet > DRUG_PAYMENT_RATE {
				var tester Actor = &dManufacturer[i]
				queue = append(queue, Transaction{From:&self, To:&tester, Type:"DRUGPAYMENT", Amount:DRUG_PAYMENT_RATE})
				g.Wallet -= DRUG_PAYMENT_RATE
			}else{
				log.Println("Government Exhausted")
			}
		}
	}
}


// Exchange Code
var exchange Exchange
type Exchange struct {
	Wallet float64
	Buys []*Actor
}

func (e *Exchange) Step(index int64){
	var self Actor = e
	for i := 0; i < len(e.Buys); i++ {
		if e.Wallet > FUNDING_AMOUNT {
			queue = append(queue, Transaction{From:&self, To:e.Buys[i], Type:"COINSEND", Amount:FUNDING_AMOUNT})
			e.Wallet -= FUNDING_AMOUNT
			e.Buys = e.Buys[1:]
		}else{
			log.Println("Exchange Exhausted")
		}
	}
}

// Mining Code
const MINERS = 10
const MINING_REWARD = 1
const TRANSACTIONS_PER_BLOCK = 500
var minedTransactions = 0

func Mine(a *Actor) {
	for i := 0; i < TRANSACTIONS_PER_BLOCK; i++ {
		if len(queue) > 0 {
			minedTransactions += TRANSACTIONS_PER_BLOCK

			t := queue[0]
			if verbose {
				log.Printf("Processing: %+v", t)
			}

			switch t.Type {
			case "COINSELL":
				exchange := interface{}(*t.To).(*Exchange)
				exchange.Wallet += t.Amount
			case "COINBUY":
				exchange := interface{}(*t.To).(*Exchange)
				exchange.Buys = append(exchange.Buys, t.From)
			case "COINSEND":
				switch to := interface{}(*t.To).(type) {
				case *CardManufacturer:
					to.Wallet += t.Amount
				case *DrugManufacturer:
					to.Wallet += t.Amount
				case *Agent:
					to.Wallet += t.Amount
				case *Lab:
					to.Wallet += t.Amount
				case *Government:
					to.Wallet += t.Amount
				case *Miner:
					to.Wallet += t.Amount
				}
			case "DRUGPAYMENT":
				manufacturer := interface{}(*t.To).(*DrugManufacturer)
				manufacturer.Wallet += t.Amount
			case "CARDBUY":
				manufacturer := interface{}(*t.To).(*CardManufacturer)
				manufacturer.Wallet += t.Amount
				manufacturer.TaxCount++

				agent := interface{}(*t.From).(*Agent)
				agent.Cards++
			case "CARDTAX":
				manufacturer := interface{}(*t.From).(*CardManufacturer)
				manufacturer.TaxCount--

				gov := interface{}(*t.To).(*Government)
				gov.Wallet += t.Amount
			case "FAKEFOUND":
				lab := interface{}(*t.To).(*Lab)
				lab.Tests = append(lab.Tests, t.From)

				drug := &dManufacturer[int(t.Amount)]
				drug.LabRequest = append(drug.LabRequest, t.To)
			case "LABFEE":
				lab := interface{}(*t.To).(*Lab)
				lab.Wallet += t.Amount
			case "FAKECONFIRM":
				agent := interface{}(*t.To).(*Agent)
				agent.Wallet += t.Amount
			}

			switch x := interface{}(*a).(type) {
			case *Agent:
				x.Wallet += MINING_REWARD
			case *Miner:
				x.Wallet += MINING_REWARD
			}
			queue = queue[1:]
		}
	}
}

const MINER_SELL_VALUE = 8

var miners [MINERS]Miner
type Miner struct {
	Wallet float64
}

func (m *Miner) Step(index int64){
	var self Actor = m
	if m.Wallet > MINER_SELL_VALUE {
		var other Actor = &exchange
		queue = append(queue, Transaction{From:&self, To:&other, Type:"COINSELL", Amount:MINER_SELL_VALUE})
		m.Wallet -= MINER_SELL_VALUE
	}
	Mine(&self)
}

func AverageWallet(actors []Actor) (retVal float64) {
	for _, actor := range actors {
		switch a := interface{}(actor).(type) {
		case *CardManufacturer:
			retVal += a.Wallet
		case *DrugManufacturer:
			retVal += a.Wallet
		case *Agent:
			retVal += a.Wallet
		case *Lab:
			retVal += a.Wallet
		case *Government:
			retVal += a.Wallet
		case *Miner:
			retVal += a.Wallet
		}
	}
	return retVal / float64(len(actors))
}

func main(){
	flag.BoolVar(&verbose, "verbose", false, "")
	days := flag.Int("days", 10, "Number of days to simulate")
	flag.Parse()

	// Initialization
	goverment.Wallet = DRUG_MANUFACTURERS * DRUG_PAYMENT_RATE * 10
	actors = append(actors, &goverment)

	exchange.Wallet = 100000000
	actors = append(actors, &exchange)

	var cmSlice []Actor
	for i := 0; i < CARD_MANUFACTURERS; i++ {
		cmSlice = append(cmSlice, &cManufacturer[i])
		actors = append(actors, &cManufacturer[i])
	}

	var dmSlice []Actor
	for i := 0; i < DRUG_MANUFACTURERS; i++ {
		dmSlice = append(dmSlice, &dManufacturer[i])
		actors = append(actors, &dManufacturer[i])
	}

	var aSlice []Actor
	for i := 0; i < AGENTS; i++ {
		aSlice = append(aSlice, &agents[i])
		actors = append(actors, &agents[i])
	}

	var lSlice []Actor
	for i := 0; i < LABS; i++ {
		lSlice = append(lSlice, &labs[i])
		actors = append(actors, &labs[i])
	}

	var mSlice []Actor
	for i := 0; i < MINERS; i++ {
		mSlice = append(mSlice, &miners[i])
		actors = append(actors, &miners[i])
	}

	// Simulation
	for i := 0; i < *days; i++ {
		minedTransactions, cardsBought, fakesFound, fakesVerified = 0, 0, 0, 0
		for a := 0; a < len(actors); a++ {
			actors[a].Step(int64(i))
		}

		log.Println("Day:", i)
		log.Println("Transactions Pending:", len(queue), "Mined:", minedTransactions)
		log.Println("Cards Purchased:", cardsBought)
		log.Println("Fakes Found:", fakesFound, "Verified:", fakesVerified)
		log.Println("Exchange Coins:", exchange.Wallet)
		log.Println("Government Coins:", goverment.Wallet)
		log.Println("Card Manufacturer Average:", AverageWallet(cmSlice))
		log.Println("Drug Manufacturer Average:", AverageWallet(dmSlice))
		log.Println("Agent Average:", AverageWallet(aSlice))
		log.Println("Lab Average:", AverageWallet(lSlice))
		log.Println("Miner Average:", AverageWallet(mSlice))
		log.Println()
	}
	
	// Statistics
	//for 
}