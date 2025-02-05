# Use a lightweight base image for the final container
FROM golang:1.20-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the locally built binary to the container
COPY main .

# Copy the Google Sheets credentials file to the container
COPY creds.json .

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./main"]