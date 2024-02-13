#!/bin/bash


# Navigate to the AIAgent project directory and start the Go application
cd AgentAPI
# Source the .env file if your application needs environment variables
# source .env
# Replace `./myGoApp` with the command to start your Go application
./agentapi &
goPID=$!



# Navigate to the AgentAPI project directory and start the Python application
cd ../AIAgent
# Source the .env file if your application needs environment variables
# source .env
python3 main.py &
pythonPID=$!


# Wait for both applications to finish
wait $pythonPID
wait $goPID
