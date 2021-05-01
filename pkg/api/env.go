package api

type Env struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type EnvAttributes struct {
	Name string `json:"name,omitempty"`
}
