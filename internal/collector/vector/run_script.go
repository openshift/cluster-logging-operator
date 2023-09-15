package vector

// RunVectorScript is the run-vector.sh script for launching the Vector container process
// will override content of /scripts/run-vector.sh in https://github.com/ViaQ/vector
const RunVectorScript = `
#!/bin/bash
# The directory used for persisting Vector state, such as on-disk buffers, file checkpoints, and more.
VECTOR_DATA_DIR=%s
echo "Creating the directory used for persisting Vector state $VECTOR_DATA_DIR"
mkdir -p $VECTOR_DATA_DIR
echo "Starting Vector process..."
exec "/usr/bin/vector"

`
