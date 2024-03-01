package vector

// RunVectorScript is the run-vector.sh script for launching the Vector container process
// will override content of /scripts/run-vector.sh in https://github.com/ViaQ/vector
const RunVectorScript = `#!/bin/bash
# The directory used for persisting Vector state, such as on-disk buffers, file checkpoints, and more.
VECTOR_DATA_DIR=%s
echo "Creating the directory used for persisting Vector state $VECTOR_DATA_DIR"
mkdir -p ${VECTOR_DATA_DIR}

echo "Checking for buffer lock files"
# Vector does not appear to release locks when the process terminates. Try to clean up for restart
pushd ${VECTOR_DATA_DIR}
  locks=$(ls -R . | awk '/:$/&&f{s=$0;f=0}
/:$/&&!f{sub(/:$/,"");s=$0;f=1;next}
NF&&f{ print s"/"$0 }' | grep \.lock)
  rc=$?
  if [ $rc -gt 1 ] ; then
    echo "Error checking for buffer lock files returncode=$rc: '${locks}'"
    exit $rc
  fi
  sleep 10s  #try to allow other owners to finish or cleanup
  for lock in "${locks}"; do
    echo "removing file: ${lock}"
    rm $lock
  done
popd

echo "Starting Vector process..."
exec /usr/bin/vector --config-toml /etc/vector/vector.toml
`
