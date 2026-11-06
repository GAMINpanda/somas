package objects

import (
	"math/rand/v2"
	"github.com/MattSScott/basePlatformSOMAS/v2/pkg/agent" //Library packages IAgent,BaseAgent and CreateBaseAgent
	"github.com/google/uuid"
)

// IAmoeBoid is the interface the server uses (Interface AmeoBoid)
type IAmoeBoid interface {
	agent.IAgent[IAmoeBoid]
	HandleProvisionMessage(msg *ProvisionMessage)
	SendProvisionMessage(amount float64)
	HandleAppropriationMessage(msg *AppropriationMessage)
	SendAppropriationMessage(amount float64)
	GetProvisionMap() map[uuid.UUID]float64
	GetAppropriationMap() map[uuid.UUID]float64
	Eat(amount float64)
	AddBrain(amount float64)
	GetUtility() float64
	GetBody() float64
	GetBrain() float64
	SpendBrain(amount float64) float64
	Metabolise(cost float64) bool
	DecideAppropriation(cpr float64) float64
	DecideProvision() float64
}

type BaseAmoeBoid struct { // add private variables like energy etc
	*agent.BaseAgent[IAmoeBoid]
	provisionMap     map[uuid.UUID]float64	// The latest amount each other agents says it gave to CTK
	appropriationMap map[uuid.UUID]float64	// The latest amount each other agents says it took from CPR
	body 			 float64 				// Energy from vegetables
	brain 			 float64                // Energy from sunlight 
}

func CreateBaseAmoeBoid(serv agent.IExposedServerFunctions[IAmoeBoid]) *BaseAmoeBoid { //constructor returns a BaseAmoeBoid registered with serv (needed since CreateBaseAgent in basePlatform needs server to be passsed) plus a map of provisions and appropriations (for messaging)
	return &BaseAmoeBoid{
		BaseAgent:        agent.CreateBaseAgent(serv),
		provisionMap:     make(map[uuid.UUID]float64),
		appropriationMap: make(map[uuid.UUID]float64),
	}
}

// Agent state/energy functions
func (ba *BaseAmoeBoid) Eat(amount float64)      { 
	ba.body += amount 
}

func (ba *BaseAmoeBoid) AddBrain(amount float64) { 
	ba.brain += amount 
}

func (ba *BaseAmoeBoid) GetUtility() float64     { 
	return min(ba.body, ba.brain) 
}

func (ba *BaseAmoeBoid) GetBody() float64 {
	return ba.body
}

func (ba *BaseAmoeBoid) GetBrain() float64 {
	return ba.brain
}

func (ba *BaseAmoeBoid) SpendBrain(amount float64) float64 {
	spent := max(0, min(amount, ba.brain))
	ba.brain -= spent
	return spent
}

func (ba *BaseAmoeBoid) Metabolise(cost float64) bool {
	ba.body -= cost
	ba.brain -= cost
	return ba.body > 0 && ba.brain > 0
}

// Game strategy: random for now
func (ba *BaseAmoeBoid) DecideAppropriation(cpr float64) float64 { 
	return rand.Float64() * 50 
}

func (ba *BaseAmoeBoid) DecideProvision() float64 { 
	return rand.Float64() * ba.brain 
}


// Functions for sending and recieving messages, send an amount and store recv amounts in map
func (ba *BaseAmoeBoid) HandleProvisionMessage(msg *ProvisionMessage) {
	ba.provisionMap[msg.GetSender()] = msg.Amount
}

func (ba *BaseAmoeBoid) HandleAppropriationMessage(msg *AppropriationMessage) {
	ba.appropriationMap[msg.GetSender()] = msg.Amount
}

func (ba *BaseAmoeBoid) SendProvisionMessage(amount float64) {
	msg := &ProvisionMessage{BaseMessage: ba.CreateBaseMessage(), Amount: amount}
	ba.BroadcastSynchronousMessage(msg)
}

func (ba *BaseAmoeBoid) SendAppropriationMessage(amount float64) {
	msg := &AppropriationMessage{BaseMessage: ba.CreateBaseMessage(), Amount: amount}
	ba.BroadcastSynchronousMessage(msg)
}

// Return the latest provisions each agent has sent
func (ba *BaseAmoeBoid) GetProvisionMap() map[uuid.UUID]float64 {
	return ba.provisionMap
}

// Return the latest appropriations each agent has sent
func (ba *BaseAmoeBoid) GetAppropriationMap() map[uuid.UUID]float64 {
	return ba.appropriationMap
}
