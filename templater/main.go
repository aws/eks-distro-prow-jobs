package main

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/eks-distro-prow-jobs/templater/jobs"
	"github.com/aws/eks-distro-prow-jobs/templater/jobs/types"
	"github.com/aws/eks-distro-prow-jobs/templater/jobs/utils"
)

var jobsFolder = "jobs"
var orgsSupported = []string{"aws"}
var jobTypes = []string{"periodic", "postsubmit", "presubmit"}

//go:embed templates/presubmits.yaml
var presubmitTemplate string

//go:embed templates/postsubmits.yaml
var postsubmitTemplate string

//go:embed templates/periodics.yaml
var periodicTemplate string

//go:embed templates/warning.txt
var editWarning string

//go:generate cp ../BUILDER_BASE_TAG_FILE ./BUILDER_BASE_TAG_FILE
//go:embed BUILDER_BASE_TAG_FILE
var builderBaseTag string

var buildkitImageTag = "v0.10.5-rootless"

func main() {
	jobsFolderPath, err := getJobsFolderPath()
	if err != nil {
		fmt.Printf("Error getting jobs folder path: %v", err)
		os.Exit(1)
	}

	for _, org := range orgsSupported {
		if err = os.RemoveAll(filepath.Join(jobsFolderPath, org)); err != nil {
			fmt.Printf("Error removing jobs folder path: %v", err)
			os.Exit(1)
		}
	}

	for _, jobType := range jobTypes {
		jobList, err := jobs.GetJobList(jobType)
		if err != nil {
			fmt.Printf("Error getting job list: %v\n", err)
			os.Exit(1)
		}
		template, err := useTemplate(jobType)
		if err != nil {
			fmt.Printf("Error getting job list: %v\n", err)
			os.Exit(1)
		}

		for repoName, jobConfigs := range jobList {
			for fileName, jobConfig := range jobConfigs {
				envVars := jobConfig.EnvVars

				if jobConfig.UseDockerBuildX {
					envVars = append(envVars, &types.EnvVar{Name: "BUILDKITD_IMAGE", Value: "moby/buildkit:" + buildkitImageTag})
				}

				branches := jobConfig.Branches
				if jobType == "postsubmit" && len(branches) == 0 {
					branches = append(branches, "^main$")
				}

				data := map[string]interface{}{
					"architecture":                 jobConfig.Architecture,
					"repoName":                     repoName,
					"prowjobName":                  jobConfig.JobName,
					"runIfChanged":                 jobConfig.RunIfChanged,
					"skipIfOnlyChanged":            jobConfig.SkipIfOnlyChanged,
					"branches":                     branches,
					"cronExpression":               jobConfig.CronExpression,
					"maxConcurrency":               jobConfig.MaxConcurrency,
					"timeout":                      jobConfig.Timeout,
					"extraRefs":                    jobConfig.ExtraRefs,
					"imageBuild":                   jobConfig.ImageBuild,
					"useDockerBuildX":              jobConfig.UseDockerBuildX,
					"prCreation":                   jobConfig.PRCreation,
					"runtimeImage":                 jobConfig.RuntimeImage,
					"localRegistry":                jobConfig.LocalRegistry,
					"serviceAccountName":           jobConfig.ServiceAccountName,
					"command":                      strings.Join(jobConfig.Commands, "\n&&\n"),
					"builderBaseTag":               builderBaseTag,
					"buildkitImageTag":             buildkitImageTag,
					"resources":                    jobConfig.Resources,
					"envVars":                      envVars,
					"volumes":                      jobConfig.Volumes,
					"volumeMounts":                 jobConfig.VolumeMounts,
					"editWarning":                  editWarning,
					"automountServiceAccountToken": jobConfig.AutomountServiceAccountToken,
				}

				err := GenerateProwjob(fileName, template, data)
				if err != nil {
					fmt.Printf("Error generating Prowjob %s: %v\n", fileName, err)
					os.Exit(1)
				}
			}
		}
	}
}

func GenerateProwjob(prowjobFileName, templateContent string, data map[string]interface{}) error {
	bytes, err := utils.ExecuteTemplate(templateContent, data)
	if err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	jobsFolderPath, err := getJobsFolderPath()
	if err != nil {
		return fmt.Errorf("error getting jobs folder path: %v", err)
	}

	prowjobPath := filepath.Join(jobsFolderPath, data["repoName"].(string), prowjobFileName)
	if err = os.MkdirAll(filepath.Dir(prowjobPath), 0o755); err != nil {
		return fmt.Errorf("error creating Prowjob directory: %v", err)
	}

	if err = ioutil.WriteFile(prowjobPath, bytes, 0o644); err != nil {
		return fmt.Errorf("error writing to path %s: %v", prowjobPath, err)
	}

	return nil
}

func getJobsFolderPath() (string, error) {
	gitRootOutput, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("error running the git command: %v", err)
	}
	gitRoot := strings.Fields(string(gitRootOutput))[0]

	return filepath.Join(gitRoot, jobsFolder), nil
}

func useTemplate(jobType string) (string, error) {
	switch jobType {
	case "periodic":
		return periodicTemplate, nil
	case "postsubmit":
		return postsubmitTemplate, nil
	case "presubmit":
		return presubmitTemplate, nil
	default:
		return "", fmt.Errorf("Unsupported job type: %s", jobType)
	}
}
