FROM nginx:latest
MAINTAINER Tahsin Rahman <tahsinrahman@protonmail.com>

ADD . /opt/online-judge
WORKDIR /opt/online-judge

EXPOSE 4000

ENTRYPOINT ["./online-judge"]
