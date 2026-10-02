#!/bin/bash

set -uo pipefail

# The directory used for persisting Vector state, such as on-disk buffers, file checkpoints, and more.
VECTOR_DATA_DIR=%s
echo "Creating the directory used for persisting Vector state $VECTOR_DATA_DIR"
mkdir -p ${VECTOR_DATA_DIR}
# Ensure directory is writable by user and group (group 0 is supplementary group in OpenShift)
chmod 770 ${VECTOR_DATA_DIR}

echo "Starting Vector process..."
exec /usr/bin/vector --config-toml /etc/vector/vector.toml
