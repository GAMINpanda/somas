package objects

import (
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
}

type BaseAmoeBoid struct { // add private variables like energy etc
	*agent.BaseAgent[IAmoeBoid]
	provisionMap     map[uuid.UUID]float64
	appropriationMap map[uuid.UUID]float64
}

func CreateBaseAmoeBoid(serv agent.IExposedServerFunctions[IAmoeBoid]) *BaseAmoeBoid { //constructor returns a BaseAmoeBoid registered with serv (needed since CreateBaseAgent in basePlatform needs server to be passsed) plus a map of provisions and appropriations (for messaging)
	return &BaseAmoeBoid{
		BaseAgent:        agent.CreateBaseAgent(serv),
		provisionMap:     make(map[uuid.UUID]float64),
		appropriationMap: make(map[uuid.UUID]float64),
	}
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
