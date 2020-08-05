package jira

import (
	"fmt"
	"github.com/Benbentwo/atlassian-cli/pkg/cmd/common"
	"github.com/Benbentwo/utils/util"
	"github.com/coryb/figtree"
	"github.com/ghodss/yaml"
	"github.com/go-jira/jira/jiracli"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
)

// options for the command
//noinspection ALL
type JiraConfigureOptions struct {
	*common.CommonOptions
	batch bool
}

var jsonKeys = []string{"endpoint", "password-source", "user", "project", "editor"}

type JiraConfig struct {
	Endpoint       string `json:"endpoint,omitempty"`
	PasswordSource string `json:"password-source,omitempty"`
	User           string `json:"user,omitempty"`
	Project        string `json:"project,omitempty"`
	Editor         string `json:"editor,omitempty"`
}

var (
	jiraConfigureLong    = `runs some prompts to configure your local config.yml for jira`
	jiraConfigureExample = `atl jira configure`
)

func NewCmdJiraConfigure(commonOpts *common.CommonOptions) *cobra.Command {
	options := &JiraConfigureOptions{
		CommonOptions: commonOpts,
	}

	cmd := &cobra.Command{
		Use:     "configure",
		Short:   "configure your default settings",
		Long:    jiraConfigureLong,
		Example: jiraConfigureExample,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			common.CheckErr(err)
		},
	}

	return cmd
}

// Run implements this command
func (o *JiraConfigureOptions) Run() error {
	configDir := ".jira.d"
	fig := figtree.NewFigTree(
		figtree.WithHome(jiracli.Homedir()),
		figtree.WithEnvPrefix("JIRA"),
		figtree.WithConfigDir(configDir),
	)

	if err := os.MkdirAll(filepath.Join(jiracli.Homedir(), configDir), 0755); err != nil {
		log.Errorf("%s", err)
		panic(jiracli.Exit{Code: 1})
	}

	var tmp map[string]interface{}
	err := fig.LoadAllConfigs("config.yml", &tmp)
	if err != nil {
		util.Logger().Errorf("Error loading configs: %s", err)
	}
	updateConfigMapping(&tmp, "endpoint", "What is the endpoint for your atlassian server?", "should be something like https://company.atlassian.net/", "")
	updateConfigMapping(&tmp, "user", "What is your username", "email address for atlassian", "")
	updateConfigMapping(&tmp, "password-source", "Password Source?", "[keyring | pass | gopass]", "keyring")
	jc, _ := LoadConfigFromMap(tmp)
	err = WriteConfig(jc, util.StripTrailingSlash(filepath.Join(jiracli.Homedir(), configDir))+"/config.yml")
	if err != nil {
		util.Logger().Errorf("saving file failed: %s", err)
	}
	//updateConfigMapping(&tmp, "endpoint", "What is the endpoint for your atlassian server?", "should be something like https://company.atlassian.net/", "")
	//updateConfigMapping(&tmp, "endpoint", "What is the endpoint for your atlassian server?", "should be something like https://company.atlassian.net/", "")

	return nil
}

func updateConfigMapping(tmp *map[string]interface{}, key string, prompt string, help string, defaultChoice string) {
	mapping := *tmp

	ok := mapping[key]
	util.Logger().Debugf("Key: %s, OK: %s", key, ok)
	if ok == nil {
		resp, err := util.PromptValue(prompt, defaultChoice, help)
		if err != nil {
			util.Logger().Errorf("response error: %s", err)
		}
		util.Logger().Debugf("Response: %s", resp)
		//tmp[key] = resp
	} else {
		util.Logger().Printf("Found KV Pair |%s: %s", util.ColorInfo(key), util.ColorInfo(ok))
	}
}

func WriteConfig(config *JiraConfig, fileName string) error {
	if fileName == "" {
		return fmt.Errorf("no filename defined")
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fileName, data, util.DefaultWritePermissions)
}

func LoadConfig(fileName string) (*JiraConfig, error) {
	jc := &JiraConfig{}
	if fileName != "" {
		exists, err := util.FileExists(fileName)
		if err != nil {
			return jc, fmt.Errorf("Could not check if file exists %s due to %s", fileName, err)
		}
		if exists {
			data, err := ioutil.ReadFile(fileName)
			if err != nil {
				return jc, fmt.Errorf("Failed to load file %s due to %s", fileName, err)
			}
			err = yaml.Unmarshal(data, jc)
			if err != nil {
				return jc, fmt.Errorf("Failed to unmarshal YAML file %s due to %s", fileName, err)
			}
		}
	}
	return jc, nil
}

func LoadConfigFromMap(mapping map[string]interface{}) (*JiraConfig, error) {
	endpoint := mapping["endpoint"]
	passwordSource := mapping["password-source"]
	user := mapping["user"]
	project := mapping["project"]
	editor := mapping["editor"]

	jc := &JiraConfig{
		Endpoint:       endpoint.(string),
		PasswordSource: passwordSource.(string),
		User:           user.(string),
		Project:        project.(string),
		Editor:         editor.(string),
	}
	return jc, nil
}
