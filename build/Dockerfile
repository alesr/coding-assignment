FROM alpine:latest

RUN apk add --update go

WORKDIR /app

COPY go-alessandro-resta .

EXPOSE 8080

ENTRYPOINT ["./go-alessandro-resta"]
