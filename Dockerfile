FROM alpine:latest

RUN apk update && \
    apk add --no-cache git

COPY forjj /bin/forjj

ENTRYPOINT ["/bin/forjj"]
CMD ["--help"]

