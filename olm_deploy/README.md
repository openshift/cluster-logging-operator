# OLM Registry

This directory is used to build an OLM registry image containing the bundle defined by ../bundle.
This is required by the current CI setup which expects to build a registry.

The registry is no longer required and may be removed in future.
You can build a bundle image and deploy it directly with operator-sdk  as described in ../Makefile.
