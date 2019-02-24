// Copyright 2019 Cohesity Inc.


package main

import (
  "flag"
  "github.com/cohesity/cohesity-app-spec/viewbrowser/server"
)

func main() {
  flag.Parse()
  rs := viewbrowserserver.NewFileBrowserServer()
  rs.Start()
}
