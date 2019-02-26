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

[Validating AppSpec](tools/appspecvalidator/README.md)

## Create Tarball

Create a tarball consisting of App.json, Docker Image and AppSpec.
```bash
tar cvzf view-browser.tar.gz view-browser:latest development/app.json 
viewbrowser_spec.yaml 
``` 

## Validation by Cohesity
Send this tar.gz package to developer@cohesity.com to get this package validated. 
