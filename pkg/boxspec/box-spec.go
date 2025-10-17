package boxspec

type BoxSpec struct {
	Uuid string `json:"uuid"`

	Volumes []DboxedVolume `json:"volumes,omitempty"`

	ComposeProjects map[string]string `json:"composeProjects,omitempty"`
}
