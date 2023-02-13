FROM golang:1.20-alpine as build
RUN apk add -u git
WORKDIR /app
COPY . .
RUN go build -o /bundler

FROM alpine
RUN apk add -u ca-certificates
COPY --from=build /bundler /app/

LABEL "name"="Dependabot Bundler"
LABEL "maintainer"="Gergely Brautigam <gergely@gergelybrautigam.com>"
LABEL "version"="0.0.1"

LABEL "com.github.actions.name"="Dependabot Bundler"
LABEL "com.github.actions.description"="Bundle Dependabot PRs into one."
LABEL "com.github.actions.icon"="package"
LABEL "com.github.actions.color"="purple"

WORKDIR /app/
ENTRYPOINT [ "/app/bundler" ]