#!/bin/bash
cat <<EOF
{
 "mysecret_credentials":{"value":"$(cat /var/run/ocp-collector/secrets/mysecret/credentials)","error":null},
 "my_little_secret_password":{"value":"$(cat /var/run/ocp-collector/secrets/my-little-secret/password)","error":null},
 "other_secret_some_token":{"value":"$(cat /var/run/ocp-collector/secrets/other-secret/some-token)","error":null}
}
EOF