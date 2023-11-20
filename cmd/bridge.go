package cmd

import (
	"log"
	"sync"

	"github.com/spf13/cobra"
	"github.com/zuiwuchang/reverse/bridge"
	"github.com/zuiwuchang/reverse/configure"
)

func init() {
	var filename string
	cmd := &cobra.Command{
		Use:   `bridge`,
		Short: `Connect 'portal' to backend network`,
		Run: func(cmd *cobra.Command, args []string) {
			cnfs, e := configure.LoadBridge(filename)
			if e != nil {
				log.Fatalln(e)
			}
			count := len(cnfs)
			if count == 0 {
				return
			}
			items := make([]*bridge.Bridge, 0, count)
			for i := 0; i < count; i++ {
				item, e := bridge.New(&cnfs[i])
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
					go func(item *bridge.Bridge) {
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
		`bridge.jsonnet`,
		`configure file path`,
	)
	root.AddCommand(cmd)
}
