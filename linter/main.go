/*
Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.

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
	"strings"

	"k8s.io/test-infra/prow/config"
	yaml "sigs.k8s.io/yaml"
)

var skipPresubmitMakeTargetCheck = []string{
	"golang.*presubmit",
	"golang.*presubmits",
	"eks-distro-base-test-presubmit",
	"eks-distro-base-tooling-presubmit",
	"lint",
}

var skipPostSubmitMakeTargetCheck = []string{
	"golang.*postsubmit",
	"golang.*postsubmits",
	"build-1-2[1-9].*postsubmit",
	"announcement",
	"release",
}

type JobConstants struct {
	Bucket                          string
	Cluster                         string
	ServiceAccountName              string
	DefaultMakeTarget               string
	HelmMakeTarget                  string
	ReleaseToolingMakeTarget        string
	PostsubmitConformanceMakeTarget string
	AttributionMakeTarget           string
	TestsMakeTarget                 string
}

type UnmarshaledJobConfig struct {
	GithubRepo    string
	FileName      string
	FileContents  string
	ProwjobConfig *config.JobConfig
}

type presubmitCheck func(presubmitConfig config.Presubmit, fileContentsString string) (passed bool, lineNo int, errorMessage string)
type postsubmitCheck func(postsubmitConfig config.Postsubmit, fileContentsString string) (passed bool, lineNo int, errorMessage string)

func (jc *JobConstants) Init(jobType string) {
	if jobType == "postsubmit" {
		jc.Bucket = "s3://prow-data-devstack-prowbucket7c73355c-h7e0qjaxyp8j"
		jc.Cluster = "prow-postsubmits-cluster"
		jc.DefaultMakeTarget = "release"
		jc.PostsubmitConformanceMakeTarget = "postsubmit-conformance"
		jc.AttributionMakeTarget = "update-attribution-files"
	} else if jobType == "presubmit" {
		jc.Bucket = "s3://prow-data-presubmits-devstack-prowbucket7c73355c-11dl00wefbvlc"
		jc.Cluster = "prow-presubmits-cluster"
		jc.ServiceAccountName = "presubmits-build-account"
		jc.DefaultMakeTarget = "build"
		jc.HelmMakeTarget = "verify"
		jc.ReleaseToolingMakeTarget = "test"
		jc.TestsMakeTarget = "test"
	}
}

func findLineNumber(fileContentsString string, searchString string) int {
	fileLines := strings.Split(fileContentsString, "\n")

	for lineNo, fileLine := range fileLines {
		if strings.Contains(fileLine, searchString) {
			return lineNo + 1
		}
	}

	return 0
}

func AlwaysRunCheck() presubmitCheck {
	return presubmitCheck(func(presubmitConfig config.Presubmit, fileContentsString string) (bool, int, string) {
		if presubmitConfig.AlwaysRun {
			return false, findLineNumber(fileContentsString, "always_run:"), "Please set always_run to false"
		}
		return true, 0, ""
	})
}

func SkipReportCheck() presubmitCheck {
	return presubmitCheck(func(presubmitConfig config.Presubmit, fileContentsString string) (bool, int, string) {
		if presubmitConfig.Reporter.SkipReport {
			return false, findLineNumber(fileContentsString, "skip_report:"), "Please set always_run to false"
		}
		return true, 0, ""
	})
}

func PresubmitBucketCheck(jc *JobConstants) presubmitCheck {
	return presubmitCheck(func(presubmitConfig config.Presubmit, fileContentsString string) (bool, int, string) {
		if strings.Contains(presubmitConfig.JobBase.Name, "builder-base-tooling-presubmit") {
			return true, 0, ""
		}
		if presubmitConfig.JobBase.UtilityConfig.DecorationConfig.GCSConfiguration.Bucket != jc.Bucket {
			return false, findLineNumber(fileContentsString, "bucket:"), fmt.Sprintf(`Incorrect bucket configuration, please configure S3 bucket as => bucket: %s`, jc.Bucket)
		}
		return true, 0, ""
	})
}

func PostsubmitBucketCheck(jc *JobConstants) postsubmitCheck {
	return postsubmitCheck(func(postsubmitConfig config.Postsubmit, fileContentsString string) (bool, int, string) {
		if postsubmitConfig.JobBase.UtilityConfig.DecorationConfig.GCSConfiguration.Bucket != jc.Bucket {
			return false, findLineNumber(fileContentsString, "bucket:"), fmt.Sprintf(`Incorrect bucket configuration, please configure S3 bucket as => bucket: %s`, jc.Bucket)
		}
		return true, 0, ""
	})
}

func PresubmitClusterCheck(jc *JobConstants) presubmitCheck {
	return presubmitCheck(func(presubmitConfig config.Presubmit, fileContentsString string) (bool, int, string) {
		if strings.Contains(presubmitConfig.JobBase.Name, "builder-base-tooling-presubmit") {
			return true, 0, ""
		}
		if presubmitConfig.JobBase.Cluster != jc.Cluster {
			return false, findLineNumber(fileContentsString, "cluster:"), fmt.Sprintf(`Incorrect cluster configuration, please configure cluster as => cluster: "%s"`, jc.Cluster)
		}
		return true, 0, ""
	})
}

func PostsubmitClusterCheck(jc *JobConstants) postsubmitCheck {
	return postsubmitCheck(func(postsubmitConfig config.Postsubmit, fileContentsString string) (bool, int, string) {
		if postsubmitConfig.JobBase.Cluster != jc.Cluster {
			return false, findLineNumber(fileContentsString, "cluster:"), fmt.Sprintf(`Incorrect cluster configuration, please configure cluster as => cluster: "%s"`, jc.Cluster)
		}
		return true, 0, ""
	})
}

func PresubmitServiceAccountCheck(jc *JobConstants) presubmitCheck {
	return presubmitCheck(func(presubmitConfig config.Presubmit, fileContentsString string) (bool, int, string) {
		if strings.Contains(presubmitConfig.JobBase.Name, "builder-base-tooling-presubmit") {
			return true, 0, ""
		}
		if presubmitConfig.JobBase.Spec.ServiceAccountName != jc.ServiceAccountName {
			return false, findLineNumber(fileContentsString, "serviceaccountName:"), fmt.Sprintf(`Incorrect service account configuration, please configure service account as => serviceaccountName: %s`, jc.ServiceAccountName)
		}
		return true, 0, ""
	})
}

func PresubmitMakeTargetCheck(jc *JobConstants) presubmitCheck {
	return presubmitCheck(func(presubmitConfig config.Presubmit, fileContentsString string) (bool, int, string) {
		if arrayStringContains(skipPresubmitMakeTargetCheck, presubmitConfig.JobBase.Name) {
			fmt.Printf("Skipping check on presubmit job %v\n", presubmitConfig.JobBase.Name)
			return true, 0, ""
		}
		jobMakeTargetMatches := regexp.MustCompile(`make (\w+[-\w]+?).*`).FindStringSubmatch(strings.Join(presubmitConfig.JobBase.Spec.Containers[0].Command, " "))
		jobMakeTarget := jobMakeTargetMatches[len(jobMakeTargetMatches)-1]
		makeCommandLineNo := findLineNumber(fileContentsString, "make")
		if strings.Contains(presubmitConfig.JobBase.Name, "helm-chart") {
			if jobMakeTarget != jc.HelmMakeTarget {
				return false, makeCommandLineNo, fmt.Sprintf(`Invalid make target, please use the "%s" target`, jc.HelmMakeTarget)
			}
		} else if strings.Contains(presubmitConfig.JobBase.Name, "release-tooling") {
			if jobMakeTarget != jc.ReleaseToolingMakeTarget {
				return false, makeCommandLineNo, fmt.Sprintf(`Invalid make target, please use the "%s" target`, jc.ReleaseToolingMakeTarget)
			}
		} else if strings.Contains(presubmitConfig.JobBase.Name, "test") {
			if jobMakeTarget != jc.TestsMakeTarget {
				return false, makeCommandLineNo, fmt.Sprintf(`Invalid make target, please use the "%s" target`, jc.TestsMakeTarget)
			}
		} else if jobMakeTarget != jc.DefaultMakeTarget {
			return false, makeCommandLineNo, fmt.Sprintf(`Invalid make target, please use the "%s" target`, jc.DefaultMakeTarget)
		}
		return true, 0, ""
	})
}

func PostsubmitMakeTargetCheck(jc *JobConstants) postsubmitCheck {
	return postsubmitCheck(func(postsubmitConfig config.Postsubmit, fileContentsString string) (bool, int, string) {
		if arrayStringContains(skipPostSubmitMakeTargetCheck, postsubmitConfig.JobBase.Name) {
			fmt.Printf("Skipping check on postsubmit job %v\n", postsubmitConfig.JobBase.Name)
			return true, 0, ""
		}
		jobMakeTargetMatches := regexp.MustCompile(`make (\w+[-\w]*)`).FindStringSubmatch(strings.Join(postsubmitConfig.JobBase.Spec.Containers[0].Command, " "))
		jobMakeTarget := jobMakeTargetMatches[len(jobMakeTargetMatches)-1]
		makeCommandLineNo := findLineNumber(fileContentsString, "make")
		if strings.HasPrefix(postsubmitConfig.JobBase.Name, "build-") {
			if jobMakeTarget != jc.PostsubmitConformanceMakeTarget {
				return false, makeCommandLineNo, fmt.Sprintf(`Invalid make target, please use the "%s" target`, jc.PostsubmitConformanceMakeTarget)
			}
		} else if strings.Contains(postsubmitConfig.JobBase.Name, "attribution") {
			if jobMakeTarget != jc.AttributionMakeTarget {
				return false, makeCommandLineNo, fmt.Sprintf(`Invalid make target, please use the "%s" target`, jc.AttributionMakeTarget)
			}
		} else if jobMakeTarget != jc.DefaultMakeTarget {
			return false, makeCommandLineNo, fmt.Sprintf(`Invalid make target, please use the "%s" target`, jc.DefaultMakeTarget)
		}
		return true, 0, ""
	})
}

func PresubmitNameDuplicationCheck(jobNamesCount map[string]int) presubmitCheck {
	return presubmitCheck(func(presubmitConfig config.Presubmit, fileContentsString string) (bool, int, string) {
		if jobNamesCount[presubmitConfig.JobBase.Name] > 1 {
			return false, findLineNumber(fileContentsString, fmt.Sprintf("name: %s", presubmitConfig.JobBase.Name)), fmt.Sprintf(`Duplicate job name => name: %s`, presubmitConfig.JobBase.Name)
		}
		return true, 0, ""
	})
}

func PostsubmitNameDuplicationCheck(jobNamesCount map[string]int) postsubmitCheck {
	return postsubmitCheck(func(postsubmitConfig config.Postsubmit, fileContentsString string) (bool, int, string) {
		if jobNamesCount[postsubmitConfig.JobBase.Name] > 1 {
			return false, findLineNumber(fileContentsString, fmt.Sprintf("name: %s", postsubmitConfig.JobBase.Name)), fmt.Sprintf(`Duplicate job name => name: %s`, postsubmitConfig.JobBase.Name)
		}
		return true, 0, ""
	})
}

func getJobNamesMap(gitRoot string) (map[string]int, error) {
	jobNamesCountMap := map[string]int{}
	allJobNames := []string{}
	yqNameEvalExpressions := map[string]string{"presubmit": ".*.*[0].name", "postsubmit": ".*.*[0].name", "periodic": ".*[0].name"}
	for jobType, expression := range yqNameEvalExpressions {
		jobNamesOutput, err := exec.Command("bash", "-c", fmt.Sprintf("find %s -type f -name '*%s*' -exec yq '%s' {} + | grep -v '\\-\\-\\-'", filepath.Join(gitRoot, "jobs/aws"), jobType, expression)).CombinedOutput()
		if err != nil {
			return nil, err
		}
		jobNames := strings.Fields(string(jobNamesOutput))
		allJobNames = append(allJobNames, jobNames...)
	}

	for _, jobName := range allJobNames {
		jobNamesCountMap[jobName]++
	}

	return jobNamesCountMap, nil
}

func getFilesChanged(gitRoot string, pullBaseSha string, pullPullSha string) ([]string, []string, error) {
	presubmitFiles := []string{}
	postsubmitFiles := []string{}
	gitDiffCommand := []string{"git", "-C", gitRoot, "diff", "--name-only", pullBaseSha, pullPullSha}
	fmt.Println("\n", strings.Join(gitDiffCommand, " "))

	gitDiffOutput, err := exec.Command("git", gitDiffCommand[1:]...).Output()

	filesChanged := strings.Fields(string(gitDiffOutput))
	for _, file := range filesChanged {
		if strings.Contains(file, "presubmits") && strings.HasPrefix(file, "jobs") && strings.HasSuffix(file, "yaml") {
			presubmitFiles = append(presubmitFiles, file)
		}
		if strings.Contains(file, "postsubmits") && strings.HasPrefix(file, "jobs") && strings.HasSuffix(file, "yaml") {
			postsubmitFiles = append(postsubmitFiles, file)
		}
	}
	return presubmitFiles, postsubmitFiles, err
}

func unmarshalJobFile(gitRoot, filePath string, jobConfig *config.JobConfig) (*UnmarshaledJobConfig, error, error) {
	absoluteFilePath := filepath.Join(gitRoot, filePath)
	if _, err := os.Stat(absoluteFilePath); os.IsNotExist(err) {
		return nil, nil, nil
	}
	unmarshaledJobConfig := new(UnmarshaledJobConfig)
	unmarshaledJobConfig.GithubRepo = strings.Replace(filepath.Dir(filePath), "jobs/", "", 1)
	unmarshaledJobConfig.FileName = filepath.Base(filePath)

	fileContents, fileReadError := ioutil.ReadFile(absoluteFilePath)
	unmarshaledJobConfig.FileContents = string(fileContents)

	fileUnmarshalError := yaml.Unmarshal(fileContents, &jobConfig)
	unmarshaledJobConfig.ProwjobConfig = jobConfig

	return unmarshaledJobConfig, fileReadError, fileUnmarshalError
}

func displayConfigErrors(fileErrorMap map[string][]string) bool {
	errorsExist := false
	for file, configErrors := range fileErrorMap {
		if len(configErrors) > 0 {
			errorsExist = true
			fmt.Printf("\n%s:\n", file)
			for _, errorString := range configErrors {
				fmt.Println(errorString)
			}
		}
	}
	fmt.Println()
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

	gitRootOutput, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		log.Fatalf("There was an error running the git command: %v", err)
	}
	gitRoot := strings.Fields(string(gitRootOutput))[0]

	jobNamesCount, err := getJobNamesMap(gitRoot)
	if err != nil {
		log.Fatalf("There was an error getting the job names count: %v", err)
	}

	presubmitFiles, postsubmitFiles, err := getFilesChanged(gitRoot, pullBaseSha, pullPullSha)
	if err != nil {
		log.Fatalf("There was an error running the git command: %v", err)
	}

	presubmitCheckFunctions := []presubmitCheck{
		AlwaysRunCheck(),
		PresubmitClusterCheck(presubmitConstants),
		SkipReportCheck(),
		PresubmitBucketCheck(presubmitConstants),
		PresubmitServiceAccountCheck(presubmitConstants),
		PresubmitMakeTargetCheck(presubmitConstants),
		PresubmitNameDuplicationCheck(jobNamesCount),
	}

	postsubmitCheckFunctions := []postsubmitCheck{
		PostsubmitClusterCheck(postsubmitConstants),
		PostsubmitBucketCheck(postsubmitConstants),
		PostsubmitMakeTargetCheck(postsubmitConstants),
		PostsubmitNameDuplicationCheck(jobNamesCount),
	}

	for _, presubmitFile := range presubmitFiles {
		unmarshaledJobConfig, fileReadError, fileUnmarshalError := unmarshalJobFile(gitRoot, presubmitFile, &jobConfig)

		// Skip linting if file is not found
		if unmarshaledJobConfig == nil {
			continue
		}
		if fileReadError != nil {
			log.Fatalf("Error reading contents of %s: %v", presubmitFile, fileReadError)
		}
		if fileUnmarshalError != nil {
			log.Fatalf("Error unmarshaling contents of %s: %v", presubmitFile, fileUnmarshalError)
		}

		presubmitConfigs, ok := unmarshaledJobConfig.ProwjobConfig.PresubmitsStatic[unmarshaledJobConfig.GithubRepo]
		if !ok {
			log.Fatalf("Key %s does not exist in Presubmit configuration map", unmarshaledJobConfig.GithubRepo)
		}
		if len(presubmitConfigs) < 1 {
			log.Fatalf("Presubmit configuration for the %s repo is empty", unmarshaledJobConfig.GithubRepo)
		}
		presubmitConfig := presubmitConfigs[0]
		for _, check := range presubmitCheckFunctions {
			passed, lineNum, errMessage := check(presubmitConfig, unmarshaledJobConfig.FileContents)
			if !passed {
				errorString := fmt.Sprintf("%d\t%s", lineNum, errMessage)
				presubmitErrors[unmarshaledJobConfig.FileName] = append(presubmitErrors[unmarshaledJobConfig.FileName], errorString)
			}
		}
	}

	for _, postsubmitFile := range postsubmitFiles {
		unmarshaledJobConfig, fileReadError, fileUnmarshalError := unmarshalJobFile(gitRoot, postsubmitFile, &jobConfig)
		// Skip linting if file is not found
		if unmarshaledJobConfig == nil {
			continue
		}
		if fileReadError != nil {
			log.Fatalf("Error reading contents of %s: %v", postsubmitFile, fileReadError)
		}
		if fileUnmarshalError != nil {
			log.Fatalf("Error unmarshaling contents of %s: %v", postsubmitFile, fileUnmarshalError)
		}

		postsubmitConfigs, ok := unmarshaledJobConfig.ProwjobConfig.PostsubmitsStatic[unmarshaledJobConfig.GithubRepo]
		if !ok {
			log.Fatalf("Key %s does not exist in Postsubmit configuration map", unmarshaledJobConfig.GithubRepo)
		}
		if len(postsubmitConfigs) < 1 {
			log.Fatalf("Postsubmit configuration for the %s repo is empty", unmarshaledJobConfig.GithubRepo)
		}
		postsubmitConfig := postsubmitConfigs[0]
		for _, check := range postsubmitCheckFunctions {
			passed, lineNum, errMessage := check(postsubmitConfig, unmarshaledJobConfig.FileContents)
			if !passed {
				errorString := fmt.Sprintf("%d\t%s", lineNum, errMessage)
				postsubmitErrors[unmarshaledJobConfig.FileName] = append(postsubmitErrors[unmarshaledJobConfig.FileName], errorString)
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

func arrayStringContains(inputArray []string, str string) bool {
	for _, v := range inputArray {
		if regexp.MustCompile(v).MatchString(str) {
			return true
		}
	}

	return false
}
