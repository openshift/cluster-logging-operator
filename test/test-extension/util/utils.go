package util

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	"github.com/openshift/origin/test/extended/util/compat_otp"

	exutil "github.com/openshift/origin/test/extended/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"
	e2eoutput "k8s.io/kubernetes/test/e2e/framework/pod/output"
)

func GetRandomString() string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	buffer := make([]byte, 8)
	for index := range buffer {
		buffer[index] = chars[seed.Intn(len(chars))]
	}
	return string(buffer)
}

// Contain checks if b is an elememt of a
func Contain(a []string, b string) bool {
	for _, c := range a {
		if c == b {
			return true
		}
	}
	return false
}

// ContainSubstring checks if b is a's element's substring
func ContainSubstring(a interface{}, b string) bool {
	switch reflect.TypeOf(a).Kind() {
	case reflect.Slice, reflect.Array:
		s := reflect.ValueOf(a)
		for i := 0; i < s.Len(); i++ {
			if strings.Contains(fmt.Sprintln(s.Index(i)), b) {
				return true
			}
		}
	}
	return false
}

func ProcessTemplate(oc *exutil.CLI, parameters ...string) (string, error) {
	var configFile string
	filename := GetRandomString() + ".json"
	err := wait.PollUntilContextTimeout(context.Background(), 3*time.Second, 15*time.Second, true, func(context.Context) (done bool, err error) {
		output, err := oc.AsAdmin().Run("process").Args(parameters...).OutputToFile(filename)
		if err != nil {
			e2e.Logf("the err:%v, and try next round", err)
			return false, nil
		}
		configFile = output
		return true, nil
	})
	if err != nil {
		return configFile, fmt.Errorf("failed to process template with the provided parameters")
	}
	return configFile, nil
}

func ApplyResourceFromTemplate(oc *exutil.CLI, namespace string, parameters ...string) error {
	if namespace != "" {
		parameters = append(parameters, "-n", namespace)
	}
	file, err := ProcessTemplate(oc, parameters...)
	defer os.Remove(file)
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("Can not process %v", parameters))
	args := []string{"-f", file}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	output, err := oc.AsAdmin().WithoutNamespace().Run("apply").Args(args...).Output()
	if err != nil {
		return fmt.Errorf("can't apply resource: %s", output)
	}
	return nil
}

// WaitForDeploymentPodsToBeReady waits for the specific deployment to be ready
func WaitForDeploymentPodsToBeReady(oc *exutil.CLI, namespace string, name string) {
	var selectors map[string]string
	err := wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
		deployment, err := oc.AdminKubeClient().AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				e2e.Logf("Waiting for deployment/%s to appear\n", name)
				return false, nil
			}
			return false, err
		}
		selectors = deployment.Spec.Selector.MatchLabels
		if deployment.Status.AvailableReplicas == *deployment.Spec.Replicas && deployment.Status.UpdatedReplicas == *deployment.Spec.Replicas {
			e2e.Logf("Deployment %s available (%d/%d)\n", name, deployment.Status.AvailableReplicas, *deployment.Spec.Replicas)
			return true, nil
		}
		e2e.Logf("Waiting for full availability of %s deployment (%d/%d)\n", name, deployment.Status.AvailableReplicas, *deployment.Spec.Replicas)
		return false, nil
	})
	if err != nil && len(selectors) > 0 {
		var labels []string
		for k, v := range selectors {
			labels = append(labels, k+"="+v)
		}
		label := strings.Join(labels, ",")
		_ = oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", namespace, "-l", label).Execute()
		podStatus, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", namespace, "-l", label, "-ojsonpath={.items[*].status.conditions}").Output()
		containerStatus, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", namespace, "-l", label, "-ojsonpath={.items[*].status.containerStatuses}").Output()
		e2e.Failf("deployment %s is not ready:\nconditions: %s\ncontainer status: %s", name, podStatus, containerStatus)
	}
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("deployment %s is not available", name))
}

func WaitForStatefulsetReady(oc *exutil.CLI, namespace string, name string) {
	var selectors map[string]string
	err := wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
		ss, err := oc.AdminKubeClient().AppsV1().StatefulSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				e2e.Logf("Waiting for statefulset/%s to appear\n", name)
				return false, nil
			}
			return false, err
		}
		selectors = ss.Spec.Selector.MatchLabels
		if ss.Status.ReadyReplicas == *ss.Spec.Replicas && ss.Status.UpdatedReplicas == *ss.Spec.Replicas {
			e2e.Logf("statefulset %s available (%d/%d)\n", name, ss.Status.ReadyReplicas, *ss.Spec.Replicas)
			return true, nil
		}
		e2e.Logf("Waiting for full availability of %s statefulset (%d/%d)\n", name, ss.Status.ReadyReplicas, *ss.Spec.Replicas)
		return false, nil
	})
	if err != nil && len(selectors) > 0 {
		var labels []string
		for k, v := range selectors {
			labels = append(labels, k+"="+v)
		}
		label := strings.Join(labels, ",")
		_ = oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", namespace, "-l", label).Execute()
		podStatus, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", namespace, "-l", label, "-ojsonpath={.items[*].status.conditions}").Output()
		containerStatus, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", namespace, "-l", label, "-ojsonpath={.items[*].status.containerStatuses}").Output()
		e2e.Failf("statefulset %s is not ready:\nconditions: %s\ncontainer status: %s", name, podStatus, containerStatus)
	}
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("statefulset %s is not available", name))
}

// WaitForDaemonsetPodsToBeReady waits for all the pods controlled by the ds to be ready
func WaitForDaemonsetPodsToBeReady(oc *exutil.CLI, ns string, name string) {
	var selectors map[string]string
	err := wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
		daemonset, err := oc.AdminKubeClient().AppsV1().DaemonSets(ns).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				e2e.Logf("Waiting for daemonset/%s to appear\n", name)
				return false, nil
			}
			return false, err
		}
		selectors = daemonset.Spec.Selector.MatchLabels
		if daemonset.Status.DesiredNumberScheduled > 0 && daemonset.Status.NumberReady == daemonset.Status.DesiredNumberScheduled && daemonset.Status.UpdatedNumberScheduled == daemonset.Status.DesiredNumberScheduled {
			e2e.Logf("Daemonset/%s is available (%d/%d)\n", name, daemonset.Status.NumberReady, daemonset.Status.DesiredNumberScheduled)
			return true, nil
		}
		e2e.Logf("Waiting for full availability of %s daemonset (%d/%d)\n", name, daemonset.Status.NumberReady, daemonset.Status.DesiredNumberScheduled)
		return false, nil
	})
	if err != nil && len(selectors) > 0 {
		var labels []string
		for k, v := range selectors {
			labels = append(labels, k+"="+v)
		}
		label := strings.Join(labels, ",")
		_ = oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", ns, "-l", label).Execute()
		podStatus, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", ns, "-l", label, "-ojsonpath={.items[*].status.conditions}").Output()
		containerStatus, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", ns, "-l", label, "-ojsonpath={.items[*].status.containerStatuses}").Output()
		e2e.Failf("daemonset %s is not ready:\nconditions: %s\ncontainer status: %s", name, podStatus, containerStatus)
	}
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("Daemonset %s is not available", name))
}

func WaitForPodReadyByLabel(oc *exutil.CLI, ns string, label string) {
	var count int
	err := wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
		pods, err := oc.AdminKubeClient().CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{LabelSelector: label})
		if err != nil {
			return false, err
		}
		count = len(pods.Items)
		if count == 0 {
			e2e.Logf("Waiting for pod with label %s to appear\n", label)
			return false, nil
		}
		ready := true
		for _, pod := range pods.Items {
			if pod.Status.Phase != "Running" {
				ready = false
				break
			}
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if !containerStatus.Ready {
					ready = false
					break
				}
			}
		}
		if !ready {
			e2e.Logf("Waiting for pod with label %s to be ready...\n", label)
		}
		return ready, nil
	})
	if err != nil && count != 0 {
		_ = oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", ns, "-l", label).Execute()
		podStatus, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", ns, "-l", label, "-ojsonpath={.items[*].status.conditions}").Output()
		containerStatus, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", ns, "-l", label, "-ojsonpath={.items[*].status.containerStatuses}").Output()
		e2e.Failf("pod with label %s is not ready:\nconditions: %s\ncontainer status: %s", label, podStatus, containerStatus)
	}
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("pod with label %s is not ready", label))
}

// WaitUntilPodsAreGone waits for pods selected with labelselector to be removed
func WaitUntilPodsAreGone(oc *exutil.CLI, namespace string, labelSelector string) {
	err := wait.PollUntilContextTimeout(context.Background(), 3*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
		output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "--selector="+labelSelector, "-n", namespace).Output()
		if err != nil {
			return false, err
		}
		errstring := fmt.Sprintf("%v", output)
		if strings.Contains(errstring, "No resources found") {
			return true, nil
		}
		return false, nil
	})
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("Error waiting for pods to be removed using label selector %s", labelSelector))
}

func GetPodNamesByLabel(oc *exutil.CLI, ns, label string) ([]string, error) {
	var names []string
	pods, err := oc.AdminKubeClient().CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{LabelSelector: label})
	if err != nil {
		return names, err
	}
	if len(pods.Items) == 0 {
		return names, fmt.Errorf("no pod(s) match label %s in namespace %s", label, ns)
	}
	for _, pod := range pods.Items {
		names = append(names, pod.Name)
	}
	return names, nil
}

// WaitUntilResourceIsGone waits for the resource to be removed cluster
func WaitUntilResourceIsGone(oc *exutil.CLI, kind, name, namespace string) error {
	args := []string{kind, name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	err := wait.PollUntilContextTimeout(context.Background(), 3*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
		output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(args...).Output()
		if err != nil {
			errstring := fmt.Sprintf("%v", output)
			if strings.Contains(errstring, "NotFound") || strings.Contains(errstring, "the server doesn't have a resource type") {
				return true, nil
			}
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("can't remove %s/%s", kind, name)
	}
	return nil
}

// DeleteResourceFromCluster deletes the resource from the cluster
func DeleteResourceFromCluster(oc *exutil.CLI, kind, name, namespace string) error {
	args := []string{kind, name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	msg, err := oc.AsAdmin().WithoutNamespace().Run("delete").Args(args...).Output()
	if err != nil {
		errstring := fmt.Sprintf("%v", msg)
		if strings.Contains(errstring, "NotFound") || strings.Contains(errstring, "the server doesn't have a resource type") {
			return nil
		}
		return err
	}
	return WaitUntilResourceIsGone(oc, kind, name, namespace)
}

func WaitForResourceToAppear(oc *exutil.CLI, kind, name, namespace string) error {
	args := []string{kind, name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	err := wait.PollUntilContextTimeout(context.Background(), 3*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
		e2e.Logf("wait %s %s ready ... ", kind, name)
		output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(args...).Output()
		if err != nil {
			msg := fmt.Sprintf("%v", output)
			if strings.Contains(msg, "NotFound") {
				return false, nil
			}
			return false, err
		}
		e2e.Logf("found %s %s", kind, name)
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("can't find %s/%s", kind, name)
	}
	return nil
}

// SubscriptionObjects objects are used to create operators via OLM
type SubscriptionObjects struct {
	OperatorName       string
	Namespace          string
	OperatorGroup      string // the file used to create operator group
	Subscription       string // the file used to create subscription
	PackageName        string
	OperatorPodLabel   string               //The operator pod label which is used to select pod
	CatalogSource      CatalogSourceObjects `json:",omitempty"`
	SkipCaseWhenFailed bool                 // if true, the case will be skipped when operator is not ready, otherwise, the case will be marked as failed
}

// CatalogSourceObjects defines the source used to subscribe an operator
type CatalogSourceObjects struct {
	Channel         string `json:",omitempty"`
	SourceName      string `json:",omitempty"`
	SourceNamespace string `json:",omitempty"`
}

// WaitForPackagemanifestAppear waits for the packagemanifest to appear in the cluster
// chSource: bool value, true means the packagemanifests' source name must match the so.CatalogSource.SourceName, e.g.: oc get packagemanifests xxxx -l catalog=$source-name
func (so *SubscriptionObjects) WaitForPackagemanifestAppear(oc *exutil.CLI, chSource bool) {
	args := []string{"-n", so.CatalogSource.SourceNamespace, "packagemanifests"}
	if chSource {
		args = append(args, "-l", "catalog="+so.CatalogSource.SourceName)
	} else {
		args = append(args, so.PackageName)
	}
	err := wait.PollUntilContextTimeout(context.Background(), 10*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
		packages, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(args...).Output()
		if err != nil {
			msg := fmt.Sprintf("%v", err)
			if strings.Contains(msg, "No resources found") || strings.Contains(msg, "NotFound") {
				return false, nil
			}
			return false, err
		}
		if strings.Contains(packages, so.PackageName) {
			return true, nil
		}
		e2e.Logf("Waiting for packagemanifest/%s to appear", so.PackageName)
		return false, nil
	})
	if err != nil {
		if so.SkipCaseWhenFailed {
			g.Skip(fmt.Sprintf("Skip the case for can't find packagemanifest/%s", so.PackageName))
		} else {
			e2e.Failf("Packagemanifest %s is not available", so.PackageName)
		}
	}
	//check channel
	if chSource {
		args = append(args, `-ojsonpath={.items[?(@.metadata.name=="`+so.PackageName+`")].status.channels[*].name}`)
	} else {
		args = append(args, `-ojsonpath={.status.channels[*].name}`)
	}
	output, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args(args...).Output()
	channels := strings.Split(output, " ")
	if !Contain(channels, so.CatalogSource.Channel) {
		e2e.Logf("Find channels %v from packagemanifest/%s", channels, so.PackageName)
		if so.SkipCaseWhenFailed {
			g.Skip(fmt.Sprintf("Skip the case for packagemanifest/%s doesn't have target channel %s", so.PackageName, so.CatalogSource.Channel))
		} else {
			e2e.Failf("Packagemanifest %s doesn't have target channel %s", so.PackageName, so.CatalogSource.Channel)
		}
	}
}

// setCatalogSourceObjects set the default values of channel, source namespace and source name if they're not specified
func (so *SubscriptionObjects) setCatalogSourceObjects(oc *exutil.CLI) {
	// set channel
	if so.CatalogSource.Channel == "" {
		so.CatalogSource.Channel = "stable-6.4"
	}

	// set source namespace
	if so.CatalogSource.SourceNamespace == "" {
		so.CatalogSource.SourceNamespace = "openshift-marketplace"
	}

	// set source and check if the packagemanifest exists or not
	if so.CatalogSource.SourceName != "" {
		so.WaitForPackagemanifestAppear(oc, true)
	} else {
		output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("catsrc", "-n", so.CatalogSource.SourceNamespace, "-ojsonpath={.items[*].metadata.name}").Output()
		if err != nil {
			e2e.Logf("can't list catalog source in project %s: %v", so.CatalogSource.SourceNamespace, err)
		}
		catsrcs := strings.Split(output, " ")
		if Contain(catsrcs, "auto-release-app-registry") {
			if Contain(catsrcs, "redhat-operators") {
				// do not subscribe source auto-release-app-registry as we want to test GAed logging in auto release jobs
				so.CatalogSource.SourceName = "redhat-operators"
				so.WaitForPackagemanifestAppear(oc, true)
			} else {
				if so.SkipCaseWhenFailed {
					g.Skip("skip the case because the cluster doesn't have proper catalog source for logging")
				}
			}
		} else if Contain(catsrcs, "qe-app-registry") {
			so.CatalogSource.SourceName = "qe-app-registry"
			so.WaitForPackagemanifestAppear(oc, true)
		} else {
			so.WaitForPackagemanifestAppear(oc, false)
			source, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifests", so.PackageName, "-o", "jsonpath={.status.catalogSource}").Output()
			if err != nil {
				e2e.Logf("error getting catalog source name: %v", err)
			}
			so.CatalogSource.SourceName = source
		}
	}
}

// SubscribeOperator is used to deploy operators
func (so *SubscriptionObjects) SubscribeOperator(oc *exutil.CLI) {
	// check if the namespace exists, if it doesn't exist, create the namespace
	if so.OperatorPodLabel == "" {
		so.OperatorPodLabel = "name=" + so.OperatorName
	}
	_, err := oc.AdminKubeClient().CoreV1().Namespaces().Get(context.Background(), so.Namespace, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			e2e.Logf("The project %s is not found, create it now...", so.Namespace)
			namespaceTemplate := compat_otp.FixturePath("testdata", "logging", "subscription", "namespace.yaml")
			namespaceFile, err := ProcessTemplate(oc, "-f", namespaceTemplate, "-p", "NAMESPACE_NAME="+so.Namespace)
			o.Expect(err).NotTo(o.HaveOccurred())
			defer os.Remove(namespaceFile)
			err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 60*time.Second, true, func(context.Context) (done bool, err error) {
				output, err := oc.AsAdmin().Run("apply").Args("-f", namespaceFile).Output()
				if err != nil {
					if strings.Contains(output, "AlreadyExists") {
						return true, nil
					}
					return false, err
				}
				return true, nil
			})
			compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("can't create project %s", so.Namespace))
		}
	}

	// check the operator group, if no object found, then create an operator group in the project
	og, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", so.Namespace, "og").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	msg := fmt.Sprintf("%v", og)
	if strings.Contains(msg, "No resources found") {
		// create operator group
		ogFile, err := ProcessTemplate(oc, "-n", so.Namespace, "-f", so.OperatorGroup, "-p", "OG_NAME="+so.Namespace, "NAMESPACE="+so.Namespace)
		o.Expect(err).NotTo(o.HaveOccurred())
		defer os.Remove(ogFile)
		err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 60*time.Second, true, func(context.Context) (done bool, err error) {
			output, err := oc.AsAdmin().Run("apply").Args("-f", ogFile, "-n", so.Namespace).Output()
			if err != nil {
				if strings.Contains(output, "AlreadyExists") {
					return true, nil
				}
				return false, err
			}
			return true, nil
		})
		compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("can't create operatorgroup %s in %s project", so.Namespace, so.Namespace))
	}

	// check subscription, if there is no subscription objets, then create one
	installedPackages, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("sub", "-n", so.Namespace, "-ojsonpath={.items[*].spec.name}").Output()
	if err != nil {
		if so.SkipCaseWhenFailed {
			g.Skip("can't get subscriptions: " + err.Error() + ", skip the case")
		} else {
			e2e.Failf("can't get subscriptions: %v", err)
		}
	} else {
		if strings.Contains(installedPackages, so.PackageName) {
			e2e.Logf("operator %s is already installed", so.PackageName)
		} else {
			so.setCatalogSourceObjects(oc)
			//create subscription object
			subscriptionFile, err := ProcessTemplate(oc, "-n", so.Namespace, "-f", so.Subscription, "-p", "PACKAGE_NAME="+so.PackageName, "NAMESPACE="+so.Namespace, "CHANNEL="+so.CatalogSource.Channel, "SOURCE="+so.CatalogSource.SourceName, "SOURCE_NAMESPACE="+so.CatalogSource.SourceNamespace)
			if err != nil {
				if so.SkipCaseWhenFailed {
					g.Skip("hit error when processing subscription template: " + err.Error() + ", skip the case")
				} else {
					e2e.Failf("hit error when processing subscription template: %v", err)
				}
			}
			defer os.Remove(subscriptionFile)
			err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 60*time.Second, true, func(context.Context) (done bool, err error) {
				output, err := oc.AsAdmin().Run("apply").Args("-f", subscriptionFile, "-n", so.Namespace).Output()
				if err != nil {
					if strings.Contains(output, "AlreadyExists") {
						return true, nil
					}
					return false, err
				}
				return true, nil
			})
			if err != nil {
				if so.SkipCaseWhenFailed {
					g.Skip("hit error when creating subscription, skip the case")
				} else {
					e2e.Failf("hit error when creating subscription")
				}
			}
			compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("can't create subscription %s in %s project", so.PackageName, so.Namespace))
			// check status in subscription
			err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 120*time.Second, true, func(context.Context) (done bool, err error) {
				output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", so.Namespace, "sub", so.PackageName, `-ojsonpath={.status.state}`).Output()
				if err != nil {
					e2e.Logf("error getting subscription/%s: %v", so.PackageName, err)
					return false, nil
				}
				return strings.Contains(output, "AtLatestKnown"), nil
			})
			if err != nil {
				out, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", so.Namespace, "sub", so.PackageName, `-ojsonpath={.status.conditions}`).Output()
				e2e.Logf("subscription/%s is not ready, conditions: %v", so.PackageName, out)
				if so.SkipCaseWhenFailed {
					g.Skip(fmt.Sprintf("Skip the case for the operator %s is not ready", so.OperatorName))
				} else {
					e2e.Failf("can't deploy operator %s", so.OperatorName)
				}
			}
		}
	}

	// check pod status
	err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 240*time.Second, true, func(context.Context) (done bool, err error) {
		pods, err := oc.AdminKubeClient().CoreV1().Pods(so.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: so.OperatorPodLabel})
		if err != nil {
			e2e.Logf("Hit error %v when getting pods", err)
			return false, nil
		}
		if len(pods.Items) == 0 {
			e2e.Logf("Waiting for pod with label %s to appear\n", so.OperatorPodLabel)
			return false, nil
		}
		ready := true
		for _, pod := range pods.Items {
			if pod.Status.Phase != "Running" {
				ready = false
				e2e.Logf("Pod %s is not running: %v", pod.Name, pod.Status.Phase)
				break
			}
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if !containerStatus.Ready {
					ready = false
					e2e.Logf("Container %s in pod %s is not ready", containerStatus.Name, pod.Name)
					break
				}
			}
		}
		return ready, nil
	})
	if err != nil {
		_ = oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", so.Namespace, "-l", so.OperatorPodLabel).Execute()
		podStatus, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", so.Namespace, "-l", so.OperatorPodLabel, "-ojsonpath={.items[*].status.conditions}").Output()
		containerStatus, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", so.Namespace, "-l", so.OperatorPodLabel, "-ojsonpath={.items[*].status.containerStatuses}").Output()
		e2e.Logf("pod with label %s is not ready:\nconditions: %s\ncontainer status: %s", so.OperatorPodLabel, podStatus, containerStatus)
		if so.SkipCaseWhenFailed {
			g.Skip(fmt.Sprintf("Skip the case for the operator %s is not ready", so.OperatorName))
		} else {
			e2e.Failf("can't deploy operator %s", so.OperatorName)
		}
	}
}

func (so *SubscriptionObjects) UninstallOperator(oc *exutil.CLI) {
	//csv, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", so.Namespace, "sub/"+so.PackageName, "-ojsonpath={.status.installedCSV}").Output()
	_ = DeleteResourceFromCluster(oc, "subscription", so.PackageName, so.Namespace)
	//_ = oc.AsAdmin().WithoutNamespace().Run("delete").Args("-n", so.Namespace, "csv", csv).Execute()
	_ = oc.AsAdmin().WithoutNamespace().Run("delete").Args("-n", so.Namespace, "csv", "-l", "operators.coreos.com/"+so.PackageName+"."+so.Namespace+"=").Execute()
	// do not remove namespace openshift-logging and openshift-operators-redhat, and preserve the operatorgroup as there may have several operators deployed in one namespace
	// for example: loki-operator
	if so.Namespace != "openshift-logging" && so.Namespace != "openshift-operators-redhat" && !strings.HasPrefix(so.Namespace, "e2e-test-") {
		DeleteNamespace(oc, so.Namespace)
	}
}

func (so *SubscriptionObjects) GetInstalledCSV(oc *exutil.CLI) string {
	installedCSV, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", so.Namespace, "sub", so.PackageName, "-ojsonpath={.status.installedCSV}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	return installedCSV
}

func CreateServiceAccount(oc *exutil.CLI, namespace, name string) error {
	err := oc.AsAdmin().WithoutNamespace().Run("create").Args("serviceaccount", name, "-n", namespace).Execute()
	return err
}

func AddClusterRoleToServiceAccount(oc *exutil.CLI, namespace, serviceAccountName, clusterRole string) error {
	return oc.AsAdmin().WithoutNamespace().Run("adm").Args("policy", "add-cluster-role-to-user", clusterRole, fmt.Sprintf("system:serviceaccount:%s:%s", namespace, serviceAccountName)).Execute()
}

func RemoveClusterRoleFromServiceAccount(oc *exutil.CLI, namespace, serviceAccountName, clusterRole string) error {
	return oc.AsAdmin().WithoutNamespace().Run("adm").Args("policy", "remove-cluster-role-from-user", clusterRole, fmt.Sprintf("system:serviceaccount:%s:%s", namespace, serviceAccountName)).Execute()
}

func DeleteNamespace(oc *exutil.CLI, ns string) {
	err := oc.AdminKubeClient().CoreV1().Namespaces().Delete(context.Background(), ns, metav1.DeleteOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			err = nil
		}
	}
	o.Expect(err).NotTo(o.HaveOccurred())
	err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
		_, err = oc.AdminKubeClient().CoreV1().Namespaces().Get(context.Background(), ns, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("Namespace %s is not deleted in 3 minutes", ns))
}

// Assert the status of a resource
func AssertResourceStatus(oc *exutil.CLI, kind, name, namespace, jsonpath, exptdStatus string) {
	parameters := []string{kind, name, "-o", "jsonpath=" + jsonpath}
	if namespace != "" {
		parameters = append(parameters, "-n", namespace)
	}
	err := wait.PollUntilContextTimeout(context.Background(), 10*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
		status, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(parameters...).Output()
		if err != nil {
			return false, err
		}
		if strings.Compare(status, exptdStatus) != 0 {
			return false, nil
		}
		return true, nil
	})
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("%s/%s value for %s is not %s", kind, name, jsonpath, exptdStatus))
}

// expect: true means we want the resource contain/compare with the expectedContent, false means the resource is expected not to compare with/contain the expectedContent;
// compare: true means compare the expectedContent with the resource content, false means check if the resource contains the expectedContent;
// args are the arguments used to execute command `oc.AsAdmin.WithoutNamespace().Run("get").Args(args...).Output()`;
func CheckResource(oc *exutil.CLI, expect bool, compare bool, expectedContent string, args []string) {
	err := wait.PollUntilContextTimeout(context.Background(), 10*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
		output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(args...).Output()
		if err != nil {
			if strings.Contains(output, "NotFound") {
				return false, nil
			}
			return false, err
		}
		if compare {
			res := strings.Compare(output, expectedContent)
			if (res == 0 && expect) || (res != 0 && !expect) {
				return true, nil
			}
			return false, nil
		}
		res := strings.Contains(output, expectedContent)
		if (res && expect) || (!res && !expect) {
			return true, nil
		}
		return false, nil
	})
	if expect {
		compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("The content doesn't match/contain %s", expectedContent))
	} else {
		compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("The %s still exists in the resource", expectedContent))
	}
}

func GetRouteAddress(oc *exutil.CLI, ns, routeName string) string {
	route, err := oc.AdminRouteClient().RouteV1().Routes(ns).Get(context.Background(), routeName, metav1.GetOptions{})
	o.Expect(err).NotTo(o.HaveOccurred())
	return route.Spec.Host
}

func GetSAToken(oc *exutil.CLI, name, ns string) string {
	token, err := oc.AsAdmin().WithoutNamespace().Run("create").Args("token", name, "-n", ns).Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	return token
}

// enableClusterMonitoring add label `openshift.io/cluster-monitoring: "true"` to the project, and create role/prometheus-k8s rolebinding/prometheus-k8s in the namespace
func EnableClusterMonitoring(oc *exutil.CLI, namespace string) {
	err := oc.AsAdmin().WithoutNamespace().Run("label").Args("ns", namespace, "openshift.io/cluster-monitoring=true").Execute()
	o.Expect(err).NotTo(o.HaveOccurred())

	file := compat_otp.FixturePath("testdata", "logging", "prometheus-k8s-rbac.yaml")
	err = oc.AsAdmin().WithoutNamespace().Run("apply").Args("-n", namespace, "-f", file).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
}

// queryPrometheus returns the promtheus metrics which match the query string
// token: the user token used to run the http request, if it's not specified, it will use the token of sa/prometheus-k8s in openshift-monitoring project
// path: the api path, for example: /api/v1/query?
// query: the metric/alert you want to search, e.g.: es_index_namespaces_total
// action: it can be "GET", "get", "Get", "POST", "post", "Post"
func QueryPrometheus(oc *exutil.CLI, token string, path string, query string, action string) (*prometheusQueryResult, error) {
	var bearerToken string
	var err error
	if token == "" {
		bearerToken = GetSAToken(oc, "prometheus-k8s", "openshift-monitoring")
	} else {
		bearerToken = token
	}
	address := "https://" + GetRouteAddress(oc, "openshift-monitoring", "prometheus-k8s")

	h := make(http.Header)
	h.Add("Content-Type", "application/json")
	h.Add("Authorization", "Bearer "+bearerToken)

	params := url.Values{}
	if len(query) > 0 {
		params.Add("query", query)
	}

	var p prometheusQueryResult
	resp, err := DoHTTPRequest(h, address, path, params.Encode(), action, true, 5, nil, 200)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp, &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func GetMetric(oc *exutil.CLI, token, query string) ([]metric, error) {
	res, err := QueryPrometheus(oc, token, "/api/v1/query", query, "GET")
	if err != nil {
		return []metric{}, err
	}
	return res.Data.Result, nil
}

func CheckMetric(oc *exutil.CLI, token, query string, timeInMinutes int) {
	err := wait.PollUntilContextTimeout(context.Background(), 5*time.Second, time.Duration(timeInMinutes)*time.Minute, true, func(context.Context) (done bool, err error) {
		metrics, err := GetMetric(oc, token, query)
		if err != nil {
			return false, err
		}
		if len(metrics) == 0 {
			e2e.Logf("no metrics found by query: %s, try next time", query)
			return false, nil
		}
		return true, nil
	})
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("can't find metrics by %s in %d minutes", query, timeInMinutes))
}

func FindMetric(oc *exutil.CLI, token, query string, timeInMinutes int) bool {
	err := wait.PollUntilContextTimeout(context.Background(), 5*time.Second, time.Duration(timeInMinutes)*time.Minute, true, func(context.Context) (done bool, err error) {
		metrics, err := GetMetric(oc, token, query)
		if err != nil {
			return false, err
		}
		if len(metrics) == 0 {
			e2e.Logf("no metric match query: %s, try next time", query)
			return false, nil
		}
		return true, nil
	})
	return err == nil
}

func GetAlert(oc *exutil.CLI, token, alertSelector string) ([]alert, error) {
	var al []alert
	alerts, err := QueryPrometheus(oc, token, "/api/v1/alerts", "", "GET")
	if err != nil {
		return al, err
	}
	for i := 0; i < len(alerts.Data.Alerts); i++ {
		if alerts.Data.Alerts[i].Labels.AlertName == alertSelector {
			al = append(al, alerts.Data.Alerts[i])
		}
	}
	return al, nil
}

func CheckAlert(oc *exutil.CLI, token, alertName, status string, timeInMinutes int) {
	err := wait.PollUntilContextTimeout(context.Background(), 30*time.Second, time.Duration(timeInMinutes)*time.Minute, true, func(context.Context) (done bool, err error) {
		alerts, err := GetAlert(oc, token, alertName)
		if err != nil {
			return false, err
		}
		for _, alert := range alerts {
			if strings.Contains(status, alert.State) {
				return true, nil
			}
		}
		e2e.Logf("Waiting for alert %s to be in state %s...", alertName, status)
		return false, nil
	})
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("%s alert is not %s in %d minutes", alertName, status, timeInMinutes))
}

// Check logs from resource
func CheckLogsFromRs(oc *exutil.CLI, kind, name, namespace, containerName, expected string) {
	err := wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
		output, err := oc.AsAdmin().WithoutNamespace().Run("logs").Args(kind+`/`+name, "-n", namespace, "-c", containerName).Output()
		if err != nil {
			e2e.Logf("Can't get logs from resource, error: %s. Trying again", err)
			return false, nil
		}
		if matched, _ := regexp.Match(expected, []byte(output)); !matched {
			e2e.Logf("Can't find the expected string\n")
			return false, nil
		}
		e2e.Logf("Check the logs succeed!!\n")
		return true, nil
	})
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("%s is not expected for %s", expected, name))
}

func GetCurrentCSVFromPackage(oc *exutil.CLI, source, channel, packagemanifest string) string {
	var currentCSV string
	output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifest", "-n", "openshift-marketplace", "-l", "catalog="+source, "-ojsonpath={.items}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	packMS := []PackageManifest{}
	json.Unmarshal([]byte(output), &packMS)
	for _, pm := range packMS {
		if pm.Name == packagemanifest {
			for _, channels := range pm.Status.Channels {
				if channels.Name == channel {
					currentCSV = channels.CurrentCSV
					break
				}
			}
		}
	}
	return currentCSV
}

func CheckCiphers(oc *exutil.CLI, tlsVer string, ciphers []string, server string, caFile string, cloNS string, timeInSec int) error {
	delay := time.Duration(timeInSec) * time.Second
	for _, cipher := range ciphers {
		e2e.Logf("Testing %s...", cipher)

		clPod, err := oc.AdminKubeClient().CoreV1().Pods(cloNS).List(context.Background(), metav1.ListOptions{LabelSelector: "name=cluster-logging-operator"})
		if err != nil {
			return fmt.Errorf("failed to get pods: %w", err)
		}

		cmd := fmt.Sprintf("openssl s_client -%s -cipher %s -CAfile %s -connect %s", tlsVer, cipher, caFile, server)
		result, err := e2eoutput.RunHostCmdWithRetries(cloNS, clPod.Items[0].Name, cmd, 3*time.Second, 30*time.Second)

		if err != nil {
			return fmt.Errorf("failed to run command: %w", err)
		}

		if strings.Contains(string(result), ":error:") {
			errorStr := strings.Split(string(result), ":")[5]
			return fmt.Errorf("error: NOT SUPPORTED (%s)", errorStr)
		} else if strings.Contains(string(result), fmt.Sprintf("Cipher is %s", cipher)) || strings.Contains(string(result), "Cipher    :") {
			e2e.Logf("SUPPORTED")
		} else {
			errorStr := string(result)
			return fmt.Errorf("error: UNKNOWN RESPONSE %s", errorStr)
		}

		time.Sleep(delay)
	}

	return nil
}

func CheckTLSVer(oc *exutil.CLI, tlsVer string, server string, caFile string, cloNS string) error {

	e2e.Logf("Testing TLS %s ", tlsVer)

	clPod, err := oc.AdminKubeClient().CoreV1().Pods(cloNS).List(context.Background(), metav1.ListOptions{LabelSelector: "name=cluster-logging-operator"})
	if err != nil {
		return fmt.Errorf("failed to get pods: %w", err)
	}

	cmd := fmt.Sprintf("openssl s_client -%s -CAfile %s -connect %s", tlsVer, caFile, server)
	result, err := e2eoutput.RunHostCmdWithRetries(cloNS, clPod.Items[0].Name, cmd, 3*time.Second, 30*time.Second)

	if err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}

	if strings.Contains(string(result), ":error:") {
		errorStr := strings.Split(string(result), ":")[5]
		return fmt.Errorf("error: NOT SUPPORTED (%s)", errorStr)
	} else if strings.Contains(string(result), "Cipher is ") || strings.Contains(string(result), "Cipher    :") {
		e2e.Logf("SUPPORTED")
	} else {
		errorStr := string(result)
		return fmt.Errorf("error: UNKNOWN RESPONSE %s", errorStr)
	}

	return nil
}

func CheckTLSProfile(oc *exutil.CLI, profile string, algo string, server string, caFile string, cloNS string, timeInSec int) bool {
	var ciphers []string
	var tlsVer string

	if profile == "modern" {
		e2e.Logf("Modern profile is currently not supported, please select from old, intermediate, custom")
		return false
	}

	if IsFipsEnabled(oc) {
		switch profile {
		case "old":
			e2e.Logf("Checking old profile with TLS v1.3")
			tlsVer = "tls1_3"
			err := CheckTLSVer(oc, tlsVer, server, caFile, cloNS)
			o.Expect(err).NotTo(o.HaveOccurred())

			e2e.Logf("Checking old profile with TLS v1.2")
			switch algo {
			case "ECDSA":
				{
					ciphers = []string{"ECDHE-ECDSA-AES128-GCM-SHA256", "ECDHE-ECDSA-AES256-GCM-SHA384", "ECDHE-ECDSA-CHACHA20-POLY1305", "ECDHE-ECDSA-AES128-SHA256", "ECDHE-ECDSA-AES128-SHA", "ECDHE-ECDSA-AES256-SHA384", "ECDHE-ECDSA-AES256-SHA"}
				}
			case "RSA":
				{
					ciphers = []string{"ECDHE-RSA-AES256-GCM-SHA384", "ECDHE-RSA-AES128-GCM-SHA256", "ECDHE-RSA-AES128-GCM-SHA256"}
				}
			}
			tlsVer = "tls1_2"
			err = CheckCiphers(oc, tlsVer, ciphers, server, caFile, cloNS, timeInSec)
			o.Expect(err).NotTo(o.HaveOccurred())

		case "intermediate":
			e2e.Logf("Setting alogorith to %s", algo)
			e2e.Logf("Checking intermediate profile with TLS v1.3")
			tlsVer = "tls1_3"
			err := CheckTLSVer(oc, tlsVer, server, caFile, cloNS)
			o.Expect(err).NotTo(o.HaveOccurred())

			e2e.Logf("Checking intermediate ciphers with TLS v1.3")
			//  as openssl-3.0.7-24.el9 in CLO pod failed as below, no such issue in openssl-3.0.9-2.fc38.  use TLS 1.3 to test TSL 1.2 here.
			//  openssl s_client -tls1_2 -cipher ECDHE-RSA-AES128-GCM-SHA256 -CAfile /run/secrets/kubernetes.io/serviceaccount/service-ca.crt -connect lokistack-sample-gateway-http:8081
			//  20B4A391FFFF0000:error:1C8000E9:Provider routines:kdf_tls1_prf_derive:ems not enabled:providers/implementations/kdfs/tls1_prf.c:200:
			//  20B4A391FFFF0000:error:0A08010C:SSL routines:tls1_PRF:unsupported:ssl/t1_enc.c:83:
			tlsVer = "tls1_3"
			switch algo {
			case "ECDSA":
				{
					ciphers = []string{"ECDHE-ECDSA-AES128-GCM-SHA256", "ECDHE-ECDSA-AES256-GCM-SHA384"}
				}
			case "RSA":
				{
					ciphers = []string{"ECDHE-RSA-AES128-GCM-SHA256", "ECDHE-RSA-AES256-GCM-SHA384"}
				}
			}
			err = CheckCiphers(oc, tlsVer, ciphers, server, caFile, cloNS, timeInSec)
			o.Expect(err).NotTo(o.HaveOccurred())

			e2e.Logf("Checking intermediate profile with TLS v1.1")
			tlsVer = "tls1_1"
			err = CheckCiphers(oc, tlsVer, ciphers, server, caFile, cloNS, timeInSec)
			o.Expect(err).To(o.HaveOccurred())

		case "custom":
			e2e.Logf("Setting alogorith to %s", algo)
			e2e.Logf("Checking custom profile with TLS v1.3")
			tlsVer = "tls1_3"
			err := CheckTLSVer(oc, tlsVer, server, caFile, cloNS)
			o.Expect(err).NotTo(o.HaveOccurred())

			e2e.Logf("Checking custom profile ciphers with TLS v1.3")
			//  as openssl-3.0.7-24.el9 in CLO pod failed as below, no such issue in openssl-3.0.9-2.fc38.  use TLS 1.3 to test TSL 1.2 here.
			//  openssl s_client -tls1_2 -cipher ECDHE-RSA-AES128-GCM-SHA256 -CAfile /run/secrets/kubernetes.io/serviceaccount/service-ca.crt -connect lokistack-sample-gateway-http:8081
			//  20B4A391FFFF0000:error:1C8000E9:Provider routines:kdf_tls1_prf_derive:ems not enabled:providers/implementations/kdfs/tls1_prf.c:200:
			//  20B4A391FFFF0000:error:0A08010C:SSL routines:tls1_PRF:unsupported:ssl/t1_enc.c:83:
			tlsVer = "tls1_3"
			switch algo {
			case "ECDSA":
				{
					ciphers = []string{"ECDHE-ECDSA-CHACHA20-POLY1305", "ECDHE-ECDSA-AES128-GCM-SHA256"}
				}
			case "RSA":
				{
					ciphers = []string{"ECDHE-RSA-AES128-GCM-SHA256"}
				}
			}
			err = CheckCiphers(oc, tlsVer, ciphers, server, caFile, cloNS, timeInSec)
			o.Expect(err).NotTo(o.HaveOccurred())

			e2e.Logf("Checking ciphers on in custom profile with TLS v1.3")
			tlsVer = "tls1_3"
			switch algo {
			case "ECDSA", "RSA":
				{
					ciphers = []string{"TLS_AES_128_GCM_SHA256"}
				}
			}
			err = CheckCiphers(oc, tlsVer, ciphers, server, caFile, cloNS, timeInSec)
			o.Expect(err).To(o.HaveOccurred())
		}

	} else {
		switch profile {
		case "old":
			e2e.Logf("Checking old profile with TLS v1.3")
			tlsVer = "tls1_3"
			err := CheckTLSVer(oc, tlsVer, server, caFile, cloNS)
			o.Expect(err).NotTo(o.HaveOccurred())

			e2e.Logf("Checking old profile with TLS v1.2")
			switch algo {
			case "ECDSA":
				{
					ciphers = []string{"ECDHE-ECDSA-AES128-GCM-SHA256", "ECDHE-ECDSA-AES256-GCM-SHA384", "ECDHE-ECDSA-CHACHA20-POLY1305", "ECDHE-ECDSA-AES128-SHA256", "ECDHE-ECDSA-AES128-SHA", "ECDHE-ECDSA-AES256-SHA384", "ECDHE-ECDSA-AES256-SHA"}
				}
			case "RSA":
				{
					ciphers = []string{"ECDHE-RSA-AES128-GCM-SHA256", "ECDHE-RSA-AES128-SHA256", "ECDHE-RSA-AES128-SHA", "ECDHE-RSA-AES256-SHA", "AES128-GCM-SHA256", "AES256-GCM-SHA384", "AES128-SHA256", "AES128-SHA", "AES256-SHA"}
				}
			}
			tlsVer = "tls1_2"
			err = CheckCiphers(oc, tlsVer, ciphers, server, caFile, cloNS, timeInSec)
			o.Expect(err).NotTo(o.HaveOccurred())

			e2e.Logf("Checking old profile with TLS v1.1")
			//  remove these ciphers as openssl-3.0.7-24.el9  s_client -tls1_1 -cipher <ciphers> failed.
			switch algo {
			case "ECDSA":
				{
					ciphers = []string{"ECDHE-ECDSA-AES128-SHA", "ECDHE-ECDSA-AES256-SHA"}
				}
			case "RSA":
				{
					ciphers = []string{"AES128-SHA", "AES256-SHA"}
				}
			}
			tlsVer = "tls1_1"
			err = CheckCiphers(oc, tlsVer, ciphers, server, caFile, cloNS, timeInSec)
			o.Expect(err).NotTo(o.HaveOccurred())

		case "intermediate":
			e2e.Logf("Setting alogorith to %s", algo)
			e2e.Logf("Checking intermediate profile with TLS v1.3")
			tlsVer = "tls1_3"
			err := CheckTLSVer(oc, tlsVer, server, caFile, cloNS)
			o.Expect(err).NotTo(o.HaveOccurred())

			e2e.Logf("Checking intermediate profile ciphers with TLS v1.2")
			tlsVer = "tls1_2"
			switch algo {
			case "ECDSA":
				{
					ciphers = []string{"ECDHE-ECDSA-AES128-GCM-SHA256", "ECDHE-ECDSA-AES256-GCM-SHA384", "ECDHE-ECDSA-CHACHA20-POLY1305"}
				}
			case "RSA":
				{
					ciphers = []string{"ECDHE-RSA-AES128-GCM-SHA256", "ECDHE-RSA-AES256-GCM-SHA384"}
				}
			}
			err = CheckCiphers(oc, tlsVer, ciphers, server, caFile, cloNS, timeInSec)
			o.Expect(err).NotTo(o.HaveOccurred())

			e2e.Logf("Checking intermediate profile with TLS v1.1")
			// replace checkCiphers with checkTLSVer as we needn't check all v1.1 Ciphers
			tlsVer = "tls1_1"
			err = CheckTLSVer(oc, tlsVer, server, caFile, cloNS)
			o.Expect(err).To(o.HaveOccurred())

		case "custom":
			e2e.Logf("Setting alogorith to %s", algo)

			e2e.Logf("Checking custom profile with TLS v1.3")
			tlsVer = "tls1_3"
			err := CheckTLSVer(oc, tlsVer, server, caFile, cloNS)
			o.Expect(err).NotTo(o.HaveOccurred())

			e2e.Logf("Checking custom profile with TLS v1.2")
			tlsVer = "tls1_2"
			switch algo {
			case "ECDSA":
				{
					ciphers = []string{"ECDHE-ECDSA-AES128-GCM-SHA256"}
				}
			case "RSA":
				{
					ciphers = []string{"ECDHE-RSA-AES128-GCM-SHA256"}
				}
			}
			err = CheckCiphers(oc, tlsVer, ciphers, server, caFile, cloNS, timeInSec)
			o.Expect(err).NotTo(o.HaveOccurred())

			e2e.Logf("Checking ciphers not in custom profile with TLS v1.3")
			tlsVer = "tls1_3"
			switch algo {
			case "ECDSA":
				{
					ciphers = []string{"ECDHE-ECDSA-AES128-GCM-SHA256"}
				}
			case "RSA":
				{
					ciphers = []string{"TLS_AES_128_GCM_SHA256"}
				}
			}
			err = CheckCiphers(oc, tlsVer, ciphers, server, caFile, cloNS, timeInSec)
			o.Expect(err).To(o.HaveOccurred())
		}
	}
	return true
}

func CheckClusterOperatorsRunning(oc *exutil.CLI) (bool, error) {
	jpath := `{range .items[*]}{.metadata.name}:{.status.conditions[?(@.type=='Available')].status}{':'}{.status.conditions[?(@.type=='Degraded')].status}{'\n'}{end}`
	output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("clusteroperators.config.openshift.io", "-o", "jsonpath="+jpath).Output()
	if err != nil {
		return false, fmt.Errorf("failed to execute 'oc get clusteroperators.config.openshift.io' command: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		e2e.Logf("%s", line)
		parts := strings.Split(line, ":")
		available := parts[1] == "True"
		degraded := parts[2] == "False"

		if !available || !degraded {
			return false, nil
		}
	}

	return true, nil
}

func WaitForClusterOperatorsRunning(oc *exutil.CLI) {
	e2e.Logf("Wait a minute to allow the cluster to reconcile the config changes.")
	time.Sleep(1 * time.Minute)
	err := wait.PollUntilContextTimeout(context.Background(), 3*time.Minute, 21*time.Minute, true, func(context.Context) (done bool, err error) {
		return CheckClusterOperatorsRunning(oc)
	})
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("Failed to wait for operators to be running: %v", err))
}

func DoHTTPRequest(header http.Header, address, path, query, method string, quiet bool, attempts int, requestBody io.Reader, expectedStatusCode int) ([]byte, error) {
	us, err := buildURL(address, path, query)
	if err != nil {
		return nil, err
	}
	if !quiet {
		e2e.Logf("the URL is: %s", us)
	}

	req, err := http.NewRequest(strings.ToUpper(method), us, requestBody)
	if err != nil {
		return nil, err
	}

	req.Header = header

	var tr *http.Transport
	proxy := GetProxyFromEnv()
	if len(proxy) > 0 {
		proxyURL, err := url.Parse(proxy)
		o.Expect(err).NotTo(o.HaveOccurred())
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyURL(proxyURL),
		}
	} else {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	client := &http.Client{Transport: tr}

	var resp *http.Response
	success := false

	for attempts > 0 {
		attempts--

		resp, err = client.Do(req)
		if err != nil {
			e2e.Logf("error sending request %v", err)
			continue
		}
		if resp.StatusCode != expectedStatusCode {
			buf, _ := io.ReadAll(resp.Body) // nolint
			e2e.Logf("Error response from server: %s %s (%v), attempts remaining: %d", resp.Status, string(buf), err, attempts)
			if err := resp.Body.Close(); err != nil {
				e2e.Logf("error closing body: %v", err)
			}
			// sleep 5 second before doing next request
			time.Sleep(5 * time.Second)
			continue
		}
		success = true
		break
	}
	if !success {
		return nil, fmt.Errorf("run out of attempts while querying the server")
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			e2e.Logf("error closing body: %v", err)
		}
	}()
	return io.ReadAll(resp.Body)
}

// buildURL concats a url `http://foo/bar` with a path `/buzz`.
func buildURL(u, p, q string) (string, error) {
	url, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	url.Path = path.Join(url.Path, p)
	url.RawQuery = q
	return url.String(), nil
}

// ConvertInterfaceToArray converts interface{} to []string
/*
	example of interface{}:
	  [
	    timestamp,
		log data
	  ],
	  [
	    timestamp,
		count
	  ]
*/
func ConvertInterfaceToArray(t interface{}) []string {
	var data []string
	switch reflect.TypeOf(t).Kind() {
	case reflect.Slice, reflect.Array:
		s := reflect.ValueOf(t)
		for i := 0; i < s.Len(); i++ {
			data = append(data, fmt.Sprint(s.Index(i)))
		}
	}
	return data
}

// send logs over http
func PostDataToHttpserver(oc *exutil.CLI, clfNS string, httpURL string, postJsonString string) bool {
	collectorPods, err := oc.AdminKubeClient().CoreV1().Pods(clfNS).List(context.Background(), metav1.ListOptions{LabelSelector: "app.kubernetes.io/component=collector"})
	if err != nil || len(collectorPods.Items) < 1 {
		e2e.Logf("failed to get pods by label app.kubernetes.io/component=collector")
		return false
	}
	//ToDo, send logs to httpserver using service ca, oc get cm/openshift-service-ca.crt -o json |jq '.data."service-ca.crt"'
	cmd := `curl -s -k -w "%{http_code}" ` + httpURL + " -d '" + postJsonString + "'"
	result, err := e2eoutput.RunHostCmdWithRetries(clfNS, collectorPods.Items[0].Name, cmd, 3*time.Second, 30*time.Second)
	if err != nil {
		e2e.Logf("Show more status as data can not be sent to httpserver")
		oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", clfNS, "endpoints").Output()
		oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", clfNS, "pods").Output()
		return false
	}
	if result == "200" {
		return true
	} else {
		e2e.Logf("Show result as return code is not 200")
		e2e.Logf("result=%v", result)
		return false
	}
}

// create job for rapiddast test
// Run a job to do rapiddast, the scan result will be written into pod logs and store in artifactdirPath
func RapidastScan(oc *exutil.CLI, ns, configFile string, scanPolicyFile string, apiGroupName string) (bool, error) {
	//update the token and create a new config file
	content, err := os.ReadFile(configFile)
	jobName := "rapidast-" + GetRandomString()
	if err != nil {
		e2e.Logf("rapidastScan abort! Open file %s failed", configFile)
		e2e.Logf("rapidast result: riskHigh=unknown riskMedium=unknown")
		return false, err
	}
	defer oc.AsAdmin().WithoutNamespace().Run("adm").Args("policy", "remove-cluster-role-from-user", "cluster-admin", fmt.Sprintf("system:serviceaccount:%s:default", ns)).Execute()
	oc.AsAdmin().WithoutNamespace().Run("adm").Args("policy", "add-cluster-role-to-user", "cluster-admin", fmt.Sprintf("system:serviceaccount:%s:default", ns)).Execute()
	token := GetSAToken(oc, "default", ns)
	originConfig := string(content)
	targetConfig := strings.Replace(originConfig, "Bearer sha256~xxxxxxxx", "Bearer "+token, -1)
	newConfigFile := "/tmp/logdast" + GetRandomString()
	f, err := os.Create(newConfigFile)
	if err != nil {
		e2e.Logf("rapidastScan abort! prepare configfile %s failed", newConfigFile)
		e2e.Logf("rapidast result: riskHigh=unknown riskMedium=unknown")
		return false, err
	}
	defer f.Close()
	defer exec.Command("rm", newConfigFile).Output()
	f.WriteString(targetConfig)

	//Create configmap
	err = oc.WithoutNamespace().Run("create").Args("-n", ns, "configmap", jobName, "--from-file=rapidastconfig.yaml="+newConfigFile, "--from-file=customscan.policy="+scanPolicyFile).Execute()
	if err != nil {
		e2e.Logf("rapidastScan abort! create configmap rapidast-configmap failed")
		e2e.Logf("rapidast result: riskHigh=unknown riskMedium=unknown")
		return false, err
	}

	//Create job
	loggingBaseDir := compat_otp.FixturePath("testdata", "logging")
	jobTemplate := filepath.Join(loggingBaseDir, "rapidast/job_rapidast.yaml")
	err = ApplyResourceFromTemplate(oc, ns, "-f", jobTemplate, "-p", "NAME="+jobName)
	if err != nil {
		e2e.Logf("rapidastScan abort! create rapidast job failed")
		e2e.Logf("rapidast result: riskHigh=unknown riskMedium=unknown")
		return false, err
	}
	//Waiting up to 3 minutes until pod Failed or Success
	wait.PollUntilContextTimeout(context.Background(), 30*time.Second, 3*time.Minute, true, func(context.Context) (done bool, err error) {
		jobStatus, err1 := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", ns, "pod", "-l", "job-name="+jobName, "-ojsonpath={.items[0].status.phase}").Output()
		e2e.Logf(" rapidast Job status %s ", jobStatus)
		if err1 != nil {
			return false, nil
		}
		if jobStatus == "Pending" || jobStatus == "Running" {
			return false, nil
		}
		if jobStatus == "Failed" {
			e2e.Logf("rapidast-job %s failed", jobName)
			return true, nil
		}
		if jobStatus == "Succeeded" {
			return true, nil
		}
		return false, nil
	})
	// Get the rapidast pod name
	jobPods, err := oc.AdminKubeClient().CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{LabelSelector: "job-name=" + jobName})
	if err != nil {
		e2e.Logf("rapidastScan abort! can not find rapidast scan job ")
		e2e.Logf("rapidast result: riskHigh=unknown riskMedium=unknown")
		return false, err
	}
	podLogs, err := oc.AsAdmin().WithoutNamespace().Run("logs").Args("-n", ns, jobPods.Items[0].Name).Output()
	if err != nil {
		e2e.Logf("rapidastScan abort! can not fetch logs from rapidast-scan pod %s", jobPods.Items[0].Name)
		e2e.Logf("rapidast result: riskHigh=unknown riskMedium=unknown")
		return false, err
	}

	// Copy DAST Report into $ARTIFACT_DIR
	artifactAvaiable := true
	artifactdirPath := os.Getenv("ARTIFACT_DIR")
	if artifactdirPath == "" {
		artifactAvaiable = false
	}
	info, err := os.Stat(artifactdirPath)
	if err != nil {
		e2e.Logf("%s doesn't exist", artifactdirPath)
		artifactAvaiable = false
	} else if !info.IsDir() {
		e2e.Logf("%s isn't a directory", artifactdirPath)
		artifactAvaiable = false
	}

	if artifactAvaiable {
		rapidastResultsSubDir := artifactdirPath + "/rapiddastresultslogging"
		err = os.MkdirAll(rapidastResultsSubDir, 0755)
		if err != nil {
			e2e.Logf("failed to create %s", rapidastResultsSubDir)
		}
		artifactFile := rapidastResultsSubDir + "/" + apiGroupName + "_rapidast.result.txt"
		e2e.Logf("Write report into %s", artifactFile)
		f1, err := os.Create(artifactFile)
		if err != nil {
			e2e.Logf("failed to create artifactFile %s", artifactFile)
		}
		defer f1.Close()
		_, err = f1.WriteString(podLogs)
		if err != nil {
			e2e.Logf("failed to write logs into artifactFile %s", artifactFile)
		}
	} else {
		// print pod logs if artifactdirPath is not writable
		e2e.Logf("#oc logs -n %s %s \n %s", jobPods.Items[0].Name, ns, podLogs)
	}

	//return false, if high risk is reported
	podLogA := strings.Split(podLogs, "\n")
	riskHigh := 0
	riskMedium := 0
	re1 := regexp.MustCompile(`"riskdesc": .*High`)
	re2 := regexp.MustCompile(`"riskdesc": .*Medium`)
	for _, item := range podLogA {
		if re1.MatchString(item) {
			riskHigh++
		}
		if re2.MatchString(item) {
			riskMedium++
		}
	}
	e2e.Logf("rapidast result: riskHigh=%v riskMedium=%v", riskHigh, riskMedium)

	if riskHigh > 0 {
		return false, fmt.Errorf("high risk alert, please check the scan result report")
	}
	return true, nil
}

// Create a linux audit policy to generate audit logs in one schedulable worker
func GenLinuxAuditLogsOnWorker(oc *exutil.CLI) (string, error) {
	workerNodes, err := compat_otp.GetSchedulableLinuxWorkerNodes(oc)
	if err != nil || len(workerNodes) == 0 {
		return "", fmt.Errorf("can not find schedulable worker to enable audit policy")
	}
	result, err := compat_otp.DebugNodeWithChroot(oc, workerNodes[0].Name, "bash", "-c", "auditctl -w /var/log/pods/ -p rwa -k logging-qe-test-read-write-pod-logs")
	if err != nil && strings.Contains(result, "Rule exists") {
		//Note: we still provide the nodeName here, the policy will be deleted if `defer deleteLinuxAuditPolicyFromNodes` is called.
		return workerNodes[0].Name, nil
	}
	return workerNodes[0].Name, err
}

// delete the linux audit policy
func DeleteLinuxAuditPolicyFromNode(oc *exutil.CLI, nodeName string) error {
	if nodeName == "" {
		return fmt.Errorf("nodeName can not be empty")
	}
	_, err := compat_otp.DebugNodeWithChroot(oc, nodeName, "bash", "-c", "auditctl -W /var/log/pods/ -p rwa -k logging-qe-test-read-write-pod-logs")
	return err
}

func ListFilesAndDirectories(dirPath string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return files, err
	}

	for _, entry := range entries {
		path := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			pathes, err := ListFilesAndDirectories(path)
			if err != nil {
				return files, err
			}
			files = append(files, pathes...)
		}
		//fmt.Printf("\n%s\n", path)
		files = append(files, path)
	}
	return files, nil
}
