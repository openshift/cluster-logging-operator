#!/bin/sh

set -uo

# Ensure data directory exists (init container handles this in production, but needed for tests)
# Extract data_dir from Vector config and create the full path if writable
CONFIG_FILE="/etc/vector/vector.toml"
if [ -f "$CONFIG_FILE" ]; then
    # Parse data_dir from TOML (supports both single and double quotes)
    DATA_DIR=$(grep -E '^\s*data_dir\s*=' "$CONFIG_FILE" | head -1 | sed -E "s/.*data_dir\s*=\s*[\"']([^\"']+)[\"'].*/\1/")
    if [ -n "$DATA_DIR" ]; then
        # Try to create the directory; ignore errors if filesystem is read-only or inaccessible
        mkdir -p "$DATA_DIR" 2>/dev/null || true
    fi
fi

echo "Starting Vector process..."
exec /usr/bin/vector --config-toml /etc/vector/vector.toml
