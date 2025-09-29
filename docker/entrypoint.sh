#!/bin/sh
set -e

# Start the Go backend application in the background.
echo "Entrypoint: Starting Go backend..."
/usr/local/bin/server &

# Start Nginx in the foreground.
# 'exec' is crucial as it replaces the shell process with the Nginx process,
# making Nginx the main process (PID 1) that Docker monitors.
echo "Entrypoint: Starting Nginx..."
exec nginx -g 'daemon off;'