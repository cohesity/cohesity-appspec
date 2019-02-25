// Copyright 2019 Cohesity Inc.
//
// Utility to parse and validate developer's appspec.
// Build the appspecvalidator_exec binary and pass the absolute appspec path
// as commandline  argument. Eg. ./appspecvalidator_exec appspecpath

package main

import (
  "fmt"
   "os"

  "cohesity/athena/cohesity-app-spec/appspecvalidator/appspec_validator"
)

func main() {
  // Path of the app spec.
  appSpecPath := os.Args[1]
  err := appspecvalidator.ParseAndValidateAppSpec(appSpecPath)
  if err != nil {
    fmt.Println(err)
  } else {
    fmt.Println("Valid App Spec.")
  }
}
