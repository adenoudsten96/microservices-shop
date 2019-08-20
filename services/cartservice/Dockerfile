FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY main .
ENTRYPOINT ./main
LABEL Name=CartService Version=0.0.1
EXPOSE 8081