export KUBECONFIG="${KUBECONFIG:-$HOME/.kube/config}"
#The suite can be executed. logging/fast, logging/slow, logging/serial,logging/parallel 
export TEST_SUITE="${TEST_SUITE:-logging/fast}"

#The aws credentials
#export CLUSTER_PROFILE_DIR #the dir where .awscred
#export AWS_SHARED_CREDENTIALS_FILE
#export AWS_CLUSTER_PROFILE_DIR
#export AWS_ACCESS_KEY_ID
#export AWS_SECRET_ACCESS_KEY

#The azure credentials
# export CLUSTER_PROFILE_DIR #the dir where osServicePrincipal.json exist
# export AZURE_AUTH_LOCATION

#The google credentials
#export GOOGLE_APPLICATION_CREDENTIALS="${GOOGLE_APPLICATION_CREDENTIALS:-}"
