# Use an official Ubuntu base image
FROM ubuntu:latest

# Install Go
RUN apt-get update && \
    apt-get -y install wget && \
    wget https://dl.google.com/go/go1.21.6.linux-amd64.tar.gz && \
    tar -xvf go1.21.6.linux-amd64.tar.gz && \
    mv go /usr/local

# Set Go environment variables
ENV GOROOT=/usr/local/go
ENV GOPATH=$HOME/go
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# Install Python
RUN apt-get -y install python3 python3-pip

# Copy the Go project
COPY ./AgentAPI /AgentAPI
WORKDIR /AgentAPI
# Build your Go project here, if necessary
RUN go build -o agentapi


# Copy the Python project
COPY ./AIAgent /AIAgent
WORKDIR /AIAgent
# Install Python dependencies
RUN pip3 install -r requirements.txt

WORKDIR /

COPY run.sh /run.sh
RUN chmod 755 /run.sh
# Optional: specify default command
CMD ["./run.sh"]
