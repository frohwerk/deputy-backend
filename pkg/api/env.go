package api

type Env struct {
	Id    *string `json:"id,omitempty"`
	Name  *string `json:"name,omitempty"`
	Order *int    `json:"order,omitempty"`
}

type EnvAttributes struct {
	Name string `json:"name,omitempty"`
}
