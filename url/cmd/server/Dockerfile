FROM golang:alpine AS build

LABEL MAINTAINER="Farzad Shami"

RUN apk update && \
    apk add curl \
            git \
            bash \
            make \
            ca-certificates && \
    rm -rf /var/cache/apk/*

WORKDIR /app

# copy module files first so that they don't need to be downloaded again if no change
COPY go.* ./
RUN go mod download
RUN go mod verify

# copy source files and build the binary
COPY .. .
RUN make build


FROM alpine:latest

RUN apk --no-cache add ca-certificates bash
RUN mkdir -p /var/log/app

WORKDIR /app/

COPY --from=build /app/server .
COPY --from=build /app/cmd/server/entrypoint.sh .
COPY --from=build /app/config/* ./config/

# copy swagger docs
#COPY --from=build /app/docs/docs.go ./docs/docs.go
#COPY --from=build /app/docs/swagger.json ./docs/swagger.json
#COPY --from=build /app/docs/swagger.yaml ./docs/swagger.yaml


RUN ls -la
RUN ["chmod", "+x", "./entrypoint.sh"]

ENTRYPOINT ["./entrypoint.sh"]
