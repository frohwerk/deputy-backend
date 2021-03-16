package client

type ArtifactEvent struct {
	Domain    string      `json:"domain,omitempty"`
	EventType string      `json:"event_type,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

type ArtifactInfo struct {
	Name   string `json:"name,omitempty"`
	Path   string `json:"path,omitempty"`
	Repo   string `json:"repo_key,omitempty"`
	Sha256 string `json:"sha256,omitempty"`
	Size   int    `json:"size,omitempty"`
}
