FROM golang:1.21-alpine

RUN apk add --no-cache make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make deps
RUN make

EXPOSE 8889
CMD ["./bin/commune"]

