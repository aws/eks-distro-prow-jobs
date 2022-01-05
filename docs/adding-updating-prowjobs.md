# Adding or updating Prowjobs

To add/update a Prowjob:
1. Add/update the job configuration under the appropriate file under the [templater folder](../templater/jobs) based on job type. The current configuration struct looks like this:
```Go
type JobConfig struct {
	JobName            string         `json:"jobName,omitempty"`
	RunIfChanged       string         `json:"runIfChanged,omitempty"`
	Branches           []string       `json:"branches,omitempty"`
	MaxConcurrency     int            `json:"maxConcurrency,omitempty"`
	CronExpression     string         `json:"cronExpression,omitempty"`
	Timeout            string         `json:"timeout,omitempty"`
	ImageBuild         bool           `json:"imageBuild,omitempty"`
	PRCreation         bool           `json:"prCreation,omitempty"`
	RuntimeImage       string         `json:"runtimeImage,omitempty"`
	LocalRegistry      bool           `json:"localRegistry,omitempty"`
	ExtraRefs          []*ExtraRef    `json:"extraRefs,omitempty"`
	ServiceAccountName string         `json:"serviceAccountName,omitempty"`
	EnvVars            []*EnvVar      `json:"envVars,omitempty"`
	Commands           []string       `json:"commands,omitempty"`
	Resources          *Resources     `json:"resources,omitempty"`
	VolumeMounts       []*VolumeMount `json:"volumeMounts,omitempty"`
	Volumes            []*Volume      `json:"volumes,omitempty"`
}
```
This struct is extensible to support more fields and fine-tuned configuration.

2. Run `make verify-prowjobs -C templater` to re-generate the `jobs/` folder and verify that there are no
   discrepancies between the generated Prowjobs and the ones committed.

3. After committing changes and rebasing with `main`, run `make lint -C lint` to verify that the added/updated job(s) succeeds the linting.

