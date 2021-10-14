## NOTE: This image uses goreleaser to build image
# if building manually please run go build ./cmd/yot first and then build

# Choose alpine as a base image to make this useful for CI, as many
# CI tools expect an interactive shell inside the container
FROM alpine:latest as production

COPY operator-builder /usr/bin/operator-builder
RUN chmod +x /usr/bin/operator-builder

WORKDIR /workdir

ENTRYPOINT ["/usr/bin/operator-builder"]
CMD ["--help"]