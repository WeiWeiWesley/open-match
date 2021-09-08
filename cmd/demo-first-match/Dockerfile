FROM golang:1.14.8-alpine3.11 as builder

WORKDIR /go/src/demo-first-match
COPY ./demo-first-match ./static /go/src/demo-first-match

FROM gcr.io/distroless/static:nonroot
WORKDIR /app/
COPY --from=builder --chown=nonroot /go/src/demo-first-match /app/

ENTRYPOINT ["/app/demo-first-match"]