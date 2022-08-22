# Creating overlays for custom builds

An overlay specifies alternate image names and tags for the operator and related images.

Use config/overlays/stable as a starting point for a custom overlay:

1. Copy the `stable` directory tree, for example `cp -r stable custom`
2. Edit the YAML files in your new overlay, see the EDIT comments for more.
3. In the repository root run `make OVERLAY=config/overlays/custom deploy-bundle run-bundle`

NOTE: The overlay directory name will be the OLM channel name for the bundle.
Avoid standard channel names like 'stable' and 'preview', otherwise any name will do.
