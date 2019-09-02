FROM golang:1.12-alpine AS build-env

#RUN apk add --no-cache postgresql git make musl-dev gcc
RUN apk add --no-cache make git

WORKDIR /Will.IAM

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .
RUN make build

FROM alpine:3.8

RUN apk add --no-cache ca-certificates
  
WORKDIR /app

COPY --from=build-env /Will.IAM/bin/Will.IAM /app
COPY --from=build-env /Will.IAM/config /app/config
COPY --from=build-env /Will.IAM/assets /app/assets

EXPOSE 4040

CMD /app/Will.IAM start-api
