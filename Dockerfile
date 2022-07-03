FROM golang:1.18-buster as builder

# go deps
ENV GO111MODULE=on \
    GOOS=linux \
    GOARCH=amd64

COPY go.mod go.sum build/

WORKDIR build

RUN go mod download

COPY . .

WORKDIR cmd/

RUN CGO_ENABLED=0 go build -o /main



FROM alpine:latest

# time zone
RUN apk add --no-cache tzdata
ENV TZ=Europe/Moscow

WORKDIR /app/

COPY --from=builder /main .

#RUN mkdir configs
#WORKDIR configs

COPY --from=builder /go/build/configs configs/




CMD ["/app/main"]
