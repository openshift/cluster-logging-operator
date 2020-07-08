# Cluster Logging Operator Troubleshooting

## Known issues

### After Upgrading CLO and then EO, CLO doesn't know about `kibana` type
The responsibility of managing Kibana was moved to the EO from CLO as part of the 4.5 release feature work. This included creating a new CRD for Kibana that the EO watches for.

If CLO is upgraded before EO, it will try to create a Kibana CR but the Kibana CRD has not yet been created. This will cause CLO to error out with messages indicating it does not know about the type "Kibana".

If this happens, ensure that EO has been updated to at least 4.5 as well so that the Kibana CRD will be created and then delete your running CLO pod. It will restart without the prior error messages and you will see a new Kibana instance roll out (it will be managed by EO instead).