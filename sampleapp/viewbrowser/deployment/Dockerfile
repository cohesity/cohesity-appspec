FROM centos:centos7

WORKDIR /opt/viewbrowser/bin
ADD view_browser_exec /opt/viewbrowser/bin/
ADD wrapper.sh /opt/viewbrowser/bin/

CMD ["/bin/bash", "/opt/viewbrowser/bin/wrapper.sh", "-stderrthreshold=INFO"]
