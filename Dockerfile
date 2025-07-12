# Use a simple approach - copy pre-built binary
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy pre-built binary (build locally first)
COPY build/porta .

# Copy configuration files
COPY examples/etc/ ./config/

# Clear proxy settings
ENV http_proxy=
ENV https_proxy=
ENV ftp_proxy=
ENV HTTP_PROXY=
ENV HTTPS_PROXY=
ENV FTP_PROXY=

# Expose port
EXPOSE 8080

# Run the application
CMD ["./porta", "-c", "./config/config.yaml", "-p", "8080"] 