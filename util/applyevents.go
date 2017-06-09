package util

func ApplyEvents(events []interface{}) {
	for _, e := range events {
		if !ApplyOneEvent(e) {
			panic("Unknown event")
		}
	}
}
