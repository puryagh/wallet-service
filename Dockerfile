# Tip: fetch grpcurl binary and copy it to destination build alpine image
FROM fullstorydev/grpcurl:latest AS grpcurl-bin

# Build stage
FROM golang:1.25-alpine AS builder

# Install required build tools
RUN apk add --no-cache gcc musl-dev make git openssh

# Set working directory
WORKDIR /app

# Configure Git to use SSH
RUN git config --global url."https://puryagh:glpat-RHYxJbSsPn_5MM1Qugxf@git.iranpishran.com/".insteadOf "https://git.iranpishran.com/"

# Copy go mod files
COPY go.mod ./
COPY go.sum ./
COPY schemas ./schemas

# Download dependencies
RUN GOPRIVATE=git.iranpishran.com/* go mod download
RUN GOPRIVATE=git.iranpishran.com/* go mod tidy

# Copy source code
COPY . .

# Build the application
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/bin/noghrestan-be ./cmd/noghrestan-be



# Final stage
FROM alpine:3.19

# Install ca-certificates for HTTPS and tzdata for timezone support
RUN apk add --no-cache ca-certificates tzdata curl

COPY --from=grpcurl-bin /bin/grpcurl /usr/local/bin/grpcurl

HEALTHCHECK --interval=10s --timeout=3s CMD grpcurl -plaintext localhost:8080 pb.FrameworkHealthService/Check || exit 1

# Create non-root user
RUN adduser -D -H -h /app appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/noghrestan-be .
COPY --from=builder /app/schemas .

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose necessary ports
EXPOSE 8080 8081

# Set environment variables
ENV TZ=UTC \
    GO_ENV=production

# Run the application
ENTRYPOINT ["./noghrestan-be"]