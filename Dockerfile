ARG GO_VERSION=1.21
FROM golang:${GO_VERSION} as builder

# Create the app directory and set it as the working directory
RUN mkdir -p /app
WORKDIR /app

# Copy the source code into the container
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build the Go program
RUN make build

FROM debian:bookworm-slim

WORKDIR /norns

COPY --from=builder /app/bin/norns .

CMD ["./norns"]
