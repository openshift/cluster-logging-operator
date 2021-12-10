package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"strconv"

	"github.com/openshift/cluster-logging-operator/internal/pkg/generator/forwarder"
	"github.com/spf13/pflag"

	"github.com/ViaQ/logerr/log"
)

// HACK - This command is for development use only
func main() {
	logLevel, present := os.LookupEnv("LOG_LEVEL")
	if present {
		verbosity, err := strconv.Atoi(logLevel)
		if err != nil {
			log.Error(err, "LOG_LEVEL must be an integer", "value", logLevel)
			os.Exit(1)
		}
		log.MustInit("cluster-logging-operator")
		log.SetLogLevel(verbosity)
	} else {
		log.MustInit("cluster-logging-operator")
	}

	yamlFile := flag.String("file", "", "ClusterLogForwarder yaml file. - for stdin")
	includeDefaultLogStore := flag.Bool("include-default-store", true, "Include the default storage when generating the config")
	debugOutput := flag.Bool("debug-output", false, "Generate config normally, but replace output plugins with @stdout plugin, so that records can be printed in collector logs.")
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
			return ioutil.ReadAll(stdin)
		}
	case "":
		log.Info("received empty yamlfile")
		reader = func() ([]byte, error) { return []byte{}, nil }
	default:
		log.Info("reading log forwarder from yaml file", "filename", *yamlFile)
		reader = func() ([]byte, error) { return ioutil.ReadFile(*yamlFile) }
	}

	content, err := reader()
	if err != nil {
		log.Error(err, "Error Unmarshalling file ", "file", yamlFile)
		os.Exit(1)
	}

	log.Info("Finished reading yaml", "content", string(content))

	clfYaml := string(content)
	client := buildClient(clfYaml)
	generatedConfig, err := forwarder.Generate(clfYaml, *includeDefaultLogStore, *debugOutput, client)
	if err != nil {
		log.Error(err, "Unable to generate log configuration")
		os.Exit(1)
	}
	fmt.Println(generatedConfig)
}

func buildClient(clfYaml string) client.Client {

	instance, err := forwarder.UnMarshalClusterLogForwarder(clfYaml)
	if err != nil {
		log.Error(err, "Unable to generate log configuration")
		os.Exit(1)
	}

	secrets := []client.Object{}
	for _, output := range instance.Spec.Outputs {
		if output.Secret != nil {
			s := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: constants.OpenshiftNS,
					Name:      output.Secret.Name,
				},
				Data: map[string][]byte{
					constants.ClientPrivateKey:   []byte("value"),
					constants.ClientCertKey:      []byte("value"),
					constants.TrustedCABundleKey: []byte("value"),
				},
			}
			secrets = append(secrets, s)
		}
	}
	secrets = append(secrets, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.CollectorSecretName,
			Namespace: constants.OpenshiftNS,
		},
		Data: map[string][]byte{
			constants.ClientPrivateKey:   []byte("value"),
			constants.ClientCertKey:      []byte("value"),
			constants.TrustedCABundleKey: []byte("value"),
		},
	})
	return fake.NewClientBuilder().WithObjects(secrets...).Build()
}
