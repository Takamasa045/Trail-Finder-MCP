FROM golang:1.23-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /trail-finder-mcp ./cmd/trail-finder-mcp

FROM gcr.io/distroless/static-debian12
COPY --from=builder /trail-finder-mcp /trail-finder-mcp
USER nonroot:nonroot
ENTRYPOINT ["/trail-finder-mcp"]
