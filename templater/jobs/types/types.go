package types

type ExtraRef struct {
	BaseRef string `json:"baseRef,omitempty"`
	Org     string `json:"org,omitempty"`
	Repo    string `json:"repo,omitempty"`
}

type EnvVar struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type Resources struct {
	Requests *ResourceConfig `json:"requests,omitempty"`
	Limits   *ResourceConfig `json:"limits,omitempty"`
}

type ResourceConfig struct {
	CPU              string `json:"cpu,omitempty"`
	Memory           string `json:"memory,omitempty"`
	EphemeralStorage string `json:"ephemeral-storage,omitempty"`
}

type HostPath struct {
	Path string `json:"path,omitempty"`
	Type string `json:"type,omitempty"`
}

type Secret struct {
	Name        string `json:"name,omitempty"`
	DefaultMode int    `json:"defaultMode,omitempty"`
}

type Volume struct {
	Name       string    `json:"name,omitempty"`
	VolumeType string    `json:"volumeType,omitempty"`
	HostPath   *HostPath `json:"hostPath,omitempty"`
	Secret     *Secret   `json:"secret,omitempty"`
}

type VolumeMount struct {
	Name      string `json:"name,omitempty"`
	MountPath string `json:"mountPath,omitempty"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
}

type JobConfig struct {
	Architecture                 string         `json:"architecture,omitempty"`
	JobName                      string         `json:"jobName,omitempty"`
	RunIfChanged                 string         `json:"runIfChanged,omitempty"`
	SkipIfOnlyChanged            string         `json:"skipIfOnlyChanged,omitempty"`
	Branches                     []string       `json:"branches,omitempty"`
	MaxConcurrency               int            `json:"maxConcurrency,omitempty"`
	CronExpression               string         `json:"cronExpression,omitempty"`
	Timeout                      string         `json:"timeout,omitempty"`
	ImageBuild                   bool           `json:"imageBuild,omitempty"`
	UseDockerBuildX              bool           `json:"useDockerBuildX,omitempty"`
	UseMinimalBuilderBase        bool           `json:"useMinimalBuilderBase,omitempty"`
	PRCreation                   bool           `json:"prCreation,omitempty"`
	RuntimeImage                 string         `json:"runtimeImage,omitempty"`
	LocalRegistry                bool           `json:"localRegistry,omitempty"`
	ExtraRefs                    []*ExtraRef    `json:"extraRefs,omitempty"`
	ServiceAccountName           string         `json:"serviceAccountName,omitempty"`
	EnvVars                      []*EnvVar      `json:"envVars,omitempty"`
	Commands                     []string       `json:"commands,omitempty"`
	Resources                    *Resources     `json:"resources,omitempty"`
	VolumeMounts                 []*VolumeMount `json:"volumeMounts,omitempty"`
	Volumes                      []*Volume      `json:"volumes,omitempty"`
	AutomountServiceAccountToken string         `json:"automountServiceAccountToken,omitempty"`
	Cluster                      string         `json:"cluster,omitempty"`
	Bucket                       string         `json:"bucket,omitempty"`
	ProjectPath                  string         `json:"projectPath,omitempty"`
	RunAsUser                    string         `json:"runAsUser,omitempty"`
	RunAsGroup                   string         `json:"runAsGroup,omitempty"`
}
