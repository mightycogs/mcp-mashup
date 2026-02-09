# syntax=docker/dockerfile:1.4

FROM alpine:3.19
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY bin/mcp-mashup-${TARGETOS}-${TARGETARCH} /usr/local/bin/mcp-mashup
RUN chmod +x /usr/local/bin/mcp-mashup

USER 1000:1000

ENTRYPOINT ["/usr/local/bin/mcp-mashup"]
