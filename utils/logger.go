package utils

import (
	"github.com/topfreegames/Will.IAM/constants"
	"github.com/sirupsen/logrus"
)

// GetLogger returns a logrus.FieldLogger with bind, port, verbosity,
// and logJSON set
func GetLogger(bind string, port, verbosity int, logJSON bool) logrus.FieldLogger {
	log := logrus.New()
	switch verbosity {
	case 0:
		log.Level = logrus.InfoLevel
	case 1:
		log.Level = logrus.WarnLevel
	case 3:
		log.Level = logrus.DebugLevel
	default:
		log.Level = logrus.InfoLevel
	}
	if logJSON {
		log.Formatter = new(logrus.JSONFormatter)
	}
	fieldLogger := log.WithFields(logrus.Fields{
		"source":  constants.AppInfo.Name,
		"version": constants.AppInfo.Version,
		"bind":    bind,
		"port":    port,
	})
	return fieldLogger
}
