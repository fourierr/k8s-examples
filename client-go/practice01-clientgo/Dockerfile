FROM golang:1.17 as builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -o svc-cm-controller main.go

FROM alpine:3.15.3

WORKDIR /app

COPY --from=builder /app/svc-cm-controller .

CMD ["./svc-cm-controller"]

# docker build -t fourier/svc-cm-controller:v1.0.0 .