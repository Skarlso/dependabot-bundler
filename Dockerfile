FROM golang:1.20-alpine as build
RUN apk add -u git
WORKDIR /app
COPY . .
RUN go build -o /bundler

FROM alpine
RUN apk add -u ca-certificates
COPY --from=build /bundler /app/

WORKDIR /app/
ENTRYPOINT [ "/app/bundler" ]