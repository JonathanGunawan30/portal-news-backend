FROM golang:1.25 AS builder

ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-w -s" -o main .

FROM gcr.io/distroless/static-debian11

COPY --from=builder /app/main /app/main

COPY ./docs /app/docs

USER nonroot:nonroot

WORKDIR /app

EXPOSE 8000

CMD ["/app/main"]