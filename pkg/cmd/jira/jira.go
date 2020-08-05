package jira

import (
	"github.com/Benbentwo/atlassian-cli/pkg/cmd/common"
	"github.com/coryb/figtree"
	"github.com/coryb/oreo"
	"github.com/go-jira/jira/jiracli"
	"github.com/go-jira/jira/jiracmd"
	"github.com/spf13/cobra"
	"gopkg.in/coryb/yaml.v2"
	"gopkg.in/op/go-logging.v1"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
)

// options for the command
type JiraOptions struct {
	*common.CommonOptions
}

func NewCmdJira(commonOpts *common.CommonOptions) *cobra.Command {
	options := &JiraOptions{
		CommonOptions: commonOpts,
	}

	cmd := &cobra.Command{
		Use: "jira",
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			err := options.Run()
			common.CheckErr(err)
		},
	}
	// the line below (Section to...) is for the generate-function command to add a template_command to.
	// the line above this and below can be deleted.
	// DO NOT DELETE THE FOLLOWING LINE:
	// Section to add commands to:
	cmd.AddCommand(NewCmdJiraConfigure(commonOpts))
	cmd.DisableFlagParsing = true

	return cmd
}

type oreoLogger struct {
	logger *logging.Logger
}

var log = logging.MustGetLogger("jira")

func (ol *oreoLogger) Printf(format string, args ...interface{}) {
	ol.logger.Debugf(format, args...)
}

// Run implements this command
func (o *JiraOptions) Run() error {
	// Copy from go-jira/jira/

	defer jiracli.HandleExit()

	jiracli.InitLogging()

	configDir := ".jira.d"

	yaml.UseMapType(reflect.TypeOf(map[string]interface{}{}))
	defer yaml.RestoreMapType()

	fig := figtree.NewFigTree(
		figtree.WithHome(jiracli.Homedir()),
		figtree.WithEnvPrefix("JIRA"),
		figtree.WithConfigDir(configDir),
	)

	if err := os.MkdirAll(filepath.Join(jiracli.Homedir(), configDir), 0755); err != nil {
		log.Errorf("%s", err)
		panic(jiracli.Exit{Code: 1})
	}

	ore := oreo.New().WithCookieFile(filepath.Join(jiracli.Homedir(), configDir, "cookies.js")).WithLogger(&oreoLogger{log})

	jiracmd.RegisterAllCommands()

	app := jiracli.CommandLine(fig, ore)
	//jiracli.ParseCommandLine(app, os.Args[2:])
	ctx, err := app.ParseContext(os.Args[2:])
	if err != nil && ctx == nil {
		// This is an internal kingpin usage error, duplicate options/commands
		log.Errorf("error: %s, ctx: %v", err, ctx)
	}
	if ctx != nil {
		if ctx.SelectedCommand == nil {
			next := ctx.Next()
			if next != nil {
				if ok, err := regexp.MatchString("^([A-Z]+-)?[0-9]+$", next.Value); err != nil {
					log.Errorf("Invalid Regex: %s", err)
				} else if ok {
					// insert "view" at i=1 (2nd position)
					os.Args = append(os.Args[:], append([]string{"view"}, os.Args[2:]...)...)
				}
			}
		}
	}

	if _, err := app.Parse(os.Args[2:]); err != nil {
		if _, ok := err.(*jiracli.Error); ok {
			log.Errorf("%s", err)
			panic(jiracli.Exit{Code: 1})
		}
		ctx, _ := app.ParseContext(os.Args[2:])
		if ctx != nil {
			app.UsageForContext(ctx)
		}
		log.Errorf("Invalid Usage: %s", err)
		panic(jiracli.Exit{Code: 1})
	}
	//jiracli.ParseCommandLine(app, os.Args[2:])
	return nil
}

var usage = `{{define "FormatCommand"}}\
{{if .FlagSummary}} {{.FlagSummary}}{{end}}\
{{range .Args}} {{if not .Required}}[{{end}}<{{.Name}}>{{if .Value|IsCumulative}}...{{end}}{{if not .Required}}]{{end}}{{end}}\
{{end}}\

{{define "FormatBriefCommands"}}\
{{range .FlattenedCommands}}\
{{if not .Hidden}}\
  {{ print .FullCommand ":" | printf "%-20s"}} {{.Help}}
{{end}}\
{{end}}\
{{end}}\

{{define "FormatCommands"}}\
{{range .FlattenedCommands}}\
{{if not .Hidden}}\
  {{.FullCommand}}{{if .Default}}*{{end}}{{template "FormatCommand" .}}
{{.Help|Wrap 4}}
{{with .Flags|FlagsToTwoColumns}}{{FormatTwoColumnsWithIndent . 4 2}}{{end}}
{{end}}\
{{end}}\
{{end}}\

{{define "FormatUsage"}}\
{{template "FormatCommand" .}}{{if .Commands}} <command> [<args> ...]{{end}}
{{if .Help}}
{{.Help|Wrap 0}}\
{{end}}\

{{end}}\

{{if .Context.SelectedCommand}}\
usage: {{.App.Name}} {{.Context.SelectedCommand}}{{template "FormatCommand" .Context.SelectedCommand}}
{{if .Context.SelectedCommand.Aliases }}\
{{range $top := .App.Commands}}\
{{if eq $top.FullCommand $.Context.SelectedCommand.FullCommand}}\
{{range $alias := $.Context.SelectedCommand.Aliases}}\
alias: {{$.App.Name}} {{$alias}}{{template "FormatCommand" $.Context.SelectedCommand}}
{{end}}\
{{else}}\
{{range $sub := $top.Commands}}\
{{if eq $sub.FullCommand $.Context.SelectedCommand.FullCommand}}\
{{range $alias := $.Context.SelectedCommand.Aliases}}\
alias: {{$.App.Name}} {{$top.Name}} {{$alias}}{{template "FormatCommand" $.Context.SelectedCommand}}
{{end}}\
{{end}}\
{{end}}\
{{end}}\
{{end}}\
{{end}}
{{if .Context.SelectedCommand.Help}}\
{{.Context.SelectedCommand.Help|Wrap 0}}
{{end}}\
{{else}}\
usage: {{.App.Name}}{{template "FormatUsage" .App}}
{{end}}\

{{if .App.Flags}}\
Global flags:
{{.App.Flags|FlagsToTwoColumns|FormatTwoColumns}}
{{end}}\
{{if .Context.SelectedCommand}}\
{{if and .Context.SelectedCommand.Flags|RequiredFlags}}\
Required flags:
{{.Context.SelectedCommand.Flags|RequiredFlags|FlagsToTwoColumns|FormatTwoColumns}}
{{end}}\
{{if .Context.SelectedCommand.Flags|OptionalFlags}}\
Optional flags:
{{.Context.SelectedCommand.Flags|OptionalFlags|FlagsToTwoColumns|FormatTwoColumns}}
{{end}}\
{{end}}\
{{if .Context.Args}}\
Args:
{{.Context.Args|ArgsToTwoColumns|FormatTwoColumns}}
{{end}}\
{{if .Context.SelectedCommand}}\
{{if .Context.SelectedCommand.Commands}}\
Subcommands:
{{template "FormatCommands" .Context.SelectedCommand}}
{{end}}\
{{else if .App.Commands}}\
Commands:
{{template "FormatBriefCommands" .App}}
{{end}}\
`
