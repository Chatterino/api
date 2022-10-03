FROM ubuntu:22.04 AS build
ENV GOVER=1.19.2
ADD . /src
RUN apt update && apt -y install libvips-dev wget build-essential
RUN wget -qO- https://go.dev/dl/go$GOVER.linux-amd64.tar.gz | tar -C /src -xzf -
RUN cd /src/cmd/api && /src/go/bin/go build

FROM ubuntu:22.04
WORKDIR /app
COPY --from=build /src/cmd/api/api /app/
RUN apt update && apt install -y ca-certificates libvips && apt clean
CMD ["./api"]
