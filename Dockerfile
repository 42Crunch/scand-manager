FROM golang:1.24.4-alpine3.22 AS builder
WORKDIR /app
COPY go.mod go.sum *.go ./
RUN go mod download -x
RUN CGO_ENABLED=0 GOOS=linux go build -o scand-manager .

FROM busybox
# Create a non-root user and group with UID/GID 1001
RUN adduser -D -u 1001 scanduser
COPY --from=builder /app/scand-manager /bin/
USER scanduser

EXPOSE 8090
ENTRYPOINT ["/bin/scand-manager"]
