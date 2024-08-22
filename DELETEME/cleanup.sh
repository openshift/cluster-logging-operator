#!/bin/bash
# Remove description text and sort keys for easier diff.

yq eval 'del(.. | .description?)' "$1" | yq eval 'sort_keys(..)'

