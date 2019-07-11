FROM golang:1.11-alpine AS build-env

RUN mkdir -p /app/bin

RUN apk add --no-cache postgresql git make musl-dev gcc

ADD . /go/src/github.com/ghostec/Will.IAM
RUN cd /go/src/github.com/ghostec/Will.IAM && \
  make setup-project && \
  make build && \
  mv bin/Will.IAM /app/Will.IAM && \
  mv config /app/config && \
  mv assets /app/assets && \
  mv Makefile /app/Makefile

FROM alpine:3.8

RUN apk add --no-cache ca-certificates
  
WORKDIR /app

COPY --from=build-env /app/Will.IAM /app
COPY --from=build-env /app/config /app/config
COPY --from=build-env /app/assets /app/assets
COPY --from=build-env /app/Makefile /app

EXPOSE 4040

CMD /app/Will.IAM start-api
