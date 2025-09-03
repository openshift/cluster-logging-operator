package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	utilflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"

	"github.com/openshift-eng/openshift-tests-extension/pkg/cmd"
	e "github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	et "github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
	g "github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"
	clusterdiscovery "github.com/openshift/origin/pkg/clioptions/clusterdiscovery"
	"github.com/openshift/origin/pkg/clioptions/imagesetup"
	"github.com/openshift/origin/test/extended/util/image"
	"k8s.io/klog/v2"

	exutil "github.com/openshift/origin/test/extended/util"

	// If using ginkgo, import your tests here
	_ "github.com/openshift/cluster-logging-operator/test/test-extension/specs"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()
	pflag.CommandLine.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)

	// Extension registry
	registry := e.NewRegistry()
	ext := e.NewExtension("openshift-logging", "non-payload", "cluster-logging-operator")

	ext.AddSuite(e.Suite{
		Name: "logging/fast",
		Qualifiers: []string{
			`!name.contains("[Slow]")`,
		},
	})
	ext.AddSuite(e.Suite{
		Name: "logging/slow",
		Qualifiers: []string{
			`name.contains("[Slow]")`,
		},
	})
	ext.AddSuite(e.Suite{
		Name: "logging/serial",
		Qualifiers: []string{
			`(name.contains("[Serial]") && !name.contains("[Disruptive]"))`,
		},
	})
	ext.AddSuite(e.Suite{
		Name: "logging/parallel",
		Qualifiers: []string{
			`(!name.contains("[Serial]") && !name.contains("[Disruptive]"))`,
		},
	})
	specs, err := g.BuildExtensionTestSpecsFromOpenShiftGinkgoSuite()
	if err != nil {
		panic(fmt.Sprintf("couldn't build extension test specs from ginkgo: %+v", err.Error()))
	}
	// specs = specs.MustFilter([]string{`name.contains("sig-openshift-logging")`}) //This works
	specs, err = specs.MustSelect(et.NameContains("sig-openshift-logging"))
	if err != nil {
		panic(fmt.Sprintf("no specs found: %v", err))
	}

	specs.AddBeforeAll(func() {
		config, err := clusterdiscovery.DecodeProvider(os.Getenv("TEST_PROVIDER"), false, false, nil)
		if err != nil {
			panic(err)
		}
		if err := clusterdiscovery.InitializeTestFramework(exutil.TestContext, config, false); err != nil {
			panic(err)
		}
		klog.V(4).Infof("Loaded test configuration: %#v", exutil.TestContext)

		exutil.TestContext.ReportDir = os.Getenv("TEST_JUNIT_DIR")

		image.InitializeImages(os.Getenv("KUBE_TEST_REPO"))

		if err := imagesetup.VerifyImages(); err != nil {
			panic(err)
		}

	})
	ext.AddSpecs(specs)
	registry.Register(ext)

	root := &cobra.Command{
		Long: "OpenShift Logging extended tests",
	}
	root.AddCommand(
		cmd.DefaultExtensionCommands(registry)...,
	)

	f := flag.CommandLine.Lookup("v")
	root.PersistentFlags().AddGoFlag(f)
	pflag.CommandLine = pflag.NewFlagSet("empty", pflag.ExitOnError)
	flag.CommandLine = flag.NewFlagSet("empty", flag.ExitOnError)
	exutil.InitStandardFlags()

	if err := func() error {
		return root.Execute()
	}(); err != nil {
		os.Exit(1)
	}
}
