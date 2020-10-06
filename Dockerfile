#build stage
FROM golang:1.14 AS build-env

LABEL maintainer="Jonathan Morais <jonathan.m.lucena@gmail.com>"

WORKDIR /go/src/app/

#copy to workdir path
COPY . .

ENV CGO_ENABLED=0

#build the go app
RUN go build

# final stage
FROM alpine
RUN apk --no-cache add ca-certificates
WORKDIR /app/

#copy the compilate binary for workdir
COPY --from=build-env /go/src/app .
ENTRYPOINT ["./duf"]