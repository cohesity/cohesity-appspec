// Copyright 2019 Cohesity Inc.

package main

import (
  "flag"
  "fmt"
  "os"

  "github.com/golang/glog"
  "github.com/cohesity/cohesity-appspec/tools/cohesity_mount/cohesitymount"
)

func main() {
  err := cohesitymount.Init()
  if err != nil {
    glog.Errorln("Error in intialization: " + fmt.Sprint(err))
    os.Exit(1)
  }
  flag.Parse()
  err = cohesitymount.RunCohesityMount()
  if err != nil {
    glog.Errorln("Mount Request failed.Error: " + fmt.Sprint(err))
    os.Exit(1)
  }
  glog.Infoln("Mount Request Successful.")
  os.Exit(0)
}
