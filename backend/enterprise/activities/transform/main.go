package main

import (
	"encoding/json"
	"log"
	"os"
	"postgres-explain/backend/enterprise/activities"
	"strings"
)

func LoadWaitEventsMapFromFile() map[string]activities.WaitEvent {
	file, err := os.ReadFile("wait_events.txt")
	if err != nil {
		log.Fatalln(err)
	}

	filestr := string(file)
	pgWaitEvents := make(map[string]activities.WaitEvent)
	for _, line := range strings.Split(filestr, "\n") {
		eventPlusDescription := strings.Split(line, "__")            // [LWLock:WALBufMapping, Waiting for...]
		classPlusType := strings.Split(eventPlusDescription[0], ":") //[LWLock,WALBufMapping]
		pgWaitEvents[classPlusType[1]] = activities.WaitEvent{
			Type:        classPlusType[1],
			Class:       classPlusType[0],
			Description: eventPlusDescription[1],
		}
	}

	return pgWaitEvents
}

func main() {
	file := LoadWaitEventsMapFromFile()
	marshal, err := json.MarshalIndent(file, "", "    ")
	if err != nil {
		log.Fatalln(err)
	}

	if err := os.WriteFile("../wait_events.json", marshal, 0700); err != nil {
		log.Fatalln(err)
	}

	if err := os.Chmod("wait_events.json", 0777); err != nil {
		log.Fatalln(err)
	}
}
