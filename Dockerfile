FROM golang:1.12

LABEL maintainer="Sail Kumar Ashwin <h1210093@nushigh.edu.sg>"

WORKDIR /checkin

COPY . .

# Download all the dependencies and build
RUN cd cmd/checkin && go build .

# This container exposes port 5000 to the outside world
EXPOSE 5000

# Run the executable
CMD ["cmd/checkin/checkin"]