package statusinfrastructure

const (
	EventStatusEndpoint = "http://localhost:10000/eventstatus"
)

type EventNodeStatus struct {
	Name string `json:"name"`
}
