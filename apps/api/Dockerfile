FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /voxtrace-api .
FROM alpine:3.21
RUN adduser -D -u 10001 voxtrace
USER voxtrace
COPY --from=build /voxtrace-api /voxtrace-api
ENTRYPOINT ["/voxtrace-api"]
