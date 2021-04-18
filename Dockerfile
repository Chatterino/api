FROM golang:latest AS build
ADD . /src
RUN cd /src/cmd/api && GOOS=linux GOARCH=amd64 go build -ldflags '-extldflags "-static"'

FROM alpine:latest
WORKDIR /app
COPY --from=build /src/cmd/api/api /app/
RUN apk add --no-cache ca-certificates
CMD ["./api"]
