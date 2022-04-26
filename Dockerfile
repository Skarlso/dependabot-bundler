FROM golang:1.18-alpine as build
WORKDIR /app
COPY . .
RUN go build -o /bundler

FROM alpine
RUN apk add -u ca-certificates
COPY --from=build /bundler /app/

LABEL "name"="Dependabot Bundler for Go"
LABEL "maintainer"="Gergely Brautigam <gergely@gergelybrautigam.com>"
LABEL "version"="0.0.1"

LABEL "com.github.actions.name"="Dependabot Bundler for Go"
LABEL "com.github.actions.description"="Bundles dependabot PRs into a single PR for Go based projects."
LABEL "com.github.actions.icon"="package"
LABEL "com.github.actions.color"="purple"

WORKDIR /app/
ENTRYPOINT [ "/app/bundler" ]
