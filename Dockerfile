FROM golang:1.22 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o /trail-finder-mcp ./cmd/trail-finder-mcp

FROM gcr.io/distroless/base-debian12
ENV PORT=8080
COPY --from=builder /trail-finder-mcp /trail-finder-mcp
EXPOSE 8080
ENTRYPOINT ["/trail-finder-mcp"]
