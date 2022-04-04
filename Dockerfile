FROM alpine

WORKDIR /app

ADD capper /app/
ENTRYPOINT /app/capper
EXPOSE 8443