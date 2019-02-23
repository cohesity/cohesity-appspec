FROM centos:centos7

WORKDIR /opt/insight/bin

ADD viewbrowser_server /opt/insight/bin/
ADD athena_mount /opt/insight/bin/
ADD athena_umount /opt/insight/bin/
ADD wrapper.sh /opt/insight/bin/

CMD ["/bin/bash", "/opt/insight/bin/wrapper.sh", "-stderrthreshold=INFO"]
