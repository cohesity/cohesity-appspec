# Change the directory to the required location
cd ~/workspace/cohesity-appsec-new/cohesity-appspec/sampleapp/viewbrowser  
echo "Switched to the View Browser Directory"

rm go.mod go.sum viewbrowser
echo "Removed the existing binaries and modules"

go mod init github.com/cohesity/cohesity-appspec/sampleapp/viewbrowser
echo "Created Go modules"

go get -v 
echo "Downloaded all dependencies"

go build .
echo "Built the Go binary"

cp viewbrowser deployment/view_browser_exec
echo "Copied the binary to the view browser folder"

cd deployment
echo "Swtiched to the deployment folder"

docker image rm view-browser:latest
echo "Removed existing docker images"

docker build -t view-browser .
echo "Docker new image build complete"

touch view-browser:latest
docker save view-browser -o view-browser:latest
echo "Docker image saved"

docker images
echo "Finished"
