FROM golang:1.18.2-buster as builder

RUN mkdir build
COPY ./src /build
WORKDIR /build

RUN go mod download

# Run tests
RUN go test

# Build and version the binary
RUN go build -ldflags "-X main.version=`git tag --sort=-version:refname | head -n 1`" -o /app

FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=builder /app /app

USER nonroot:nonroot

EXPOSE 8000

ENTRYPOINT [ "/app" ]