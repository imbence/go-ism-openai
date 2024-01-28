FROM golang
WORKDIR /app
COPY . .
RUN go get
RUN go build -o main .
ENTRYPOINT ["/app/main"]