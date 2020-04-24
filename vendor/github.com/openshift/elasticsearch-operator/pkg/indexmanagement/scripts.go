package indexmanagement

const rolloverScript = `
set -euox pipefail
decoded=$(echo $PAYLOAD | base64 -d)
code=$(curl "$ES_SERVICE/$POLICY_MAPPING-write/_rollover?pretty" \
  -w "%{response_code}" \
  -sv \
  --cacert /etc/indexmanagement/keys/admin-ca \
  --cert /etc/indexmanagement/keys/admin-cert \
  --key /etc/indexmanagement/keys/admin-key \
  -HContent-Type:application/json \
  -XPOST \
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

indices=$(curl -s $ES_SERVICE/$ALIAS/_settings/index.creation_date \
  --cacert /etc/indexmanagement/keys/admin-ca \
	--cert /etc/indexmanagement/keys/admin-cert \
	--key /etc/indexmanagement/keys/admin-key \
  -HContent-Type:application/json)

CMD=$(cat <<END
import json,sys
r=json.load(sys.stdin)
indices = [index for index in r]
indices.sort()
indices.reverse()
if len(indices) > 0:
  print indices[0] 
END
)
writeIndex=$(echo "${indices}" | python -c "$CMD")


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
  --cert /etc/indexmanagement/keys/admin-cert \
  --key /etc/indexmanagement/keys/admin-key \
  -HContent-Type:application/json \
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
