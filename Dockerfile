FROM alpine:3.7
RUN apk add --no-cache ca-certificates 

WORKDIR /app

ADD config.yml .
ADD dist/gohazel .

RUN mkdir /assets

CMD ["./gohazel"]