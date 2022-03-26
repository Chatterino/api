FROM golang:1.18 AS build
ADD . /src
RUN cd /src/cmd/api && GOOS=linux GOARCH=amd64 go build -tags netgo -ldflags '-extldflags "-static"'

FROM alpine:latest
WORKDIR /app
COPY --from=build /src/cmd/api/api /app/
RUN apk add --no-cache ca-certificates
CMD ["./api"]
