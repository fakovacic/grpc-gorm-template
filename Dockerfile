FROM golang:1.13

# Add Maintainer Info
LABEL maintainer="FilipAnteKovacic <filip.ante.kovacic@gmail.com>"

WORKDIR /app/

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o app .

# Expose port 8080 & 8090 to the outside world
EXPOSE 8080
EXPOSE 8090

# Command to run the executable
CMD ["./app"]
