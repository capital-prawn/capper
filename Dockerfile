FROM alpine

WORKDIR /app

ADD capper /app/
ENTRYPOINT /capper
EXPOSE 8443