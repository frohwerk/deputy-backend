package api

type Platform struct {
	Id        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	ServerUri string `json:"server,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Secret    string `json:"secret,omitempty"`
}
