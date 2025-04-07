package runn_client

type Process struct {
	ID           uint64 `json:"id"`
	Cmd          string `json:"cmd"`
	Dir          string `json:"dir"`
	Status       string `json:"status"`
	Disabled     bool   `json:"disabled"`
	RestartCount int    `json:"restart_count"`
	LastExitCode int    `json:"last_exit_code"`
}

type OutgoingEvent struct {
	Type    string   `json:"type"`
	Process *Process `json:"process"`
}
