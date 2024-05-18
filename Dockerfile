FROM golang:alpine as builder
LABEL maintainer="Dmytro Rudoi"

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

ENV ETHEREAL_EMAIL=elta.gibson66@ethereal.email
ENV ETHEREAL_PASSWORD=XhuJdMPYPPENEnXPsF

FROM golang:1.18-alpine

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]