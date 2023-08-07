FROM golang:1.20-alpine AS builder
WORKDIR /src/
COPY . /src/
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags='-extldflags=-static' -o /main .

FROM gcr.io/distroless/static
COPY --from=builder /main /bin/main
EXPOSE 80

ENTRYPOINT [ "/main" ]