FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /pmtop ./cmd/pmtop

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /pmtop /pmtop
ENTRYPOINT ["/pmtop"]
