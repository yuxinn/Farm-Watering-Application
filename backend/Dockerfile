# Try to do my looking at the right side code
FROM golang:latest 

# Tells docker the current directory to work on. so everything below is with respect
# To the directory here. Must be FULL PATH
WORKDIR /root/

# Copies from <source> <destination> 
# Source is where your current position is when using the docker command
# Destination is docker, and since we set WORKDIR, destination is /root 
COPY ./main .

# Expose port 8080 to the outside world
# Quite self explainatory
EXPOSE 8000

# Command to run the executable
# Run your bunary
CMD ["/root/main"]

# Always make sure, your binary is EXECUTABLE
# so just chmod 777 main
# Windows issue again. knn