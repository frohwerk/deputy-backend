package api

type App struct {
	Id        string      `json:"id"`
	Name      string      `json:"name"`
	Artifacts []Component `json:"artifacts,omitempty"`
}
