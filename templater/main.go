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
	"github.com/aws/eks-distro-prow-jobs/templater/jobs/utils"
)

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

var buildkitImageTag = "v0.9.0-rootless"

func main() {
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
				data := map[string]interface{}{
					"repoName":           repoName,
					"prowjobName":        jobConfig.JobName,
					"runIfChanged":       jobConfig.RunIfChanged,
					"branches":           jobConfig.Branches,
					"cronExpression":     jobConfig.CronExpression,
					"maxConcurrency":     jobConfig.MaxConcurrency,
					"timeout":            jobConfig.Timeout,
					"extraRefs":          jobConfig.ExtraRefs,
					"imageBuild":         jobConfig.ImageBuild,
					"prCreation":         jobConfig.PRCreation,
					"runtimeImage":       jobConfig.RuntimeImage,
					"localRegistry":      jobConfig.LocalRegistry,
					"serviceAccountName": jobConfig.ServiceAccountName,
					"command":            strings.Join(jobConfig.Commands, "\n&&\n"),
					"builderBaseTag":     builderBaseTag,
					"buildkitImageTag":   buildkitImageTag,
					"resources":          jobConfig.Resources,
					"envVars":            jobConfig.EnvVars,
					"volumes":            jobConfig.Volumes,
					"volumeMounts":       jobConfig.VolumeMounts,
					"editWarning":        editWarning,
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
	jobsFolder := "jobs"
	bytes, err := utils.ExecuteTemplate(templateContent, data)
	if err != nil {
		return err
	}

	gitRootDir, err := getGitRootDir()
	if err != nil {
		return err
	}
	prowjobPath := filepath.Join(gitRootDir, jobsFolder, data["repoName"].(string), prowjobFileName)
	err = ioutil.WriteFile(prowjobPath, bytes, 0o644)
	if err != nil {
		return fmt.Errorf("error writing to path %s: %v", prowjobPath, err)
	}

	return nil
}

func getGitRootDir() (string, error) {
	gitRootOutput, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("error running the git command: %v", err)
	}
	gitRoot := strings.Fields(string(gitRootOutput))[0]

	return gitRoot, nil
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
