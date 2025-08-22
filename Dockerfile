FROM golang:1.24.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/fetch-duck -ldflags="-w -s" .

FROM gcr.io/distroless/static-debian11

COPY --from=builder /app/fetch-duck .

COPY --from=builder /app/static ./static 

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT [ "app/fetch-duck" ]