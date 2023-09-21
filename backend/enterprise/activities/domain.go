package activities

import (
	"time"
)

type Slot map[string]float32
type Slots map[time.Time]Slot

func (s Slot) GetAAS() float32 {
	return s.GetTotalDBTime() / timeElapsed
}

func (s Slot) GetTotalDBTime() float32 {
	var total float32 = 0
	for _, value := range s {
		total = total + value
	}

	return total
}

func (s Slot) GetWaitEventFraction(waitEventName string) float32 {
	return s[waitEventName] / timeElapsed
}
