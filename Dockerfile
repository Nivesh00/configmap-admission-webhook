# syntax=docker/dockerfile:1
FROM golang:1.24.0

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY module/helper.go ./module/
COPY module/global.go ./module/
COPY main.go ./
COPY module/mutation.go ./module/
COPY module/validation.go ./module/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping

# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/engine/reference/builder/#expose
EXPOSE 80

# Run
CMD ["/docker-gs-ping"]