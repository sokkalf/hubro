# Hubro

## Description

Hubro is not a static site generator, and not a traditional database backed blog engine either. It reads markdown files from a `blog` directory and renders them to HTML using Go templates. It also reads markdown files from a `pages` directory and renders them to HTML. The `pages` directory is for static pages, like an about page or a contact page. Everything is read into memory at startup, and updated if the files change.

## Up and running

### Install TailwindCSS and ESBuild

`scripts/setup.sh` will install TailwindCSS and ESBuild.

Install Air:

```
go install github.com/air-verse/air@latest
```

### Run with Air

```
air
```

## Generate random data for testing

```
go run cli/main.go [number of posts]
```

This will put random markdown files in the `blog` directory.

### Build with Docker

```
docker build -t hubro .
docker run -e HUBRO_TITLE=ugle-z.no -e HUBRO_DESCRIPTION="Random ramblings" -v ./blog:/app/blog -v ./pages:/app/pages -p 8888:8080 -it hubro
```
