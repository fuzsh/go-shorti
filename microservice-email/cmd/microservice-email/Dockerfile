FROM golang:alpine as build

LABEL Author="Farzad Sahmi <farzadshami845@gmail.com>"

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
COPY . .
RUN make build

FROM alpine:latest

RUN apk --no-cache add ca-certificates bash
RUN mkdir -p /var/log/app

WORKDIR /app/

COPY --from=build /app/microservice-email .
COPY --from=build /app/cmd/microservice-email/entrypoint.sh .
COPY --from=build /app/config/* ./config/

# Configuration
RUN ls -la
RUN ["chmod", "+x", "./entrypoint.sh"]

ENTRYPOINT ["./entrypoint.sh"]

EXPOSE 6081 6082
