// Copyright 2019 Cohesity Inc.
//
// Author: Abhilash Velalam

package main

import (
  "github.com/cohesity-apps-dev-tools/viewbrowser/server"
  "flag"
)

func main() {
  flag.Parse()
  rs := viewbrowserserver.NewFileBrowserServer()
  rs.Start()
}
