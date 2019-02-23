// Copyright 2019 Cohesity Inc.

package main

import (
  "flag"
  "fmt"
  "os"

  "github.com/golang/glog"
  "github.com/cohesity/cohesity-appspec/tools/cohesity_unmount/cohesityumount"
)

func main() {
  err := cohesityumount.Init()
  if err != nil {
    glog.Errorln("Unmount Request failed.Error in intialization: "+ fmt.Sprint(err))
    os.Exit(1)
  }
  flag.Parse()
  err = cohesityumount.RunCohesityUnmount()
  if err != nil {
    glog.Errorln("Unmount request failed.Error: " + fmt.Sprint(err))
    os.Exit(1)
  }
  glog.Infoln("Unmount request successful.")
  os.Exit(0)
}
