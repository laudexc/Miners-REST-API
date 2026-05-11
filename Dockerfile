FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/miners-rest-api ./cmd/app

FROM alpine:3.22

WORKDIR /app
COPY --from=build /bin/miners-rest-api /app/miners-rest-api

EXPOSE 8080
CMD ["/app/miners-rest-api"]
