FROM golang

CMD go get github.com/bdelliott/withings-fatsecret-sync
#WORKDIR /go/src/github.com/bdelliott/withings-fatsecret-sync
#CMD go build

#ENTRYPOINT /go/bin/withings-fact-secret
ENTRYPOINT /bin/bash
