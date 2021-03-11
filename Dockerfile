FROM golang:1.16 as builder
WORKDIR /gapp
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o gapp .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /gapp
COPY --from=builder /gapp/gapp .
COPY --from=builder /gapp/index.html .
CMD ["./gapp"]  
