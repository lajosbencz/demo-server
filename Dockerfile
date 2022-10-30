FROM golang:1.19-alpine as builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o demo-server .

FROM scratch
COPY --from=builder /app/demo-server /usr/bin/
EXPOSE 8080
VOLUME [ "/etc/demo-server/" ]
ENTRYPOINT [ "/usr/bin/demo-server", "-h", "0.0.0.0", "-p", "8080", "-f", "/etc/demo-server/persist.json" ]
