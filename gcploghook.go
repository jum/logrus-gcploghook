package gcploghook

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"

	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/logging"
	"github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

type InstanceInfo struct {
	Zone string `json:"zone,omitempty"`
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
}

type StackDriverHook struct {
	client       *logging.Client
	errorClient  *errorreporting.Client
	logger       *logging.Logger
	instanceInfo *InstanceInfo
}

var logLevelMappings = map[logrus.Level]logging.Severity{
	logrus.TraceLevel: logging.Default,
	logrus.DebugLevel: logging.Debug,
	logrus.InfoLevel:  logging.Info,
	logrus.WarnLevel:  logging.Warning,
	logrus.ErrorLevel: logging.Error,
	logrus.FatalLevel: logging.Critical,
	logrus.PanicLevel: logging.Critical,
}

func NewStackDriverHook(googleProject string, logName string, logInstanceID string, logInstanceName string, logInstanceZone string) (*StackDriverHook, error) {
	ctx := context.Background()

	client, err := logging.NewClient(ctx, googleProject)
	if err != nil {
		return nil, err
	}

	errorClient, err := errorreporting.NewClient(ctx, googleProject, errorreporting.Config{
		ServiceName: googleProject,
		OnError: func(err error) {
			fmt.Fprintf(os.Stderr, "Could not log error: %v", err)
		},
	})
	if err != nil {
		return nil, err
	}

	instanceInfo := &InstanceInfo{
		ID:   logInstanceID,
		Name: logInstanceName,
		Zone: logInstanceZone,
	}
	if len(instanceInfo.Name) == 0 && len(instanceInfo.ID) == 0 && len(instanceInfo.Zone) == 0 {
		instanceInfo = nil
	}
	options := []logging.LoggerOption{}
	if instanceInfo != nil {
		vmMrpb := logging.CommonResource(
			&monitoredres.MonitoredResource{
				Type: "gce_instance",
				Labels: map[string]string{
					"instance_id": instanceInfo.ID,
					"zone":        instanceInfo.Zone,
				},
			},
		)
		options = []logging.LoggerOption{vmMrpb}
	}
	logger := client.Logger(logName, options...)

	return &StackDriverHook{
		client:       client,
		errorClient:  errorClient,
		logger:       logger,
		instanceInfo: instanceInfo,
	}, nil
}

func (sh *StackDriverHook) Close() {
	sh.client.Close()
	sh.errorClient.Close()
}

func (sh *StackDriverHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (sh *StackDriverHook) Fire(entry *logrus.Entry) error {
	payload := map[string]interface{}{
		"message":  entry.Message,
		"data":     entry.Data,
		"instance": sh.instanceInfo,
	}
	level := logLevelMappings[entry.Level]
	var err error
	var ok bool
	if err, ok = entry.Data[logrus.ErrorKey].(error); ok {
		entry.Data[logrus.ErrorKey] = err.Error()
	}
	sh.logger.Log(logging.Entry{Payload: payload, Severity: level})
	if int(level) >= int(logging.Error) {
		if err == nil {
			err = errors.New(entry.Message)
		}
		sh.errorClient.Report(errorreporting.Entry{
			Error: err,
			Stack: sh.getStackTrace(),
		})
	}
	return nil
}

func (sh *StackDriverHook) getStackTrace() []byte {
	stackSlice := make([]byte, 2048)
	length := runtime.Stack(stackSlice, false)
	stack := string(stackSlice[0:length])
	re := regexp.MustCompile("[\r\n].*logrus.*")
	res := re.ReplaceAllString(stack, "")
	return []byte(res)
}

func (sh *StackDriverHook) Wait() {
}
