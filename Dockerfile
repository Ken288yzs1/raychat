FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o raychat .

FROM golang:1.21-alpine AS runner

WORKDIR /app

COPY --from=builder /app/raychat .

EXPOSE 7860

CMD ["/app/raychat"]
