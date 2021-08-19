# Copyright 2019 Cohesity Inc.
#
# This is a simple wrapper around view_browser_exec that restarts it if it
# crashes.

#! /bin/bash

while true; do
  echo "Starting viewbrowser server ..."
  /opt/viewbrowser/bin/view_browser_exec $@
  if [ "$?" == "0" ]; then
    echo "Done"
    break
  fi
  echo "Sleeping ..."
  sleep 5
done
