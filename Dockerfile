FROM golang:1.22-alpine AS build

WORKDIR /

COPY [".", "."]
RUN cd cmd/nethax && go mod download

RUN cd cmd/nethax && CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s -extldflags "-static"' -a -o nethax .

FROM scratch

COPY --from=build /cmd/nethax/nethax /nethax

ENTRYPOINT ["/nethax"]
