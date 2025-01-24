FROM golang:1.23-alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . /app
ARG VERSION
ENV REVISION=$VERSION
RUN go build -ldflags="-s -w -X main.Version=$REVISION" -o /app/tmp/hubro

FROM debian:bookworm-slim AS assetsbuilder
RUN apt-get update && apt-get install -y wget
WORKDIR /app
COPY --from=base /app/view /app/view
# Helpers contains tailwind classes
COPY --from=base /app/helpers /app/helpers
RUN mkdir -p /app/scripts
COPY scripts/setup.sh /app/scripts/setup.sh
RUN chmod +x /app/scripts/setup.sh
RUN scripts/setup.sh
RUN bin/tailwindcss -i view/assets/css/app.css -m -o minified_app.css
RUN bin/esbuild view/assets/js/app.js --minify --target=es2017 --bundle --outfile=minified_app.js

FROM alpine:3.21 AS prod
RUN addgroup -S hubro && adduser -S hubro -G hubro
USER hubro
WORKDIR /app
COPY --from=base /app/tmp/hubro /app/hubro
COPY view /app/view
COPY --from=assetsbuilder /app/minified_app.css /app/view/static/app.css
COPY --from=assetsbuilder /app/minified_app.js /app/view/static/app.js
CMD ["/app/hubro"]
