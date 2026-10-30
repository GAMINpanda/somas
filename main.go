package main

import (
	"fmt"
	"time"

	"github.com/MattSScott/basePlatformSOMAS/v2/pkg/agent"
	"github.com/MattSScott/basePlatformSOMAS/v2/pkg/server"
)

// Agent embeds the platform's base agent so coursework behaviour can be added here.
type Agent struct {
	*agent.BaseAgent[*Agent]
}

func main() {
	serv := server.CreateBaseServer[*Agent](1, 1, time.Second, 10)
	ag := &Agent{BaseAgent: agent.CreateBaseAgent[*Agent](serv)}
	serv.AddAgent(ag)
	fmt.Printf("SOMAS platform ready: %d agent registered\n", len(serv.GetAgentMap()))
}
