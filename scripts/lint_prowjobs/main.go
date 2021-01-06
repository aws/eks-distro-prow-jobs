/*
Copyright 2020 Amazon.com Inc. or its affiliates. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"

	"k8s.io/test-infra/prow/config"
	yaml "sigs.k8s.io/yaml"
)

type JobConstants struct {
	Bucket                   string
	Cluster                  string
	ServiceAccountName       string
	DefaultMakeTarget        string
	CniMakeTarget            string
	HelmMakeTarget           string
	ReleaseToolingMakeTarget string
}

type presubmitCheck func(presubmitConfig config.Presubmit, fileContentsString string) (bool, string, string)
type postsubmitCheck func(postsubmitConfig config.Postsubmit, fileContentsString string) (bool, string, string)

func (jc *JobConstants) Init(jobType string) {
	if jobType == "postsubmit" {
		jc.Bucket = "s3://prowdataclusterstack-316434458-prowbucket7c73355c-1n9f9v93wpjcm"
		jc.Cluster = "prow-postsubmits-cluster"
		jc.DefaultMakeTarget = "release"
	} else if jobType == "presubmit" {
		jc.Bucket = "s3://prowpresubmitsdataclusterstack-prowbucket7c73355c-vfwwxd2eb4gp"
		jc.Cluster = "prow-presubmits-cluster"
		jc.ServiceAccountName = "presubmits-build-account"
		jc.DefaultMakeTarget = "build"
		jc.CniMakeTarget = "release"
		jc.HelmMakeTarget = "verify"
		jc.ReleaseToolingMakeTarget = "test"
	}
}

func findLineNumber(fileContentsString string, searchString string) string {
	fileLines := strings.Split(fileContentsString, "\n")

	for lineNo, fileLine := range fileLines {
		if strings.Contains(fileLine, searchString) {
			return strconv.Itoa(lineNo + 1)
		}
	}

	return ""
}

func AlwaysRunCheck() presubmitCheck {
	return presubmitCheck(func(presubmitConfig config.Presubmit, fileContentsString string) (bool, string, string) {
		if presubmitConfig.AlwaysRun {
			return false, findLineNumber(fileContentsString, "always_run:"), "Please set always_run to false"
		}
		return true, "", ""
	})
}

func SkipReportCheck() presubmitCheck {
	return presubmitCheck(func(presubmitConfig config.Presubmit, fileContentsString string) (bool, string, string) {
		if presubmitConfig.Reporter.SkipReport {
			return false, findLineNumber(fileContentsString, "skip_report:"), "Please set always_run to false"
		}
		return true, "", ""
	})
}

func PresubmitBucketCheck(jc *JobConstants) presubmitCheck {
	return presubmitCheck(func(presubmitConfig config.Presubmit, fileContentsString string) (bool, string, string) {
		if presubmitConfig.JobBase.UtilityConfig.DecorationConfig.GCSConfiguration.Bucket != jc.Bucket {
			return false, findLineNumber(fileContentsString, "bucket:"), fmt.Sprintf("Incorrect bucket configuration, please configure S3 bucket as => bucket: %s", jc.Bucket)
		}
		return true, "", ""
	})
}

func PostsubmitBucketCheck(jc *JobConstants) postsubmitCheck {
	return postsubmitCheck(func(postsubmitConfig config.Postsubmit, fileContentsString string) (bool, string, string) {
		if postsubmitConfig.JobBase.UtilityConfig.DecorationConfig.GCSConfiguration.Bucket != jc.Bucket {
			return false, findLineNumber(fileContentsString, "bucket:"), fmt.Sprintf("Incorrect bucket configuration, please configure S3 bucket as => bucket: %s", jc.Bucket)
		}
		return true, "", ""
	})
}

func PresubmitClusterCheck(jc *JobConstants) presubmitCheck {
	return presubmitCheck(func(presubmitConfig config.Presubmit, fileContentsString string) (bool, string, string) {
		if presubmitConfig.JobBase.Cluster != jc.Cluster {
			return false, findLineNumber(fileContentsString, "cluster:"), fmt.Sprintf("Incorrect cluster configuration, please configure cluster as => cluster: %s", jc.Cluster)
		}
		return true, "", ""
	})
}

func PostsubmitClusterCheck(jc *JobConstants) postsubmitCheck {
	return postsubmitCheck(func(postsubmitConfig config.Postsubmit, fileContentsString string) (bool, string, string) {
		if postsubmitConfig.JobBase.Cluster != jc.Cluster {
			return false, findLineNumber(fileContentsString, "cluster:"), fmt.Sprintf("Incorrect cluster configuration, please configure cluster as => cluster: %s", jc.Cluster)
		}
		return true, "", ""
	})
}

func ServiceAccountCheck(jc *JobConstants) presubmitCheck {
	return presubmitCheck(func(presubmitConfig config.Presubmit, fileContentsString string) (bool, string, string) {
		if presubmitConfig.JobBase.Spec.ServiceAccountName != jc.ServiceAccountName {
			return false, findLineNumber(fileContentsString, "skip_report:"), fmt.Sprintf("Incorrect service account configuration, please configure service account as => serviceaccountName: %s", jc.ServiceAccountName)
		}
		return true, "", ""
	})
}

func PresubmitMakeTargetCheck(jc *JobConstants) presubmitCheck {
	return presubmitCheck(func(presubmitConfig config.Presubmit, fileContentsString string) (bool, string, string) {
		jobMakeTargetMatches := regexp.MustCompile(`make (\w+) .*`).FindStringSubmatch(strings.Join(presubmitConfig.JobBase.Spec.Containers[0].Command, " "))
		jobMakeTarget := jobMakeTargetMatches[len(jobMakeTargetMatches)-1]
		makeCommandLineNo := findLineNumber(fileContentsString, "make")
		if strings.Contains(presubmitConfig.JobBase.Name, "cni") && jobMakeTarget != jc.CniMakeTarget {
			return false, makeCommandLineNo, fmt.Sprintf("Invalid make target, please use the \"%s\" target", jc.CniMakeTarget)
		} else if strings.Contains(presubmitConfig.JobBase.Name, "helm-chart") && jobMakeTarget != jc.HelmMakeTarget {
			return false, makeCommandLineNo, fmt.Sprintf("Invalid make target, please use the \"%s\" target", jc.HelmMakeTarget)
		} else if strings.Contains(presubmitConfig.JobBase.Name, "release-tooling") && jobMakeTarget != jc.ReleaseToolingMakeTarget {
			return false, makeCommandLineNo, fmt.Sprintf("Invalid make target, please use the \"%s\" target", jc.ReleaseToolingMakeTarget)
		} else if jobMakeTarget != jc.DefaultMakeTarget {
			return false, makeCommandLineNo, fmt.Sprintf("Invalid make target, please use the \"%s\" target", jc.DefaultMakeTarget)
		}
		return true, "", ""
	})
}

func PostsubmitMakeTargetCheck(jc *JobConstants) postsubmitCheck {
	return postsubmitCheck(func(postsubmitConfig config.Postsubmit, fileContentsString string) (bool, string, string) {
		jobMakeTargetMatches := regexp.MustCompile(`make (\w+) .*`).FindStringSubmatch(strings.Join(postsubmitConfig.JobBase.Spec.Containers[0].Command, " "))
		jobMakeTarget := jobMakeTargetMatches[len(jobMakeTargetMatches)-1]
		makeCommandLineNo := findLineNumber(fileContentsString, "make")
		if jobMakeTarget != jc.DefaultMakeTarget {
			return false, makeCommandLineNo, fmt.Sprintf("Invalid make target, please use the \"%s\" target", jc.DefaultMakeTarget)
		}
		return true, "", ""
	})
}

func getFilesChanged(gitRoot string, pullBaseSha string, pullPullSha string) ([]string, []string, error) {
	presubmitFiles := []string{}
	postsubmitFiles := []string{}
	gitDiffCommand := "git -C " + gitRoot + " diff --name-only " + pullBaseSha + " " + pullPullSha
	fmt.Println(gitDiffCommand)

	gitDiffOutput, err := exec.Command("git", strings.Split(gitDiffCommand, " ")[1:]...).Output()
	filesChanged := strings.Fields(string(gitDiffOutput))
	for _, file := range filesChanged {
		if strings.Contains(file, "presubmits") {
			presubmitFiles = append(presubmitFiles, file)
		}
		if strings.Contains(file, "postsubmits") {
			postsubmitFiles = append(postsubmitFiles, file)
		}
	}
	return presubmitFiles, postsubmitFiles, err
}

func unmarshalJobFile(filePath string, jobConfig *config.JobConfig) (string, string, string, *config.JobConfig) {
	jobOrgRepo := strings.Replace(filepath.Dir(filePath), "jobs/", "", 1)
	jobFileName := filepath.Base(filePath)
	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading contents of %s: %v", filePath, err)
	}
	fileContentsString := string(fileContents)
	if err := yaml.Unmarshal(fileContents, &jobConfig); err != nil {
		log.Fatalf("Error unmarshaling contents of %s: %v", filePath, err)
	}
	return jobOrgRepo, jobFileName, fileContentsString, jobConfig
}

func displayConfigErrors(fileErrorMap map[string][]string) bool {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	errorsExist := false
	for file, errors := range fileErrorMap {
		if len(errors) > 0 {
			errorsExist = true
		}
		fmt.Println()
		fmt.Println(file + ":")
		for _, err := range errors {
			fmt.Fprintln(w, err)
		}
	}
	w.Flush()
	return errorsExist
}

func main() {
	var jobConfig config.JobConfig
	var presubmitErrors = make(map[string][]string)
	var postsubmitErrors = make(map[string][]string)

	presubmitConstants := new(JobConstants)
	presubmitConstants.Init("presubmit")
	postsubmitConstants := new(JobConstants)
	postsubmitConstants.Init("postsubmit")

	pullBaseSha := os.Getenv("PULL_BASE_SHA")
	pullPullSha := os.Getenv("PULL_PULL_SHA")

	gitRootOutput, _ := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	gitRoot := strings.Fields(string(gitRootOutput))[0]

	presubmitFiles, postsubmitFiles, err := getFilesChanged(gitRoot, pullBaseSha, pullPullSha)
	if err != nil {
		log.Fatalf("There was an error running the git command: %v", err)
	}

	presubmitCheckFunctions := []presubmitCheck{
		AlwaysRunCheck(),
		PresubmitClusterCheck(presubmitConstants),
		SkipReportCheck(),
		PresubmitBucketCheck(presubmitConstants),
		ServiceAccountCheck(presubmitConstants),
		PresubmitMakeTargetCheck(presubmitConstants),
	}

	postsubmitCheckFunctions := []postsubmitCheck{
		PostsubmitClusterCheck(postsubmitConstants),
		PostsubmitBucketCheck(postsubmitConstants),
		PostsubmitMakeTargetCheck(postsubmitConstants),
	}

	for _, presubmitFile := range presubmitFiles {
		presubmitOrgRepo, presubmitFileName, fileContentsString, jobConfigMap := unmarshalJobFile(presubmitFile, &jobConfig)
		presubmitJobs := jobConfigMap.PresubmitsStatic
		presubmitConfig := presubmitJobs[presubmitOrgRepo][0]
		for _, check := range presubmitCheckFunctions {
			passed, lineNum, errMessage := check(presubmitConfig, fileContentsString)
			if !passed {
				errorString := lineNum + "\t" + errMessage
				presubmitErrors[presubmitFileName] = append(presubmitErrors[presubmitFileName], errorString)
			}
		}
	}

	for _, postsubmitFile := range postsubmitFiles {
		postsubmitOrgRepo, postsubmitFileName, fileContentsString, jobConfigMap := unmarshalJobFile(postsubmitFile, &jobConfig)
		postsubmitJobs := jobConfigMap.PostsubmitsStatic
		postsubmitConfig := postsubmitJobs[postsubmitOrgRepo][0]
		for _, check := range postsubmitCheckFunctions {
			passed, lineNum, errMessage := check(postsubmitConfig, fileContentsString)
			if !passed {
				errorString := lineNum + "\t" + errMessage
				postsubmitErrors[postsubmitFileName] = append(postsubmitErrors[postsubmitFileName], errorString)
			}
		}
	}

	presubmitErrorsExist := displayConfigErrors(presubmitErrors)
	postsubmitErrorsExist := displayConfigErrors(postsubmitErrors)

	if presubmitErrorsExist || postsubmitErrorsExist {
		fmt.Println("❌ Validations failed!")
		os.Exit(1)
	}
	fmt.Println("✅ Validations passed!")
}
