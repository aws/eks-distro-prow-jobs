package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"

	"github.com/aws/eks-distro-prow-jobs/templater/jobs/types"
)

var releaseBranches = []string{
	"1-20",
	"1-21",
	"1-22",
	"1-23",
}

func GetJobsByType(repos []string, jobType string) (map[string]map[string]types.JobConfig, error) {
	jobsListByType := map[string]map[string]types.JobConfig{}
	for _, repo := range repos {
		jobDir := filepath.Join("jobs", jobType, repo)

		jobList, err := UnmarshalJobs(jobDir)
		if err != nil {
			return nil, fmt.Errorf("error reading job directory %s: %v", jobDir, err)
		}

		jobsListByType[fmt.Sprintf("aws/%s", repo)] = jobList
	}

	return jobsListByType, nil
}

func UnmarshalJobs(jobDir string) (map[string]types.JobConfig, error) {
	jobList := map[string]types.JobConfig{}

	files, err := ioutil.ReadDir(jobDir)
	if err != nil {
		return nil, fmt.Errorf("error reading job directory %s: %v", jobDir, err)
	}

	for _, file := range files {
		fileName := file.Name()
		filePath := filepath.Join(jobDir, fileName)
		if strings.Contains(fileName, "1-X") {
			for i, releaseBranch := range releaseBranches {
				var jobConfig types.JobConfig
				releaseBranchBasedFileName := strings.ReplaceAll(fileName, "1-X", releaseBranch)
				otherReleaseBranches := append(append([]string{}, releaseBranches[:i]...),
					releaseBranches[i+1:]...)
				data := map[string]interface{}{
					"releaseBranch":        releaseBranch,
					"otherReleaseBranches": strings.Join(otherReleaseBranches, "|"),
				}

				contents, err := ioutil.ReadFile(filePath)
				if err != nil {
					return nil, fmt.Errorf("error reading job YAML %s: %v", filePath, err)
				}

				templatedContents, err := ExecuteTemplate(string(contents), data)
				if err != nil {
					return nil, fmt.Errorf("error executing template: %v", err)
				}

				err = yaml.Unmarshal(templatedContents, &jobConfig)
				if err != nil {
					return nil, fmt.Errorf("error unmarshaling contents of file %s: %v", filePath, err)
				}

				jobList[releaseBranchBasedFileName] = jobConfig
			}
		} else {
			var jobConfig types.JobConfig
			contents, err := ioutil.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("error reading job YAML %s: %v", filePath, err)
			}

			err = yaml.Unmarshal(contents, &jobConfig)
			if err != nil {
				return nil, fmt.Errorf("error unmarshaling contents of file %s: %v", filePath, err)
			}

			jobList[fileName] = jobConfig
		}
	}

	return jobList, nil
}

func ExecuteTemplate(templateContent string, data interface{}) ([]byte, error) {
	temp := template.New("template")
	funcMap := map[string]interface{}{
		"indent": func(spaces int, v string) string {
			pad := strings.Repeat(" ", spaces)
			return pad + strings.Replace(v, "\n", "\n"+pad, -1)
		},
		"stringsJoin": strings.Join,
		"trim":        strings.TrimSpace,
	}
	temp = temp.Funcs(funcMap)

	temp, err := temp.Parse(templateContent)
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %v", err)
	}

	var buf bytes.Buffer
	err = temp.Execute(&buf, data)
	if err != nil {
		return nil, fmt.Errorf("error substituting values for template: %v", err)
	}
	return buf.Bytes(), nil
}
