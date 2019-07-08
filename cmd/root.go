package cmd

import (
	"fmt"
	"os"

	"github.com/topfreegames/Will.IAM/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgPath string
var verbose int
var json bool
var config *viper.Viper

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "Will.IAM",
	Short: "Will.IAM",
	Long:  `Will.IAM`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().BoolVarP(
		&json, "json", "j",
		false, "json output mode")

	RootCmd.PersistentFlags().IntVarP(
		&verbose, "verbose", "v", 0,
		"Verbosity level => v0: Error, v1=Warning, v2=Info, v3=Debug",
	)

	RootCmd.PersistentFlags().StringVarP(
		&cfgPath, "config", "c", "./config/local.yaml",
		"config file",
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	cfg, err := utils.GetConfig(cfgPath)
	if err != nil {
		fmt.Printf("Config file %s failed to load: %s.\n", cfgPath, err.Error())
	} else {
		config = cfg
	}
}
