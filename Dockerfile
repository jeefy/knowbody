FROM golang:1.14

WORKDIR /go/src/github.com/jeefy/knowbody/

COPY . .
RUN make check
RUN make build

FROM gcr.io/distroless/base

COPY --from=0 /go/src/github.com/jeefy/knowbody/out/knowbody .
ENTRYPOINT [ "/knowbody" ]