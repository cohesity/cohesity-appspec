// Copyright 2019 Cohesity Inc.
//

package main

import (
  "github.com/cohesity/cohesity-appspec/sampleapp/viewbrowser/server"
)

func main() {
  rs := viewbrowser.NewViewBrowserServer()
  rs.Start()
}
