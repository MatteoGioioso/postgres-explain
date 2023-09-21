package activities

type QueryRank struct {
	ID    string
	Total float32
	Query string
}

type QuerySlot map[string]float32
type QueriesSlots map[string]QuerySlot

type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type waitEventsGroup map[string]Group

var waitEventsGroupsMap = waitEventsGroup{
	"application_name": Group{
		ID:   "application_name",
		Name: "Application",
	},
	"usename": Group{
		ID:   "usename",
		Name: "User",
	},
	"datname": Group{
		ID:   "datname",
		Name: "Database",
	},
	"instance_name": Group{
		ID:   "instance_name",
		Name: "Instance",
	},
}
