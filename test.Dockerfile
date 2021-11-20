FROM golang:1.17

WORKDIR /go/src/app
COPY . /go/src/app

RUN go get -v -t -d ./...

CMD go test -v -coverprofile=./reports/coverage.txt -covermode=atomic -coverpkg=./... ./...
