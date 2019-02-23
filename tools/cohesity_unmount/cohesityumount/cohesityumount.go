// Copyright 2019 Cohesity Inc.

package cohesityumount

import (
  "errors"
  "flag"
  "os"

  "github.com/cohesity/app-sdk-go/appsdk"
  "github.com/golang/glog"
)

var (
  // FLAGS_unmountDir specifies the directory that is to be unmounted.
  FLAGS_unmountDir string

  // The following are read from environment variable during init.

  // IP address of the host on which the container is running.
  apiEndpointIp string

  // Port of the cohesity app server
  apiEndpointPort string

  // Authentication token the app uses to authenticate with the service.
  appAuthenticationToken string

  // appClient to make cohesity appserver api calls.
  appClient CohesityAppSdk.COHESITYAPPSDK
)

func Init() error {
  // Initializing flag parameters.
  flag.StringVar(&FLAGS_unmountDir, "unmountdir", "",
    "Directory that is to be unmounted")

  // Read the environment variables.
  apiEndpointIp = os.Getenv("APPS_API_ENDPOINT_IP")
  if len(apiEndpointIp) == 0 {
    errorMsg := "APPS_API_ENDPOINT_IP not set"
    glog.Errorln(errorMsg)
    return errors.New(errorMsg)
  }

  if apiEndpointPort =
    os.Getenv("APPS_API_ENDPOINT_PORT"); len(apiEndpointPort) == 0 {
    errorMsg := "APPS_API_ENDPOINT_PORT not set"
    glog.Errorln(errorMsg)
    return errors.New(errorMsg)
  }

  if appAuthenticationToken =
    os.Getenv("APP_AUTHENTICATION_TOKEN"); len(appAuthenticationToken) == 0 {
    errorMsg := "APP_AUTHENTICATION_TOKEN not set"
    glog.Errorln(errorMsg)
    return errors.New(errorMsg)
  }

  // Initializing appclient.
  appClient = CohesityAppSdk.NewAppSdkClient(
    appAuthenticationToken, apiEndpointIp, apiEndpointPort)
  return nil
}

func RunCohesityUnmount() error {

  // Issuing unmount command using cohesityAppSdk.
  if err := appClient.Mount().DeleteUnmount(FLAGS_unmountDir); err != nil {
    errMsg := "Error: " + err.Error()
    glog.Errorln(errMsg)
    return err
  }
  return nil

}
