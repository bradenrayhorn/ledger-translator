FROM golang:1.16.4

WORKDIR /go/src/app
COPY . /go/src/app

RUN go get -v -t -d ./...
RUN ls -lah

CMD go test -v -coverprofile=./reports/coverage.txt -covermode=atomic -coverpkg=./... controller_test.go
