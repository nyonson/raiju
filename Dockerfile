FROM golang:alpine3.17 AS build-env

RUN apk add --update --no-cache git

COPY . /src/ 

RUN cd /src/ \
    && rm -rf .git \
    && go install \
    && go install ./cmd/raiju \
    && go test \
    && chmod a+x $GOPATH/bin/raiju

FROM alpine:3.17

ENV GOPATH=/go

LABEL org.opencontainers.image.source https://github.com/nyonson/raiju

COPY --from=build-env $GOPATH/bin/raiju /raiju

VOLUME [ "/root/.config/raiju" ]

WORKDIR /

ENTRYPOINT ["/raiju"]

CMD [ "/raiju", "-h" ]