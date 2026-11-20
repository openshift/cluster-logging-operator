#!/bin/sh

set -u

# Ensure data directory exists (init container handles this in production, but needed for tests)
if [ -n "${VECTOR_DATA_DIR:-}" ]; then
    # Try to create the directory; ignore errors if filesystem is read-only or inaccessible
    mkdir -p "$VECTOR_DATA_DIR" 2>/dev/null || true
fi

echo "Starting Vector process..."
exec /usr/bin/vector --config-toml /etc/vector/vector.toml
