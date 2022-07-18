# gcploghook

I used the blog artikle from Huy Ngo 
[4 tips for logging on GCP using golang and logrus](https://huynvk.dev/blog/4-tips-for-logging-on-gcp-using-golang-and-logrus)
as a base for this small module. It is basically all that is
needed to log from a go app to Google Cloud Logging. An example how
to use this in a go app:

```go
package main

import (
    "os"

    stackdriver "github.com/andyfusniak/stackdriver-gae-logrus-plugin"
    "github.com/sirupsen/logrus"
    "github.com/jum/logrus-gcploghook"
)

main() |{
    googleProject := os.Getenv("GOOGLE_CLOUD_PROJECT")
    myLog := logrus.New()
    formatter := stackdriver.GAEStandardFormatter(
        stackdriver.WithProjectID(googleProject),
    )
    myLog.SetFormatter(formatter)
    myLog.SetOutput(os.Stdout)
    log_name := os.Getenv("LOG_NAME")
    if log_name != "" {
        hook, err := gcploghook.NewStackDriverHook(app.GoogleProject, log_name, os.Getenv("LOG_INSTANCEID"), os.Getenv("LOG_INSTANCENAME"), os.Getenv("LOG_INSTANCEZONE"))
        if err != nil {
            myLog.WithError(err).Fatal("StackDriver")
        }
        myLog.AddHook(hook)
        myLog.SetOutput(io.Discard)
    }
    myLog.Info("Hello world!")
}
```

The environment variable `GOOGLE_APPLICATION_CREDENTIALS` needs to
point to your credential .json file downloaded from Google Cloud
Console.
