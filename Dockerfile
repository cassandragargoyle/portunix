# Create a minimal Docker image
FROM scratch

# Set the working directory inside the container
WORKDIR /app

# Copy the config program into the container at /app
COPY portunix .

# Command to run the executable
CMD ["./portunix]