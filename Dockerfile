#build stage
FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go get -d -v ./...
RUN go build -o /app -v ./...

#final stage
FROM gcr.io/distroless/base-debian12:latest
COPY --from=builder /app/cormorant /app/cormorant
ENTRYPOINT ["/app/cormorant"]
CMD ["cormorant"]
USER nonroot:nonroot
LABEL Name=cormorant Version=1.0.0
