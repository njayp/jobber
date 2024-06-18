FROM golang:alpine AS build
RUN apk update && \
    apk add --no-cache make
WORKDIR /app

# cache libraries
COPY go.mod go.sum ./
RUN go mod download

# build app
COPY . .
RUN go build -o output/bin/jobber cmd/jobber/main/main_linux.go

FROM alpine
COPY --from=build /app/output/bin /bin
EXPOSE 9090
ENTRYPOINT [ "jobber", "serve" ]

