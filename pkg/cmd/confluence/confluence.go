package confluence

import (
	"github.com/Benbentwo/atlassian-cli/pkg/cmd/common"
	m2c "github.com/Benbentwo/go-markdown2confluence/cmd"
	"github.com/spf13/cobra"
)

// options for the command
type ConfluenceOptions struct {
	*common.CommonOptions
}

func NewCmdConfluence(commonOpts *common.CommonOptions) *cobra.Command {
	options := &ConfluenceOptions{
		CommonOptions: commonOpts,
	}

	cmd := &cobra.Command{
		Use: "confluence",
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			common.CheckErr(err)
		},
	}
	options.AddConfluenceFlags(cmd)
	// Section to add commands to:
	markdownCmd := m2c.RootCmd
	markdownCmd.Use = `m2c`

	cmd.AddCommand(m2c.RootCmd)

	return cmd
}

// Run implements this command
func (o *ConfluenceOptions) Run() error {
	return o.Cmd.Help()
}

func (o *ConfluenceOptions) AddConfluenceFlags(cmd *cobra.Command) {
	o.Cmd = cmd
}
