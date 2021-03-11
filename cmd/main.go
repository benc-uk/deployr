package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	. "github.com/logrusorgru/aurora"
	"gopkg.in/yaml.v2"
)

type envConfig struct {
	Name    string
	Plugin  string
	Params  map[string]string
	Targets map[string]Target
}

type Target struct {
	Name   string
	Plugin string
	Params map[string]string
}

func main() {
	envConfig := envConfig{}

	envYaml, err := ioutil.ReadFile("myenv.yaml")
	check(err)

	err = yaml.Unmarshal([]byte(envYaml), &envConfig)
	check(err)

	for name, target := range envConfig.Targets {
		target.Name = name
		runModule("target", "aks", "init", target.Params, name)
	}
}

func runModule(kind string, plugin string, action string, params map[string]string, name string) {
	fmt.Printf("%s 🚀 Starting %s module %s/%s for '%s' \n", Green("»»»"), kind, plugin, action, name)
	scriptPath := fmt.Sprintf("modules/%s/%s/%s.sh", kind, plugin, action)
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		check(fmt.Errorf("Module script not found: %s", scriptPath))
	}

	file, err := os.Open(scriptPath)
	check(err)
	scanner := bufio.NewScanner(file)
	requiredParams := []string{}
	requiredCmds := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#_PARAMS") {
			requiredParams = append(requiredParams, strings.Split(line, " ")...)
		}
		if strings.HasPrefix(line, "#_CMDS") {
			requiredCmds = append(requiredCmds, strings.Split(line, " ")...)
		}
	}

	for _, reqParamName := range requiredParams {
		reqParamName = strings.TrimSpace(reqParamName)
		if reqParamName != "" && reqParamName != "#_PARAMS" {
			found := false
			for paramName := range params {
				if paramName == reqParamName {
					found = true
					break
				}
			}
			if !found {
				check(fmt.Errorf("Required parameter %s not found", reqParamName))
			}
		}
	}

	for _, reqCmdName := range requiredCmds {
		reqCmdName = strings.TrimSpace(reqCmdName)
		if reqCmdName == "#_CMDS" {
			continue
		}
		_, err := exec.LookPath(reqCmdName)
		if err != nil {
			check(fmt.Errorf("Required command %s not found in system path", reqCmdName))
		}
	}

	cmd := exec.Command("bash", scriptPath)
	for paramName, paramValue := range params {
		cmd.Env = append(cmd.Env,
			fmt.Sprintf("%s=%s", paramName, paramValue),
		)
	}
	cmd.Env = append(cmd.Env, os.Environ()...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("%s", Red(stderr.String()))
		check(err)
	}

	fmt.Println(stdout.String())
}

func check(err error) {
	if err != nil {
		//fmt.Println(Red(fmt.Sprintf("»»» %s", err.Error())))
		fmt.Printf("%s 💥 %s\n", Red("»»»"), err)
		os.Exit(1)
	}
}
