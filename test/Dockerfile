FROM golang:1.16
ARG TESTS
ENV TESTS=${TESTS}
WORKDIR /go/src/test
COPY docker-entrypoint.sh *.go ./
ADD https://raw.githubusercontent.com/eficode/wait-for/master/wait-for ./wait-for.sh
ENTRYPOINT ./docker-entrypoint.sh "$TESTS"