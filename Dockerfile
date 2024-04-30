MAINTAINER sleeps17

FROM golang:1.22-alpine AS builder

WORKDIR /go/src/app

RUN apk --no-cache add git bash make gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /go/bin/linker ./cmd/linker

FROM alpine:latest AS runner

COPY --from=builder /go/bin/linker ./
COPY config/config.yaml /config/config.yaml

ENV CONFIG_PATH=/config/config.yaml

EXPOSE 4404

CMD ["./linker"]