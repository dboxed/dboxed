package boxspec

type BoxSpec struct {
	ID           string `json:"id"`
	DesiredState string `json:"desiredState"`

	Volumes []DboxedVolume `json:"volumes,omitempty"`

	ComposeProjects map[string]string `json:"composeProjects,omitempty"`
}
