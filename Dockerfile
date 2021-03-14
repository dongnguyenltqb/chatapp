FROM golang:1.16 as builder
WORKDIR /chatapp
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o chatapp .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /chatapp
COPY --from=builder /chatapp/chatapp .
COPY --from=builder /chatapp/index.html .
COPY --from=builder /chatapp/.env .
CMD ["./chatapp"]  
