package commands

import (
	"fmt"
	
	"github.com/spf13/cobra"
)

// addSearch implements the search command
var addSearch = &cobra.Command{
	Use:   "search",
	Short: "search kubernetes manifest yaml",
	Long: `search kubernetes manifest yaml
It assumes that the directory supplied has the manifests
in it in YAML format.`,
	Run: func(cmd *cobra.Command, args []string) {
		
		dir, err := cmd.Flags().GetString("dir")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Search dir ", dir)
	},
}

func init() {
	rootCmd.AddCommand(addSearch)
	addSearch.Flags().StringP("dir", "d", "", "search directory")
}

