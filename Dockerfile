FROM alpine:3.8
RUN apk update && apk add ca-certificates
RUN mkdir /app
WORKDIR /app
ADD bizbot ./
CMD ["./bizbot", "daemon"]

