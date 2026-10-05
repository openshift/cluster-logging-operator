# Dynamic Receiver

This is a manifest for a receiver capable of receiving logs over
a plain http transport.  It is designed to create or teardown its
server based upon the configuration setting.  This is to simulate
load balancer behavior where the service is available from the client
perspective but the backing receiver is not available. The intent of
this deployment is to expose how the collector behaves when the 
receiver goes away and comes back.

* Does the collector recover? Do logs begin to flow again?
* Is the collector wedged trying to use stale sockets or connections?
