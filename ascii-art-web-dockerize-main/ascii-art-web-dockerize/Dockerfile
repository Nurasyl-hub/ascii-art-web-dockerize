FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download || true

COPY . .

ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags "-s -w" -o /out/app ./main.go

FROM scratch

ARG VERSION=1.0.0
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

LABEL org.opencontainers.image.title="ascii-art-web" \
      org.opencontainers.image.description="Go web app that renders ASCII banners (stdlib only)" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.authors="Yryszhan Sansyzbay <dev@yryszlab.kz>"

USER 10001:10001
WORKDIR /app

COPY --from=builder /out/app       ./app
COPY --from=builder /src/templates ./templates
COPY --from=builder /src/styles    ./styles
COPY --from=builder /src/banners   ./banners

EXPOSE 8080
ENTRYPOINT ["./app"]
