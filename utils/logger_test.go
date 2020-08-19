package utils_test

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/topfreegames/Will.IAM/utils"
)

func TestGetLogger(t *testing.T) {
	host := "0.0.0.0"
	port := 8080

	type testCase struct {
		TestName         string
		Verbosity        int
		LogJSON          bool
		ExpectedLogLevel logrus.Level
	}

	testCases := []testCase{
		testCase{
			TestName:         "verbosity=0,logJSON=false",
			Verbosity:        0,
			LogJSON:          false,
			ExpectedLogLevel: logrus.ErrorLevel,
		},
		testCase{
			TestName:         "verbosity=1,logJSON=false",
			Verbosity:        1,
			LogJSON:          false,
			ExpectedLogLevel: logrus.WarnLevel,
		},
		testCase{
			TestName:         "verbosity=2,logJSON=false",
			Verbosity:        2,
			LogJSON:          false,
			ExpectedLogLevel: logrus.InfoLevel,
		},
		testCase{
			TestName:         "verbosity=3,logJSON=false",
			Verbosity:        3,
			LogJSON:          false,
			ExpectedLogLevel: logrus.DebugLevel,
		},
		testCase{
			TestName:         "verbosity=99,logJSON=false",
			Verbosity:        99,
			LogJSON:          false,
			ExpectedLogLevel: logrus.InfoLevel,
		},
		testCase{
			TestName:         "verbosity=0,logJSON=true",
			Verbosity:        0,
			LogJSON:          true,
			ExpectedLogLevel: logrus.ErrorLevel,
		},
		testCase{
			TestName:         "verbosity=1,logJSON=true",
			Verbosity:        1,
			LogJSON:          true,
			ExpectedLogLevel: logrus.WarnLevel,
		},
		testCase{
			TestName:         "verbosity=2,logJSON=true",
			Verbosity:        2,
			LogJSON:          true,
			ExpectedLogLevel: logrus.InfoLevel,
		},
		testCase{
			TestName:         "verbosity=3,logJSON=true",
			Verbosity:        3,
			LogJSON:          true,
			ExpectedLogLevel: logrus.DebugLevel,
		},
		testCase{
			TestName:         "verbosity=99,logJSON=true",
			Verbosity:        99,
			LogJSON:          true,
			ExpectedLogLevel: logrus.InfoLevel,
		},
	}

	for _, testCaseData := range testCases {
		t.Run(testCaseData.TestName, func (t *testing.T) {
			fieldLogger := utils.GetLogger(host, port, testCaseData.Verbosity, testCaseData.LogJSON)
			entry := fieldLogger.(*logrus.Entry)

			if entry.Data["bind"] != host {
				t.Errorf("Expected log tag host with value [%s], got [%s]", host, entry.Data["bind"])
			}

			if entry.Data["port"] != port {
				t.Errorf("Expected log tag port with value [%s], got [%s]", host, entry.Data["port"])
			}

			if entry.Logger.Level != testCaseData.ExpectedLogLevel {
				t.Errorf("Expected log level [%s], got [%s]", testCaseData.ExpectedLogLevel, entry.Logger.Level)
			}

			if _, ok := entry.Logger.Formatter.(*logrus.JSONFormatter); testCaseData.LogJSON && !ok {
				t.Errorf("Expected log formatter [*logrus.JSONFormatter], got [%T]", entry.Logger.Formatter)
			}
		})
	}
}