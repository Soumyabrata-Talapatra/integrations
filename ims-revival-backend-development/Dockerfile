# Dockerfile for Go Application with GitLab Token Authentication
# Stage 1: Build the Go binary
FROM golang:latest AS builder
# Use build arguments for GitLab username and token
ARG GITLAB_USERNAME
ARG GITLAB_TOKEN
# Set up Git configuration with the username and token for downloading Go modules
RUN git config --global url."https://git_token:${GITLAB_TOKEN}@gitlab.cept.gov.in".insteadOf "https://gitlab.cept.gov.in"
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
# Stage 2: Create the final image
FROM alpine:latest
# Install necessary packages including tzdata for timezone handling
RUN apk --no-cache add ca-certificates tzdata
# Set the timezone environment variable
ENV TZ=Asia/Kolkata
# Create a non-root user and set the working directory
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser
WORKDIR /home/appuser/
# Copy the Go binary and configuration file from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/config.yaml .
EXPOSE 8080
CMD ["sh", "-c", "until wget -qO- http://localhost:15021/healthz/ready; do sleep 5; done && ./main"]
