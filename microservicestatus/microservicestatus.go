package microservicestatus

const (
	StatusOK   = 0
	StatusSick = 1
	StatusDead = 2
)

type Status struct {
	Version    string `json:"version"`
	Status     int    `json:"statuscode"`
	StatusInfo string `json:"statusinfo"`
}
