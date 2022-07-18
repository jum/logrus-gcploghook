package gcploghook

import (
	"io"
	"os"

	stackdriver "github.com/andyfusniak/stackdriver-gae-logrus-plugin"
	"github.com/sirupsen/logrus"
)

func ExampleNewStackDriverHook() {
	googleProject := os.Getenv("GOOGLE_CLOUD_PROJECT")
	myLog := logrus.New()
	formatter := stackdriver.GAEStandardFormatter(
		stackdriver.WithProjectID(googleProject),
	)
	myLog.SetFormatter(formatter)
	myLog.SetOutput(os.Stdout)
	log_name := os.Getenv("LOG_NAME")
	if log_name != "" {
		hook, err := NewStackDriverHook(googleProject, log_name, os.Getenv("LOG_INSTANCEID"), os.Getenv("LOG_INSTANCENAME"), os.Getenv("LOG_INSTANCEZONE"))
		if err != nil {
			myLog.WithError(err).Fatal("StackDriver")
		}
		myLog.AddHook(hook)
		myLog.SetOutput(io.Discard)
	}
	myLog.Info("Hello world!")
}
