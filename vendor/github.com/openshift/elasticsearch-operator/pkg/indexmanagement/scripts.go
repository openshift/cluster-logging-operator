package indexmanagement

const rolloverScript = `
set -euo pipefail
decoded=$(echo $PAYLOAD | base64 -d)
code=$(curl -s "$ES_SERVICE/${POLICY_MAPPING}-write/_rollover?pretty" \
  -w "%{response_code}" \
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
set -uo pipefail
ERRORS=/tmp/errors.txt
echo "" > $ERRORS

writeIndices=$(curl -s $ES_SERVICE/${POLICY_MAPPING}-*/_alias/${POLICY_MAPPING}-write \
  --cacert /etc/indexmanagement/keys/admin-ca \
  -H"Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
  -HContent-Type:application/json)

if echo "$writeIndices" | grep "\"error\"" ; then
  echo "Error while attemping to determine the active write alias: $writeIndices"
  exit 1
fi

CMD=$(cat <<END
import json,sys
r=json.load(sys.stdin)
alias="${POLICY_MAPPING}-write"
try:
  indices = [index for index in r if r[index]['aliases'][alias].get('is_write_index')]
  if len(indices) > 0:
    print(indices[0])
except:
  e = sys.exc_info()[0]
  sys.stderr.write("Error trying to determine the 'write' index from '%r': %r" % (r,e))
  sys.exit(1)
END
)
writeIndex=$(echo "${writeIndices}" | python -c "$CMD" 2>>$ERRORS)
if [ "$?" != "0" ] ; then
  cat $ERRORS
  exit 1
fi


indices=$(curl -s $ES_SERVICE/${POLICY_MAPPING}/_settings/index.creation_date \
  --cacert /etc/indexmanagement/keys/admin-ca \
  -H"Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
  -HContent-Type:application/json)

# Delete in batches of 25 for cases where there are a large number of indices to remove
nowInMillis=$(date +%s%3N)
minAgeFromEpoc=$(($nowInMillis - $MIN_AGE))
CMD=$(cat <<END
import json,sys
r=json.load(sys.stdin)
indices = []
for index in r:
  try:
    if 'settings' in r[index]:
      settings = r[index]['settings']
      if 'index' in settings:
        meta = settings['index']
        if 'creation_date' in meta:
          creation_date = meta['creation_date']
          if int(creation_date) < $minAgeFromEpoc:
            indices.append(index)
        else:
          sys.stderr.write("'creation_date' missing from index settings: %r" % (meta))
      else:
        sys.stderr.write("'index' missing from setting: %r" % (settings))
    else:
      sys.stderr.write("'settings' missing for %r" % (index))
  except:
    e = sys.exc_info()[0]
    sys.stderr.write("Error trying to evaluate index from '%r': %r" % (r,e))
if "$writeIndex" in indices:
  indices.remove("$writeIndex")
for i in range(0, len(indices), 25):
  print ','.join(indices[i:i+25])
END
)
indices=$(echo "${indices}"  | python -c "$CMD" 2>>$ERRORS)
if [ "$?" != "0" ] ; then
  cat $ERRORS
  exit 1
fi
# Dump any findings to stdout but don't error
cat $ERRORS
  
if [ "${indices}" == "" ] ; then
    echo No indices to delete
    exit 0
else
    echo deleting indices: "${indices}"
fi

for sets in ${indices}; do
code=$(curl -s $ES_SERVICE/${sets}?pretty \
  -w "%{response_code}" \
  --cacert /etc/indexmanagement/keys/admin-ca \
  -HContent-Type:application/json \
  -H"Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
  -o /tmp/response.txt \
  -XDELETE )

if [ $code -ne 200 ] ; then
  cat /tmp/response.txt
  exit 1
fi
done
`

var scriptMap = map[string]string{
	"delete":   deleteScript,
	"rollover": rolloverScript,
}
