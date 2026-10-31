package objects

import "github.com/MattSScott/basePlatformSOMAS/v2/pkg/agent" //Library packages IAgent,BaseAgent and CreateBaseAgent

// IAmoeBoid is the interface the server uses (Interface AmeoBoid)
type IAmoeBoid interface {
	agent.IAgent[IAmoeBoid]
}

type BaseAmoeBoid struct { // add private variables like energy etc
	*agent.BaseAgent[IAmoeBoid]
}

func NewBaseAmoeBoid(serv agent.IExposedServerFunctions[IAmoeBoid]) *BaseAmoeBoid { //constructor returns a BaseAmoeBoid registered with serv (needed since CreateBaseAgent in basePlatform needs server to be passsed)
	return &BaseAmoeBoid{BaseAgent: agent.CreateBaseAgent(serv)}
}

// We add any ameboid specific functions here e.g getting energy or adding energy func (a *BaseAmoeBoid) GetEnergy() float64 { return a.energy }
