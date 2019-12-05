FROM ubuntu:latest
RUN apt-get update
RUN apt-get upgrade -y
RUN apt install -y ebtables
RUN apt install -y net-tools
COPY networkmonitor /usr/bin/networkmonitor
CMD ["/usr/bin/networkmonitor"]