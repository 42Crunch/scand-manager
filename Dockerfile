FROM golang:1.15.5-alpine3.12 as builder
WORKDIR /app
COPY go.mod go.sum *.go ./
RUN go mod download -x
RUN CGO_ENABLED=0 GOOS=linux go build -o scand-manager .

FROM busybox
COPY --from=builder /app/scand-manager /bin/
EXPOSE 8090
ENTRYPOINT ["/bin/scand-manager"]
