# Prowjobs

The Prowjobs in this repo are located under the [jobs folder](../jobs) and are separated into folders based on the GitHub organization and repository they correspond to. The business logic in the [templater folder](../templater) is used to generate these Prowjobs using re-usable, extensible templates based on job type. The templater supports all types of Prowjobs - periodic, postsubmit and presubmit. This was designed for defining Prowjobs declaratively by extracting common Prowjob configuration and for ease of understanding and maintenance. This document elaborates on these settings and provides guidance on how to make changes to Prowjobs.

## Job parameters

Each Prowjob is represented as a well-defined set of parameters, which are fed as input to a YAML template to generate the final Prowjob YAML. The parameter set is unmarshalled into the following Go struct that corresponds to the standard Prowjob definition fields :
```go
type JobConfig struct {
	JobName            string         `json:"jobName,omitempty"`
	RunIfChanged       string         `json:"runIfChanged,omitempty"`
	SkipIfOnlyChanged  string         `json:"skipIfOnlyChanged,omitempty"`
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
	ProjectPath        string         `json:"projectPath,omitempty"`
}
```
This struct is extensible to support more fields and fine-tuned configuration.

The configurable parameters are explained in the table below:

|      Parameter       |                                                                                    Description                                                                                     | Periodic | Postsubmit | Presubmit |
|:--------------------:|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:| :---: | :---: | :---: |
|      `jobName`       |                                                                                Name of the Prowjob                                                                                 | ✓ | ✓ | ✓ |
|    `runIfChanged`    |                                                     Regex for the subset of files that will trigger this Prowjob when changed                                                      |  | ✓ | ✓ |
| `skipIfOnlyChanged`  |                                                                   Regex for skipping the job if all paths match                                                                    |  |  | ✓ |
|      `branches`      |                                                                    List of branches to run this Prowjob against                                                                    |   | ✓ | ✓ |
|   `maxConcurrency`   |                                                            Maximum instances of this Prowjob that can run concurrently                                                             | ✓ | ✓ | ✓ |
|   `cronExpression`   |                                                                  Cron representation of this Prowjob trigger time                                                                  | ✓ |   |   |
|      `timeout`       |                                                         The time after which this Prowjob will be aborted if not complete                                                          | ✓ | ✓ | ✓ |
|     `imageBuild`     |                                                                  Denotes if this Prowjob builds container images                                                                   | ✓ | ✓ | ✓ |
|     `prCreation`     |                                                                 Denotes if this Prowjob runs PR creation workflows                                                                 | ✓ | ✓ | ✓ |
|    `runtimeImage`    |                                                                  Image used as build environment for this Prowjob                                                                  | ✓ | ✓ | ✓ |
|   `localRegistry`    |                                                                Denotes if this job needs access to a local registry                                                                | ✓ | ✓ | ✓ |
|     `extraRefs`      |                                                    List of extra repositories that need to be cloned when running this Prowjob                                                     | ✓ | ✓ | ✓ |
| `serviceAccountName` |                                                                   The service account to use to run this Prowjob                                                                   | ✓ | ✓ | ✓ |
|      `envVars`       |                                                       List of environment variables to set in Prowjob's main build container                                                       | ✓ | ✓ | ✓ |
|      `commands`      |                                                         List of commands to be run in this Prowjob's main build container                                                          | ✓ | ✓ | ✓ |
|     `resources`      |                                                        Resource requests and limits for this Prowjob's main build container                                                        | ✓ | ✓ | ✓ |
|    `volumeMounts`    |                                                          List of pod volumes to mount into each container in this Prowjob                                                          | ✓ | ✓ | ✓ |
|      `volumes`       |                                                        List of pod volumes that can be mounted by this Prowjob's containers                                                        | ✓ | ✓ | ✓ |
|    `projectPath`     |                                      The path corresponding to this project's build scripts, against which commands are invoked after cloning                                      | ✓ | ✓ | ✓ |

Currerntly, `runIfChanged` and `skipIfOnlyChanged` cannot be used at the same time. Upstream Prow
has put this restriction in place. See the 
[Prow docs](https://github.com/kubernetes/test-infra/blob/master/prow/jobs.md#standard-triggering-and-execution-behavior-for-jobs) f
or additional information.

For example, if you want to create a presubmit:
* of name `foo-bar-job`
* stored in file `foo-bar-presubmit.yaml`
* that only runs when the `README` is changed on the `loremipsum` branch of the `rcrozean/eks-distro` repo
* that uses `foo/bar:baz` as the runtime image
* that runs image build workflows
* that runs the commands `make foo` and `docker images`,

then you will need to add the following configuration under `templater/jobs/presubmit/eks-distro/foo-bar-presubmit.yaml`:

```yaml
jobName: foo-bar-job
runIfChanged: ^README.md$
branches:
- loremipsum
runtimeImage: foo/bar:baz
imageBuild: true
commands:
- make foo
- docker images
```


## Adding or updating a Prowjob

The below steps must be followed when making changes to Prowjobs:
1. Add/update the job configuration in the appropriate file(s) under the [templater folder](../templater/jobs) based on job type.

2. Run `make verify-prowjobs -C templater` to re-generate the `jobs/` folder and verify that there are no
   discrepancies between the generated Prowjobs and the ones committed.

3. After committing changes and rebasing with `main`, run `make lint -C linter` to verify that the job(s) succeed the linting.

