package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"

	. "github.com/logrusorgru/aurora"
	"gopkg.in/yaml.v2"
)

type envConfig struct {
	Name    string
	Module  string
	Params  map[string]string
	Targets map[string]Target
}

type Target struct {
	Name   string
	Module string
	Params map[string]string
}

type Module struct {
	Actions map[string]struct {
		RequiredParams []string `yaml:"requiredParams"`
		RequiredCmds   []string `yaml:"requiredCmds"`
	}
}

//
//
//
func main() {
	var action string
	var file string
	flag.StringVar(&action, "action", "", "help message for flag n")
	flag.StringVar(&action, "a", "", "help message for flag n")
	flag.StringVar(&file, "file", "", "help message for flag n")
	flag.StringVar(&file, "f", "", "help message for flag n")
	flag.Parse()

	if action == "" {
		checkFatal(fmt.Errorf("must supply an action, e.g. -action deploy"))
	}
	if file == "" {
		checkFatal(fmt.Errorf("must supply a file, e.g. -file foo.yaml"))
	}
	fmt.Println("---- STARTING ", action)

	envConfig := envConfig{}
	envYaml, err := ioutil.ReadFile(file)
	checkFatal(err)
	err = yaml.Unmarshal([]byte(envYaml), &envConfig)
	checkFatal(err)

	// Validation checks
	if envConfig.Module == "" {
		checkFatal(fmt.Errorf("no module set for env"))
	}

	// Run init for environment
	runModule("env", envConfig.Module, "init", envConfig.Params, envConfig.Name)

	// Run init for all targets
	for name, target := range envConfig.Targets {
		target.Name = name

		if target.Module == "" {
			checkFatal(fmt.Errorf("no module set for target %s", name))
		}

		runModule("target", target.Module, "init", target.Params, name)
	}

	// Run post for env
	runModule("env", envConfig.Module, "post", envConfig.Params, envConfig.Name)
}

//
//
//
func runModule(kind string, module string, action string, params map[string]string, name string) {
	fmt.Printf("%s 🚀 Starting '%s-%s' for '%s/%s'\n", Green("»»»"), action, kind, module, name)
	scriptPath := fmt.Sprintf("modules/%s/%s/%s.sh", kind, module, action)
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		checkWarn(fmt.Errorf("Module script not found: %s", scriptPath))
		return
	}

	moduleConfig := Module{}
	moduleYaml, err := ioutil.ReadFile(fmt.Sprintf("modules/%s/%s/module.yaml", kind, module))
	checkFatal(err)
	err = yaml.Unmarshal([]byte(moduleYaml), &moduleConfig)
	checkFatal(err)

	for _, reqParamName := range moduleConfig.Actions[action].RequiredParams {
		found := false
		for paramName := range params {
			if paramName == reqParamName {
				found = true
				break
			}
		}
		if !found {
			checkFatal(fmt.Errorf("required parameter %s not found", reqParamName))
		}
	}

	for _, reqCmdName := range moduleConfig.Actions[action].RequiredCmds {
		_, err := exec.LookPath(reqCmdName)
		if err != nil {
			checkFatal(fmt.Errorf("required command %s not found in system path", reqCmdName))
		}
	}

	// Pass in params from the config to the env vars of the script
	cmd := exec.Command("bash", scriptPath)
	for paramName, paramValue := range params {
		cmd.Env = append(cmd.Env,
			fmt.Sprintf("%s=%s", paramName, paramValue),
		)
	}

	// Append system env vars, pretty rare you *wouldn't* want these
	cmd.Env = append(cmd.Env, os.Environ()...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Actually run the damn thing
	if err := cmd.Run(); err != nil {
		// On useful messsages might in stdout OR stderr
		fmt.Printf("%s", Red(stdout.String()))
		fmt.Printf("%s", Red(stderr.String()))

		// Fail on error, TODO: Add a continue option?
		checkFatal(err)
	}

	// Magic string syntax to allow scripts to set parameters from stdout
	// Just like AzDO or GH actions
	re := regexp.MustCompile(`:::setParam ([A-Za-z_]+?)="(.*?)"`)
	varMatchResult := re.FindAllStringSubmatch(stdout.String(), -1)
	for i := range varMatchResult {
		varName := varMatchResult[i][1]
		varValue := varMatchResult[i][2]
		params[varName] = varValue
		fmt.Printf("%s 💡 setParam found in output, setting parameter '%s'\n", Blue("»»»"), varName)
	}

	//fmt.Println(res, stdout.String())
}

//
//
//
func checkFatal(err error) {
	if err != nil {
		fmt.Printf("%s 💥 %s\n", Red("»»»"), err)
		os.Exit(1)
	}
}

//
//
//
func checkWarn(err error) {
	if err != nil {
		fmt.Printf("%s ⚡ %s\n", Yellow("»»»"), err)
	}
}
