package boxspec

type BoxSpec struct {
	Uuid         string `json:"uuid"`
	DesiredState string `json:"desiredState"`

	Volumes []DboxedVolume `json:"volumes,omitempty"`

	ComposeProjects map[string]string `json:"composeProjects,omitempty"`
}
