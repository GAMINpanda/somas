package objects

import (
	"github.com/MattSScott/basePlatformSOMAS/v2/pkg/message"
)

// Send how much sun light energy agent puts into CTK
type ProvisionMessage struct {
	message.BaseMessage
	Amount float64
}

// Send how much veg food agent takes from CPR
type AppropriationMessage struct {
	message.BaseMessage
	Amount float64
}

func (msg *ProvisionMessage) InvokeMessageHandler(a IAmoeBoid) {
	a.HandleProvisionMessage(msg)
}

func (msg *AppropriationMessage) InvokeMessageHandler(a IAmoeBoid) {
	a.HandleAppropriationMessage(msg)
}
