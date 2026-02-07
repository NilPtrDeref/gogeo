package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "gogeo",
	Short: "An application that allows users to analyze and manage US geography data.",
	Long: `An application to parse, simplify (clean), enrich, store, serve, project, visualize, and modify
US geography data. The aim is to support the entire pipeline from parsing data sourced from the US Census
Bureau all the way through visualizing the data and allowing users to interact with it in a browser.`,
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
