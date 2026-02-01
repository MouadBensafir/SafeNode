package backend

type BackendInfo struct {
	URL                string `json:"url"`
	Alive              bool   `json:"alive"`
	CurrentConnections int64  `json:"current_connections"`
}
