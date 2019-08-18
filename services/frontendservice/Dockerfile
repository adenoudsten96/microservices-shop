#build stage
FROM golang:alpine AS builder
WORKDIR /go/src/app
COPY . .
RUN apk add --no-cache git
RUN go get -d -v ./...
RUN go install -v ./...

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /go/bin/app /app
RUN mkdir templates
RUN mkdir static
COPY --from=builder /go/src/app/templates templates/
COPY --from=builder /go/src/app/static static/
ENTRYPOINT ./app
LABEL Name=FrontendService Version=0.0.1
EXPOSE 8082
