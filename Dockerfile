FROM golang:1.18-alpine as build

RUN apk add -u git

WORKDIR /app
COPY . .
RUN go build -o /bundler

FROM golang:1.18-alpine
COPY --from=build /bundler /app/

LABEL "name"="Dependabot Bundler for Go"
LABEL "maintainer"="Gergely Brautigam <gergely@gergelybrautigam.com>"
LABEL "version"="0.0.2"

LABEL "com.github.actions.name"="Dependabot Bundler for Go"
LABEL "com.github.actions.description"="Bundles dependabot PRs into a single PR for Go based projects."
LABEL "com.github.actions.icon"="package"
LABEL "com.github.actions.color"="purple"

WORKDIR /app/
ENTRYPOINT [ "/app/bundler" ]
