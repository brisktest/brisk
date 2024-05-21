## We specify the base image we need for our
## go application
FROM bitnami/minideb:bullseye

#FROM nestybox/ubuntu-bionic-systemd
## We create an /app directory within our
## image that will hold our application source
## files
#RUN sudo useradd -ms /bin/bash brisk

ENV TINI_VERSION v0.18.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

ARG GROUP_ID=3434
ARG USER_ID=3434

RUN [ $(getent group ${GROUP_ID} ) ] || groupadd brisk -g ${GROUP_ID} 

RUN adduser -uid ${USER_ID} -gid ${GROUP_ID} brisk

ADD ./docker/scripts/rrsync /usr/local/bin/rrsync
RUN chmod +x /usr/local/bin/rrsync
RUN ln -s /usr/local/bin/rrsync /usr/bin/rrsync
RUN apt update && apt -y install openssh-server
RUN  mkdir -p /run/sshd
RUN  apt update &&  apt-get -y install rsync
RUN  sed -i 's/UsePAM yes/UsePAM no/g' /etc/ssh/sshd_config

RUN  sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/g' /etc/ssh/sshd_config
RUN  sed -i 's|#HostKey /etc/ssh/ssh_host_|HostKey /home/brisk/.sshd/ssh_host_|g' /etc/ssh/sshd_config
#RUN  sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/g' /etc/ssh/sshd_config
# RUN mkdir /etc/rsyslog.d
# RUN touch /etc/rsyslog.d/sshd.conf
# RUN echo 'if $programname == "sshd" then /home/brisk/sshd-log' >> /etc/rsyslog.d/sshd.conf
RUN  echo "Port 2222" >> /etc/ssh/sshd_config
RUN echo "PidFile /tmp/sshd.pid" >> /etc/ssh/sshd_config
RUN echo "LogLevel DEBUG3" >> /etc/ssh/sshd_config
#RUN echo "UsePrivilegeSeparation no" >> /etc/ssh/sshd_config
RUN chown -R brisk /var
RUN mkdir -p /home/brisk/.ssh
RUN mkdir -p /tmp/remote_dir
RUN chown -R brisk /tmp/
RUN systemctl enable ssh 


RUN mkdir /home/brisk/.sshd
RUN cp /etc/ssh/ssh_host* /home/brisk/.sshd

WORKDIR /home/brisk/
ADD docker/docker-entrypoint-ssh.sh ./
RUN chmod +x docker-entrypoint-ssh.sh
ENTRYPOINT [ "/tini","-v", "--" ,"/home/brisk/docker-entrypoint-ssh.sh"]
RUN  chown -R  brisk /home/brisk

## Our start command which kicks off
## our newly created binary executable
#USER brisk
#CMD  /usr/sbin/sshd -f /etc/ssh/sshd_config && /app/super.bin

USER brisk
CMD ["/usr/sbin/sshd -D -e"]
