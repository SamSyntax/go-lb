FROM golang:1.21.3-bookworm AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download && go mod verify

COPY *.go .

RUN go build -o /myapp *.go

FROM gcr.io/distroless/base-debian12

COPY --from=build /myapp /myapp

ENTRYPOINT ["/myapp"]
