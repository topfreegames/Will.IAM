package cmd

import (
	"github.com/topfreegames/Will.IAM/api"
	"github.com/topfreegames/Will.IAM/constants"
	"github.com/topfreegames/Will.IAM/utils"
	"github.com/spf13/cobra"
)

// startAPICmd represents the start-api command
var startAPICmd = &cobra.Command{
	Use:   "start-api",
	Short: "starts the api",
	Long:  `starts the api.`,
	Run: func(cmd *cobra.Command, args []string) {
		constants.Set(config)
		log := utils.GetLogger(bind, port, verbose, json)
		log.Info("starting Will.IAM")
		app, err := api.NewApp(bind, port, config, log, nil)
		if err != nil {
			log.Panic(err.Error())
		}

		app.ListenAndServe()
	},
}

var bind string
var port int

func init() {
	startAPICmd.Flags().StringVarP(&bind, "host", "b", "0.0.0.0", "bind address")
	startAPICmd.Flags().IntVarP(&port, "port", "p", 4040, "bind port")
	RootCmd.AddCommand(startAPICmd)
}
