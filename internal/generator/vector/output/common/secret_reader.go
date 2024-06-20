package common

import (
	"fmt"
	"sort"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

const scriptTmpl = ` "%s_%s":{"value":"$(cat %s)","error":null}`

func GenerateSecretReaderScript(secrets helpers.Secrets) string {
	var scriptBuilder strings.Builder
	scriptBuilder.WriteString("#!/bin/bash\ncat <<EOF\n{\n")
	var values []string
	for _, secret := range secrets {
		for key := range secret.Data {
			values = append(values, fmt.Sprintf(scriptTmpl,
				helpers.Replacer.Replace(secret.Name),
				helpers.Replacer.Replace(key),
				common.SecretPath(secret.Name, key)))
		}
	}
	// Sort the values so that test can be reliable
	sort.Strings(values)
	scriptBuilder.WriteString(strings.Join(values, ",\n"))
	scriptBuilder.WriteString("\n}\nEOF")
	return scriptBuilder.String()
}
