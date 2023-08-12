package main

import (
	"bufio"
	"flag"
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	faketools "k8s.io/client-go/tools/record"
	"os"
	"sigs.k8s.io/yaml"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/pkg/generator/forwarder"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/spf13/pflag"
)

// HACK - This command is for development use only
func main() {
	utils.InitLoggerWithVerbosity("collection-object-generator", 1)

	yamlFile := flag.String("file", "", "ClusterLogForwarder yaml file. - for stdin")
	secretsFlag := flag.String("objects", "", "colon delimited list of objects in the form of name=key1,key1")
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
		log.V(2).Info("Reading from stdin")
		reader = func() ([]byte, error) {
			stdin := bufio.NewReader(os.Stdin)
			return io.ReadAll(stdin)
		}
	case "":
		log.V(2).Info("received empty yamlfile")
		reader = func() ([]byte, error) { return []byte{}, nil }
	default:
		log.V(2).Info("reading log forwarder from yaml file", "filename", *yamlFile)
		reader = func() ([]byte, error) { return os.ReadFile(*yamlFile) }
	}

	content, err := reader()
	if err != nil {
		log.V(2).Error(err, "Error reading file ", "file", yamlFile)
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
	clf.ObjectMeta.SetUID(types.UID(" "))
	resourceNames := factory.GenerateResourceNames(*clf)

	objects := []runtime.Object{}
	if *secretsFlag != "" {
		for _, raw := range strings.Split(*secretsFlag, ":") {
			rawentry := strings.Split(raw, "=")
			if len(rawentry) == 2 {
				name := rawentry[0]
				secret := runtime.NewSecret(clf.Namespace, name, nil)
				for _, key := range strings.Split(rawentry[1], ",") {
					secret.Data[key] = []byte(key)
				}
				objects = append(objects, secret)
			} else {
				log.V(1).Info("Unable to create rawentry from arg", "entry", raw)
			}
		}

	}
	cl := runtime.NewClusterLogging(clf.Namespace, clf.Name)
	cl.Spec = logging.ClusterLoggingSpec{
		Collection: &logging.CollectionSpec{
			Type: logging.LogCollectionTypeVector,
		},
	}
	ns := &corev1.Namespace{}
	runtime.Initialize(ns, "", clf.Namespace)
	objects = append(objects, ns)

	trustedCM := runtime.NewConfigMap(
		clf.Namespace,
		resourceNames.CaTrustBundle,
		map[string]string{
			constants.TrustedCABundleKey: "",
		},
	)
	trustedCM.Labels[constants.InjectTrustedCABundleLabel] = "true"
	utils.AddOwnerRefToObject(trustedCM, utils.AsOwner(clf))
	objects = append(objects, trustedCM)

	owc := NewObjectWriterClient(objects)

	err = k8shandler.Reconcile(cl, clf, owc, faketools.NewFakeRecorder(1000), "", "", resourceNames)
	if err != nil {
		log.Error(err, "Unable to generate log configuration")
		os.Exit(1)
	}

	for _, value := range owc.Values() {
		m, err := yaml.Marshal(value)
		if err != nil {
			log.Error(err, "Unable to unmarshal resource: %v", m)
		}
		fmt.Println(string(m))
	}

}
