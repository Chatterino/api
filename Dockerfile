FROM golang:1-alpine AS build
ADD . /src
RUN apk add --no-cache build-base pkgconfig vips-dev
RUN cd /src/cmd/api && go build

FROM alpine:latest
WORKDIR /app
COPY --from=build --link /src/cmd/api/api /app/
RUN apk add --no-cache ca-certificates vips vips-poppler font-noto
CMD ["./api"]
