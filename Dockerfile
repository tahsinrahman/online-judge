FROM golang

# Fetch dependencies
RUN go get github.com/tools/godep

# Add project directory to Docker image.
ADD . /go/src/github.com/tahsinrahman/online-judge

ENV USER tahsin
ENV HTTP_ADDR :8888
ENV HTTP_DRAIN_INTERVAL 1s
ENV COOKIE_SECRET Vdu0fUTubg8sP68i

# Replace this with actual PostgreSQL DSN.
ENV DSN $GO_BOOTSTRAP_MYSQL_DSN

WORKDIR /go/src/github.com/tahsinrahman/online-judge

RUN godep go build

EXPOSE 8888
CMD ./online-judge
