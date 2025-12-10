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
	"1-29",
	"1-30",
	"1-31",
	"1-32",
	"1-33",
	"1-34",
	"1-35",
}

var golangVersions = []string{
	"1-24",
	"1-25",
}

var pythonVersions = []string{
	"3-7",
	"3-9",
}

var alVersions = []string{
	"2",
	"2023",
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

func AppendMap(current map[string]interface{}, new map[string]interface{}) map[string]interface{} {
	newMap := map[string]interface{}{}
	for k, v := range current {
		newMap[k] = v
	}
	for k, v := range new {
		newMap[k] = v
	}

	return newMap
}

func AddALVersion(fileName string, data map[string]interface{}) map[string]map[string]interface{} {
	jobList := map[string]map[string]interface{}{}
	if !strings.Contains(fileName, "al-X") {
		return jobList
	}

	for _, version := range alVersions {
		alVersionFileName := strings.ReplaceAll(fileName, "al-X", "al-"+version)
		jobList[alVersionFileName] = AppendMap(data, map[string]interface{}{
			"alVersion": version,
		})
	}

	return jobList
}

func AddGolangVersion(fileName string, data map[string]interface{}) map[string]map[string]interface{} {
	jobList := map[string]map[string]interface{}{}
	if !strings.Contains(fileName, "golang-1-X") {
		return jobList
	}

	for _, version := range golangVersions {
		golangVersionFileName := strings.ReplaceAll(fileName, "golang-1-X", "golang-"+version)
		goVersion := strings.Replace(version, "-", ".", 1)
		jobList[golangVersionFileName] = AppendMap(data, map[string]interface{}{
			"jobGoVersion":  version,
			"golangVersion": goVersion,
		})
	}

	return jobList
}

func AddPythonVersion(fileName string, data map[string]interface{}) map[string]map[string]interface{} {
	jobList := map[string]map[string]interface{}{}
	if !strings.Contains(fileName, "python-3-X") {
		return jobList
	}

	for _, version := range pythonVersions {
		pythonVersionFileName := strings.ReplaceAll(fileName, "python-3-X", "python-"+version)
		pythonVersion := strings.Replace(version, "-", ".", 1)
		jobList[pythonVersionFileName] = AppendMap(data, map[string]interface{}{
			"jobPythonVersion": version,
			"pythonVersion":    pythonVersion,
		})
	}

	return jobList
}

func AddReleaseBranch(fileName string, data map[string]interface{}) map[string]map[string]interface{} {
	jobList := map[string]map[string]interface{}{}
	if !strings.Contains(fileName, "1-X") {
		return jobList
	}
	currentReleaseBranches := releaseBranches

	for i, releaseBranch := range currentReleaseBranches {

		releaseBranchBasedFileName := strings.ReplaceAll(fileName, "1-X", releaseBranch)
		otherReleaseBranches := append(append([]string{}, currentReleaseBranches[:i]...),
			currentReleaseBranches[i+1:]...)
		jobList[releaseBranchBasedFileName] = AppendMap(data, map[string]interface{}{
			"releaseBranch":        releaseBranch,
			"otherReleaseBranches": strings.Join(otherReleaseBranches, "|"),
		})

		// If latest release branch, check if the release branch dir exists before executing cmd
		// This allows us to experiment with adding prow jobs for new branches without failing other runs
		if len(currentReleaseBranches)-1 == i {
			jobList[releaseBranchBasedFileName]["latestReleaseBranch"] = true
		}
	}

	return jobList
}

func RunMappers(jobsToData map[string]map[string]interface{}, mappers []func(string, map[string]interface{}) map[string]map[string]interface{}) {
	if len(mappers) == 0 {
		return
	}

	for fileName, data := range jobsToData {
		newJobList := mappers[0](fileName, data)
		if len(newJobList) == 0 {
			continue
		}

		for k, v := range newJobList {
			jobsToData[k] = v
			if _, ok := data["templateFileName"]; !ok {
				jobsToData[k]["templateFileName"] = fileName
			}
		}
		delete(jobsToData, fileName)
	}

	RunMappers(jobsToData, mappers[1:])
}

func UnmarshalJobs(jobDir string) (map[string]types.JobConfig, error) {
	files, err := ioutil.ReadDir(jobDir)
	if err != nil {
		return nil, fmt.Errorf("error reading job directory %s: %v", jobDir, err)
	}

	var mappers []func(string, map[string]interface{}) map[string]map[string]interface{}
	mappers = append(mappers, AddALVersion, AddGolangVersion, AddPythonVersion, AddReleaseBranch)

	jobsToData := map[string]map[string]interface{}{}
	for _, file := range files {
		jobsToData[file.Name()] = map[string]interface{}{}
	}

	RunMappers(jobsToData, mappers)

	finalJobList := map[string]types.JobConfig{}
	for fileName, data := range jobsToData {
		templateFileName := fileName
		if name, ok := data["templateFileName"]; ok {
			templateFileName = name.(string)
		}

		jobConfig, err := GenerateJobConfig(data, filepath.Join(jobDir, templateFileName))
		if err != nil {
			return nil, fmt.Errorf("%v", err)
		}

		if latest, ok := data["latestReleaseBranch"]; ok && latest.(bool) {
			if !jobConfig.SkipReleaseBranchCheck {
				for j, command := range jobConfig.Commands {
					jobConfig.Commands[j] = "if make check-for-supported-release-branch -C $PROJECT_PATH; then " + command + "; fi"
				}
			}
		}

		finalJobList[fileName] = jobConfig

	}
	return finalJobList, nil
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

func GenerateJobConfig(data interface{}, filePath string) (types.JobConfig, error) {
	var jobConfig types.JobConfig
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return jobConfig, fmt.Errorf("error reading job YAML %s: %v", filePath, err)
	}
	var templatedContents []byte
	if data != nil {
		templatedContents, err = ExecuteTemplate(string(contents), data)
		if err != nil {
			return jobConfig, fmt.Errorf("error executing template: %v", err)
		}
		err = yaml.Unmarshal(templatedContents, &jobConfig)
		if err != nil {
			return jobConfig, fmt.Errorf("error unmarshaling contents of file %s: %v", filePath, err)
		}
	} else {
		err = yaml.Unmarshal(contents, &jobConfig)
		if err != nil {
			return jobConfig, fmt.Errorf("error unmarshaling contents of file %s: %v", filePath, err)
		}
	}
	return jobConfig, nil
}
