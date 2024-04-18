FROM golang:1.22-alpine AS build

WORKDIR /

COPY [".", "."]
RUN ["go", "mod", "download"]

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s -extldflags "-static"' -a -o nethax .

FROM scratch

COPY --from=build /nethax /nethax

ENTRYPOINT ["/nethax"]
