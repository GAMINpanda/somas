package server_test

import (
	"testing"
	"time"

	"github.com/GAMINpanda/somas/internal/common/objects"
	"github.com/GAMINpanda/somas/internal/server"
)

func TestInitialise(t *testing.T) {
	s := server.Initialise(10, 5, 10, 10, time.Millisecond) //10 agents, 5 iterations, 10 turns, a message bw of 10 and a 1ms timeout
	if len(s.GetAgentMap()) != 10 {
		t.Error(len(s.GetAgentMap()), "agents initialised, expected 10")
	}
}

func TestRunGame(t *testing.T) {
	s := server.Initialise(10, 5, 10, 10, time.Microsecond) //test with a us turn timeout
	s.Start()
	if s.Iterations != 5 {
		t.Error(s.Iterations, "cycles completed, expected 5")
	}
}

func TestMessaging(t *testing.T) { // Followed how the library tests sync msg broadcasts
	s := server.Initialise(10, 5, 10, 10, 0)
	sender := objects.CreateBaseAmoeBoid(s)
	s.AddAgent(sender)
	senderID := sender.GetID()

	sender.SendProvisionMessage(2.5)
	sender.SendAppropriationMessage(1.5)
	for _, agent := range s.GetAgentMap() {
		provision := agent.GetProvisionMap()[senderID]
		appropriation := agent.GetAppropriationMap()[senderID]

		if agent.GetID() == senderID && (provision != 0 || appropriation != 0) {
			t.Error(agent.GetID(), "is sender and received its own message")
		} else if agent.GetID() != senderID && (provision != 2.5 || appropriation != 1.5) {
			t.Error(agent.GetID(), "didn't receive both prov and approp messages")
		}
	}
}
