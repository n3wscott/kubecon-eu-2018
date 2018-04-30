FROM alpine:3.4

RUN apk add --no-cache ca-certificates

EXPOSE 8080

ADD ./out/twitter /bin/twitter

CMD ["twitter"]
