// Copyright 2019 Cohesity Inc.
//

package main

import (
  "github.com/cohesity/cohesity-app-spec/viewbrowser/server"
)

func main() {
  rs := viewbrowserserver.NewFileBrowserServer()
  rs.Start()
}
