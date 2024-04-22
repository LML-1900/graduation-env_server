FROM golang:1.21

ENV GOPROXY=https://goproxy.cn,direct

# set a directory for the app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# copy all the files to the container
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /env_server

# tell the port number the container should expose
EXPOSE 50051

# run
ENTRYPOINT ["/env_server"]
