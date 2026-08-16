# Build the UI
FROM node:24-slim AS nodework
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
ENV CI=TRUE
RUN corepack enable

WORKDIR /app
COPY ui/ .
# CI=TRUE makes pnpm treat ignored build scripts (esbuild) as fatal, so the
# install has to opt into running them.
RUN pnpm install --config.dangerouslyAllowAllBuilds=true
RUN pnpm run build

# Build the Go binary
FROM golang:1.25 AS gowork
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
COPY --from=nodework /app/build ui/build
RUN go build -o /app/main .


# Run the binary in a minimal container
FROM ubuntu:22.04
# The base image ships no CA bundle, so any outbound HTTPS the app makes —
# purging Cloudflare's cache, for one — fails with "certificate signed by
# unknown authority".
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=gowork /app/main .
CMD [ "./main" ]
