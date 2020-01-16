# build stage
FROM golang:alpine AS build-env
# git is needed for fetching dependencies
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/app/
COPY cmd/paddleball .
RUN go get -d -v
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /go/bin/app

# final stage
FROM scratch
COPY --from=build-env /go/bin/app /app
ENTRYPOINT ["/app"]
