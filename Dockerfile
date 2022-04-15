FROM golang:1.17.9-alpine3.15 AS build

WORKDIR /go/src/hvm

COPY . .

RUN go install /go/src/hvm/cmd/hvc

FROM alpine:3.15

WORKDIR /usr/bin

COPY --from=build /go/bin/hvc .

ENTRYPOINT [ "hvc" ]

CMD [ "help" ]
