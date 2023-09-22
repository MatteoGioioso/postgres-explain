package activities

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type WaitEvent struct {
	Type        string `json:"type"`
	Class       string `json:"class"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

func LoadWaitEventsMapFromFile(dir string) map[string]WaitEvent {
	file, err := os.ReadFile(filepath.Join(dir, "wait_events.json"))
	if err != nil {
		log.Fatalln(err)
	}

	var pgWaitEvents map[string]WaitEvent
	if err := json.Unmarshal(file, &pgWaitEvents); err != nil {
		log.Fatalln(err)
	}

	return pgWaitEvents
}
