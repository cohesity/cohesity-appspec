Containerize Sample App
=======================

## Dockerfile
```bash
FROM centos:centos7

WORKDIR /opt/viewbrowser/bin
ADD view_browser_exec /opt/viewbrowser/bin/
ADD wrapper.sh /opt/viewbrowser/bin/

CMD ["/bin/bash", "/opt/viewbrowser/bin/wrapper.sh", "-stderrthreshold=INFO"]
```

## Building the go binary
```bash
go build -o deployment/view_browser_exec view_browser_exec.go
``` 

## Building an Image
Build a Docker image
```bash
docker build -t view-browser .
```

Create a file
```json
touch view-browser:latest
```

Save the Docker image
```bash
docker save view-browser -o view-browser:latest
```

Verify:
```bash
docker images
```
## AppSpec 

### App Structure
```yaml
# Copyright 2019 Cohesity Inc.

apiVersion: v1
kind: Service
metadata:
  name: view-browser-rest
  labels:
    app: view-browser
spec:
  type: NodePort
  selector:
    app: view-browser
  ports:
  - port: 8080
    protocol: TCP
    name: rest
    cohesityTag: ui
---
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: view-browser
  labels:
    app: view-browser
spec:
  replicas:
    fixed: 1
  selector:
    matchLabels:
      app: view-browser
  template:
    metadata:
      labels:
        app: view-browser
    spec:
      containers:
      - name: view-browser
        image: view-browser:latest
        resources:
          requests:
            cpu: 500m
            memory: 100Mi
```
### App Metadata
```json
{
 "id": 1,
 "name" : "View-browser",
 "version" : 1,
 "dev_version": 1.0,
 "description" : "Viewbrowser: Browse the views and files inside a view",
 "access_requirements" : {
    "read_access" : true,
    "read_write_access" : false,
    "management_access" : true
 }
}
```

[Validating AppSpec](https://github.com/cohesity/cohesity-appspec/blob/master/tools/appspecvalidator/README.md)