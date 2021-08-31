# Change the directory to the required location
cd ~/temp/cohesity-appspec/sampleapp/viewbrowser  
echo "Switched to the View Browser Directory"

go build -o view_browser_exec .
echo "Built the Go binary"

cp view_browser_exec deployment/view_browser_exec
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
