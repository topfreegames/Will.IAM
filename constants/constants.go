package constants

import "github.com/spf13/viper"

// Metrics constants
var Metrics = struct {
	ResponseTime string
}{
	ResponseTime: "response_time",
}

// AppInfo constants
var AppInfo = struct {
	Name    string
	Version string
}{
	Name:    "Will.IAM",
	Version: "1.0",
}

// constants from config
var (
	DefaultListOptionsPageSize int
)

// Set is called at start.Run
func Set(config *viper.Viper) {
	DefaultListOptionsPageSize = config.GetInt("listOptions.defaultPageSize")
}
