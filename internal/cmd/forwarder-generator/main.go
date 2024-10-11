package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/pkg/generator/forwarder"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/spf13/pflag"
)

// HACK - This command is for development use only
func main() {
	utils.InitLogger("forwarder-generator")

	yamlFile := flag.String("file", "", "ClusterLogForwarder yaml file. - for stdin")
	debugOutput := flag.Bool("debug-output", false, "Generate config normally, but replace output plugins with @stdout plugin, so that records can be printed in collector logs.")
	secretsFlag := flag.String("secrets", "", "colon delimited list of secrets in the form of name=key1,key1")
	help := flag.Bool("help", false, "This message")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	if *help {
		pflag.Usage()
		os.Exit(1)
	}
	log.V(1).Info("Forwarder Generator Main", "Args", os.Args)

	var reader func() ([]byte, error)
	switch *yamlFile {
	case "-":
		log.Info("Reading from stdin")
		reader = func() ([]byte, error) {
			stdin := bufio.NewReader(os.Stdin)
			return io.ReadAll(stdin)
		}
	case "":
		log.Info("received empty yamlfile")
		reader = func() ([]byte, error) { return []byte{}, nil }
	default:
		log.Info("reading log forwarder from yaml file", "filename", *yamlFile)
		reader = func() ([]byte, error) { return os.ReadFile(*yamlFile) }
	}

	content, err := reader()
	if err != nil {
		log.Error(err, "Error reading file ", "file", yamlFile)
		os.Exit(1)
	}

	clfYaml := string(content)
	log.Info("Finished reading yaml", "content", clfYaml)
	clf, err := forwarder.UnMarshalClusterLogForwarder(clfYaml)
	if err != nil {
		log.Error(err, "Error UnMarshalling CLF", "file", yamlFile)
		os.Exit(1)
	}
	log.V(2).Info("Unmarshalled", "forwarder", clfYaml)

	var secrets []runtime.Object
	if *secretsFlag != "" {
		for _, raw := range strings.Split(*secretsFlag, ":") {
			rawentry := strings.Split(raw, "=")
			if len(rawentry) == 2 {
				name := rawentry[0]
				secret := runtime.NewSecret(clf.Namespace, name, nil)
				for _, key := range strings.Split(rawentry[1], ",") {
					secret.Data[key] = []byte(key)
				}
				secrets = append(secrets, secret)
			} else {
				log.V(1).Info("Unable to create rawentry from arg", "entry", raw)
			}
		}
	}

	client := fake.NewFakeClient(secrets...)
	generatedConfig, err := forwarder.Generate(clfYaml, *debugOutput, client)
	if err != nil {
		log.Error(err, "Unable to generate log configuration")
		os.Exit(1)
	}
	fmt.Println(generatedConfig)
}
