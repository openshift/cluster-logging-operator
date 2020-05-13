package indexmanagement

const rolloverScript = `
set -euox pipefail
decoded=$(echo $PAYLOAD | base64 -d)
code=$(curl "$ES_SERVICE/${POLICY_MAPPING}-write/_rollover?pretty" \
  -w "%{response_code}" \
  -sv \
  --cacert /etc/indexmanagement/keys/admin-ca \
  -HContent-Type:application/json \
  -XPOST \
  -H"Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
  -o /tmp/response.txt \
  -d $decoded)
if [ "$code" == "200" ] ; then
  exit 0 
fi
cat /tmp/response.txt
exit 1
`
const deleteScript = `
set -euox pipefail

writeIndices=$(curl -s $ES_SERVICE/${POLICY_MAPPING}-*/_alias/${POLICY_MAPPING}-write \
  --cacert /etc/indexmanagement/keys/admin-ca \
  -H"Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
  -HContent-Type:application/json)

CMD=$(cat <<END
import json,sys
r=json.load(sys.stdin)
alias="${POLICY_MAPPING}-write"
indices = [index for index in r if r[index]['aliases'][alias]['is_write_index']]
if len(indices) > 0:
  print indices[0] 
END
)
writeIndex=$(echo "${writeIndices}" | python -c "$CMD")


indices=$(curl -s $ES_SERVICE/${POLICY_MAPPING}/_settings/index.creation_date \
  --cacert /etc/indexmanagement/keys/admin-ca \
  -H"Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
  -HContent-Type:application/json)

nowInMillis=$(date +%s%3N)
minAgeFromEpoc=$(($nowInMillis - $MIN_AGE))
CMD=$(cat <<END
import json,sys
r=json.load(sys.stdin)
indices = [index for index in r if int(r[index]['settings']['index']['creation_date']) < $minAgeFromEpoc ]
if "$writeIndex" in indices:
  indices.remove("$writeIndex")
indices.sort()
print ','.join(indices)
END
)
indices=$(echo "${indices}"  | python -c "$CMD")
  
if [ "${indices}" == "" ] ; then
    echo No indices to delete
    exit 0
else
    echo deleting indices: "${indices}"
fi

code=$(curl -sv $ES_SERVICE/${indices}?pretty \
  -w "%{response_code}" \
  --cacert /etc/indexmanagement/keys/admin-ca \
  -HContent-Type:application/json \
  -H"Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
  -o /tmp/response.txt \
  -XDELETE )

if [ "$code" == "200" ] ; then
  exit 0
fi
cat /tmp/response.txt
exit 1
`

var scriptMap = map[string]string{
	"delete":   deleteScript,
	"rollover": rolloverScript,
}
