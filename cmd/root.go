package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zuiwuchang/reverse/version"
)

const App = `reverse`

func createRoot() *cobra.Command {
	var v bool
	cmd := &cobra.Command{
		Use:   App,
		Short: `A reverse proxy to help you expose a local server behind a NAT or firewall to the internet`,
		Run: func(cmd *cobra.Command, args []string) {
			if v {
				fmt.Println(version.Platform)
				fmt.Println(version.Version)
				fmt.Println(version.Commit)
				fmt.Println(version.Date)
			} else {
				fmt.Println(version.Platform)
				fmt.Println(version.Version)
				fmt.Println(version.Commit)
				fmt.Println(version.Date)
				fmt.Println(`Use "` + App + ` --help" for more information about this program.`)
			}
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&v, `version`, `v`,
		false,
		`display version`,
	)
	return cmd
}

var root = createRoot()

func Execute() error {
	return root.Execute()
}
