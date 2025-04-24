ARG TARGET_APPLICATION
ARG VERSION

FROM golang:1.23-alpine3.21 AS builder

ARG TARGET_APPLICATION
ENV TARGET_APPLICATION=$TARGET_APPLICATION

ARG VERSION
ENV VERSION=$VERSION

RUN if [ -z ${TARGET_APPLICATION} ]; then echo 'TARGET_APPLICATION required' && exit 1; fi
RUN if [ -z ${VERSION} ]; then echo 'VERSION required' && exit 1; fi

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Always output as app - the binary name is inconsequential within the context of the container
RUN go build -ldflags "-X 'github.com/moonmoon1919/go-api-reference/internal/build.VERSION=${VERSION}'" -o app cmd/${TARGET_APPLICATION}/main.go

FROM alpine:3.21

COPY --from=builder /app/app .

RUN chmod +x app

CMD ["./app"]
