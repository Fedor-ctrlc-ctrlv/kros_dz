FROM golang:1.26.1-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.20
COPY --from=build /server /server
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/server"]
