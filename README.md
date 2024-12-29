# Hubro

## Up and running

### Install tailwind

`scripts/setup.sh` will install tailwind.

Install Air:

```
go install github.com/air-verse/air@latest
```

### Run with Air

```
air
```


### Build with Docker

```
docker build -t hubro .
docker run -e HUBRO_TITLE=ugle-z.no -e HUBRO_DESCRIPTION="Random ramblings" -v ./blog:/app/blog -v ./pages:/app/pages -p 8888:8080 -it hubro
```
