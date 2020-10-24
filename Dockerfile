FROM alpine:3.7
RUN apk add --no-cache ca-certificates 

WORKDIR /app

ADD config.yml .
ADD gohazel .

RUN mkdir /assets

CMD ["./gohazel"]