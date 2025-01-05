FROM golang:1.23-alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . /app
ARG VERSION
ENV REVISION=$VERSION
RUN scripts/setup.sh
RUN bin/tailwindcss -c view/tailwind.config.js -i view/assets/css/app.css -m -o minified_app.css
RUN bin/esbuild view/assets/js/app.js --minify --target=es2017 --bundle --outfile=minified_app.js
RUN go build -ldflags="-s -w -X main.Version=$REVISION" -o /app/tmp/hubro

FROM alpine:3.21 AS prod
RUN addgroup -S hubro && adduser -S hubro -G hubro
USER hubro
WORKDIR /app
COPY --from=base /app/tmp/hubro /app/hubro
COPY view /app/view
COPY --from=base /app/minified_app.css /app/view/static/app.css
COPY --from=base /app/minified_app.js /app/view/static/app.js
CMD ["/app/hubro"]
