FROM golang:1.16.4 as build

RUN mkdir /app
COPY . /app
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build .

FROM alpine:latest
COPY --from=build /app/ledger-translator /app/

EXPOSE 8080

ENTRYPOINT ["/app/ledger-translator"]
