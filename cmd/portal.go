package cmd

import (
	"log"
	"sync"

	"github.com/spf13/cobra"
	"github.com/zuiwuchang/reverse/configure"
	"github.com/zuiwuchang/reverse/portal"
)

func init() {
	var filename string
	cmd := &cobra.Command{
		Use:   `portal`,
		Short: `Receive user requests and reverse proxy to 'bridge'`,
		Run: func(cmd *cobra.Command, args []string) {
			cnfs, e := configure.LoadPortal(filename)
			if e != nil {
				log.Fatalln(e)
			}
			count := len(cnfs)
			if count == 0 {
				return
			}
			items := make([]*portal.Portal, 0, count)
			for i := 0; i < count; i++ {
				item, e := portal.New(&cnfs[i])
				if e != nil {
					log.Fatalln(e)
				}
				items = append(items, item)
			}
			if count == 0 {
				items[0].Serve()
			} else {
				var wait sync.WaitGroup
				wait.Add(count)
				for _, item := range items {
					go func(item *portal.Portal) {
						defer wait.Done()
						item.Serve()
					}(item)
				}
				wait.Wait()
			}
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&filename, `conf`, `c`,
		`portal.jsonnet`,
		`configure file path`,
	)
	root.AddCommand(cmd)
}
