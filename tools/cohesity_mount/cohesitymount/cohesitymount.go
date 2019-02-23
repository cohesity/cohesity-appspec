// Copyright 2019 Cohesity Inc.

package cohesitymount

import (
  "errors"
  "flag"
  "fmt"
  "os"

  "github.com/cohesity/app-sdk-go/appsdk"
  "github.com/cohesity/app-sdk-go/models"
  "github.com/golang/glog"
  "github.com/cohesity/cohesity-appspec/tools/cohesity_mount/utils"
)

var (
  // FLAGS_view specifies the view that is to be mounted.
  FLAGS_view string

  // FLAGS_options specifies the mount options.
  FLAGS_options string

  // FLAGS_mountDir specifies the directory on which the view is
  // to be mounted.
  FLAGS_mountDir string

  // FLAGS_protocol specifies the mount protocol(nfs/smb).
  FLAGS_protocol string

  // FLAGS_namespace specifies the namespace of the view that
  // is to be mounted.
  FLAGS_namespace string

  // Flags_username specifies the username for validation of
  // smb mount.
  FLAGS_username string

  // FLAGS_password specifies the password for validation of
  // smb mount.
  FLAGS_password string

  // The following are read from environment variable during init.

  // IP address of the host on which the container is running.
  hostIp                 string

  // Ip address of the cohesity app server
  apiEndpointIp          string
  
  // Port of the cohesity app server
  apiEndpointPort        string

  // Authentication token the app uses to authenticate with the service.
  appAuthenticationToken string
  
  // appClient to make cohesity appserver api calls.
  appClient              CohesityAppSdk.COHESITYAPPSDK
)

const (
  kSmbProtocol string = "smb"
  kNfsProtocol string = "nfs"
)

// function to initialize mount parameters.
func Init() error {

  // Initializing flag parameters.
  flag.StringVar(&FLAGS_view, "view", "", "View that is to be mounted.")
  flag.StringVar(&FLAGS_options, "options", "", "Mount options")
  flag.StringVar(&FLAGS_mountDir, "mountdir", "",
    "Directory on which the view is to be mounted.")
  flag.StringVar(&FLAGS_protocol, "protocol", "nfs", "mount protocol")
  flag.StringVar(&FLAGS_namespace, "namespace", "fs",
    "Namespace of the view that is to be mounted")
  flag.StringVar(&FLAGS_username, "username", "", "Username for smb mount.")
  flag.StringVar(&FLAGS_password, "password", "", "Password for smb mount.")

  // Read the environment variables.
  if hostIp = os.Getenv("HOST_IP"); len(hostIp) == 0 {
    errorMsg := "HOST_IP not set"
    glog.Errorln(errorMsg)
    return errors.New(errorMsg)
  }

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

  if appAuthenticationToken = os.Getenv("APP_AUTHENTICATION_TOKEN"); len(appAuthenticationToken) == 0 {
    errorMsg := "APP_AUTHENTICATION_TOKEN not set"
    glog.Errorln(errorMsg)
    return errors.New(errorMsg)
  }

  // Initializing appclient.
  appClient = CohesityAppSdk.NewAppSdkClient(
    appAuthenticationToken, apiEndpointIp, apiEndpointPort)

  return nil
}

// function to validate the user given mount parameters.
func validate() error {
  if FLAGS_view == "" {
    errorMsg := "view not specified."
    glog.Errorln(errorMsg)
    return errors.New(errorMsg)
  }

  if FLAGS_mountDir == "" {
    errorMsg := "Mount directory not specified."
    glog.Errorln(errorMsg)
    return errors.New(errorMsg)
  }

  if FLAGS_options != "options" {
    if FLAGS_protocol == "nfs" {
      utils.ValidateMountOptions(FLAGS_options, utils.NfsValidation)
    } else if FLAGS_protocol == "smb" {
      utils.ValidateMountOptions(FLAGS_options, utils.SmbValidation)
    } else {
      errorMsg := "Mount Protocol: " + FLAGS_protocol + " not supported."
      glog.Errorln(errorMsg)
      return errors.New(errorMsg)
    }
  }
  return nil
}

func RunCohesityMount() error {
  err := validate()
  if err != nil {
    glog.Errorf(fmt.Sprint(err))
    return err
  }
  var mountOptions models.MountOptions

  // Default Protocol
  if FLAGS_protocol == kSmbProtocol {
    // Setting the mount parameters.
    mountOptions = models.MountOptions{
      ViewName:      FLAGS_view,
      DirName:       FLAGS_mountDir,
      MountProtocol: models.MountProtocol_KSMB,
      MountOptions:  &FLAGS_options,
      UserName:      &FLAGS_username,
      Password:      &FLAGS_password,
      NamespaceName: &FLAGS_namespace,
    }
  } else {
    // Settings the mount parameters
    mountOptions = models.MountOptions{
      ViewName:      FLAGS_view,
      DirName:       FLAGS_mountDir,
      MountProtocol: models.MountProtocol_KNFS,
      MountOptions:  &FLAGS_options,
      NamespaceName: &FLAGS_namespace,
    }
  }

  // Making mount call using sdk.
  err = appClient.Mount().CreateMount(&mountOptions)

  if err != nil {
    glog.Errorf(fmt.Sprint(err))
    return err
  }
  return nil
}
