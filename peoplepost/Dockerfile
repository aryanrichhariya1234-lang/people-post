FROM golang:1.26.2

WORKDIR /app

# Install dependencies first (cache optimization)
COPY go.mod go.sum ./
RUN go mod download

# Copy full code
COPY . .

# Build binary
RUN go build -o main ./cmd/server

# Expose your backend port
EXPOSE 3000

# Run app
CMD ["./main"]