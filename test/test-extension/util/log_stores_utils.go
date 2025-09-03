package util

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	o "github.com/onsi/gomega"
	exutil "github.com/openshift/origin/test/extended/util"
	"github.com/openshift/origin/test/extended/util/compat_otp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	e2eoutput "k8s.io/kubernetes/test/e2e/framework/pod/output"
)

type CertsConf struct {
	ServerName string
	Namespace  string
	PassPhrase string //client private key passphrase
}

func (certs CertsConf) GenerateCerts(oc *exutil.CLI, keysPath string) {
	generateCertsSH := compat_otp.FixturePath("testdata", "logging", "external-log-stores", "cert_generation.sh")
	domain, err := GetAppDomain(oc)
	o.Expect(err).NotTo(o.HaveOccurred())
	cmd := []string{generateCertsSH, keysPath, certs.Namespace, certs.ServerName, domain}
	if certs.PassPhrase != "" {
		cmd = append(cmd, certs.PassPhrase)
	}
	err = exec.Command("sh", cmd...).Run()
	o.Expect(err).NotTo(o.HaveOccurred())
}

type externalES struct {
	namespace                  string
	version                    string // support 6 and 7
	serverName                 string // ES cluster name, configmap/sa/deploy/svc name
	httpSSL                    bool   // `true` means enable `xpack.security.http.ssl`
	clientAuth                 bool   // `true` means `xpack.security.http.ssl.client_authentication: required`, only can be set to `true` when httpSSL is `true`
	clientPrivateKeyPassphrase string // only works when clientAuth is true
	userAuth                   bool   // `true` means enable user auth
	username                   string // shouldn't be empty when `userAuth: true`
	password                   string // shouldn't be empty when `userAuth: true`
	secretName                 string //the name of the secret for the collector to use, it shouldn't be empty when `httpSSL: true` or `userAuth: true`
	loggingNS                  string //the namespace where the collector pods deployed in
}

func (es externalES) createPipelineSecret(oc *exutil.CLI, keysPath string) {
	// create pipeline secret if needed
	cmd := []string{"secret", "generic", es.secretName, "-n", es.loggingNS}
	if es.clientAuth {
		cmd = append(cmd, "--from-file=tls.key="+keysPath+"/client.key", "--from-file=tls.crt="+keysPath+"/client.crt", "--from-file=ca-bundle.crt="+keysPath+"/ca.crt")
		if es.clientPrivateKeyPassphrase != "" {
			cmd = append(cmd, "--from-literal=passphrase="+es.clientPrivateKeyPassphrase)
		}
	} else if es.httpSSL && !es.clientAuth {
		cmd = append(cmd, "--from-file=ca-bundle.crt="+keysPath+"/ca.crt")
	}
	if es.userAuth {
		cmd = append(cmd, "--from-literal=username="+es.username, "--from-literal=password="+es.password)
	}

	err := oc.AsAdmin().WithoutNamespace().Run("create").Args(cmd...).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
	o.Expect(WaitForResourceToAppear(oc, "secret", es.secretName, es.loggingNS)).NotTo(o.HaveOccurred())
}

func (es externalES) deploy(oc *exutil.CLI) {
	// create SA
	err := oc.AsAdmin().WithoutNamespace().Run("create").Args("serviceaccount", es.serverName, "-n", es.namespace).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
	o.Expect(WaitForResourceToAppear(oc, "serviceaccount", es.serverName, es.namespace)).NotTo(o.HaveOccurred())

	if es.userAuth {
		o.Expect(es.username).NotTo(o.BeEmpty(), "Please provide username!")
		o.Expect(es.password).NotTo(o.BeEmpty(), "Please provide password!")
	}

	if es.httpSSL || es.clientAuth || es.userAuth {
		o.Expect(es.secretName).NotTo(o.BeEmpty(), "Please provide pipeline secret name!")

		// create a temporary directory
		baseDir := compat_otp.FixturePath("testdata", "logging")
		keysPath := filepath.Join(baseDir, "temp"+GetRandomString())
		defer exec.Command("rm", "-r", keysPath).Output()
		err = os.MkdirAll(keysPath, 0755)
		o.Expect(err).NotTo(o.HaveOccurred())

		cert := CertsConf{es.serverName, es.namespace, es.clientPrivateKeyPassphrase}
		cert.GenerateCerts(oc, keysPath)
		// create secret for ES if needed
		if es.httpSSL || es.clientAuth {
			err = oc.WithoutNamespace().Run("create").Args("secret", "generic", "-n", es.namespace, es.serverName, "--from-file=elasticsearch.key="+keysPath+"/server.key", "--from-file=elasticsearch.crt="+keysPath+"/server.crt", "--from-file=admin-ca="+keysPath+"/ca.crt").Execute()
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(WaitForResourceToAppear(oc, "secret", es.serverName, es.namespace)).NotTo(o.HaveOccurred())
		}

		// create pipeline secret in logging project
		es.createPipelineSecret(oc, keysPath)
	}

	// get file path per the configurations
	filePath := []string{"testdata", "logging", "external-log-stores", "elasticsearch", es.version}
	if es.httpSSL {
		filePath = append(filePath, "https")
	} else {
		o.Expect(es.clientAuth).NotTo(o.BeTrue(), "Unsupported configuration, please correct it!")
		filePath = append(filePath, "http")
	}
	if es.userAuth {
		filePath = append(filePath, "user_auth")
	} else {
		filePath = append(filePath, "no_user")
	}

	// create configmap
	cmFilePath := append(filePath, "configmap.yaml")
	cmFile := compat_otp.FixturePath(cmFilePath...)
	cmPatch := []string{"-f", cmFile, "-p", "NAMESPACE=" + es.namespace, "-p", "NAME=" + es.serverName}
	if es.userAuth {
		cmPatch = append(cmPatch, "-p", "USERNAME="+es.username, "-p", "PASSWORD="+es.password)
	}
	if es.httpSSL {
		if es.clientAuth {
			cmPatch = append(cmPatch, "-p", "CLIENT_AUTH=required")
		} else {
			cmPatch = append(cmPatch, "-p", "CLIENT_AUTH=none")
		}
	}

	// set xpack.ml.enable to false when the architecture is not amd64
	nodes, err := compat_otp.GetSchedulableLinuxWorkerNodes(oc)
	o.Expect(err).NotTo(o.HaveOccurred())
	for _, node := range nodes {
		if node.Status.NodeInfo.Architecture != "amd64" {
			cmPatch = append(cmPatch, "-p", "MACHINE_LEARNING=false")
			break
		}
	}

	err = ApplyResourceFromTemplate(oc, es.namespace, cmPatch...)
	o.Expect(err).NotTo(o.HaveOccurred())

	// create deployment and expose svc
	deployFilePath := append(filePath, "deployment.yaml")
	deployFile := compat_otp.FixturePath(deployFilePath...)
	err = ApplyResourceFromTemplate(oc, es.namespace, "-f", deployFile, "-p", "NAMESPACE="+es.namespace, "-p", "NAME="+es.serverName)
	o.Expect(err).NotTo(o.HaveOccurred())
	WaitForDeploymentPodsToBeReady(oc, es.namespace, es.serverName)
	err = oc.AsAdmin().WithoutNamespace().Run("expose").Args("-n", es.namespace, "deployment", es.serverName, "--name="+es.serverName).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())

	// expose route
	if es.httpSSL {
		err = oc.AsAdmin().WithoutNamespace().Run("create").Args("-n", es.namespace, "route", "passthrough", "--service="+es.serverName, "--port=9200").Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
	} else {
		err = oc.AsAdmin().WithoutNamespace().Run("expose").Args("svc/"+es.serverName, "-n", es.namespace, "--port=9200").Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
	}
}

func (es externalES) remove(oc *exutil.CLI) {
	_ = DeleteResourceFromCluster(oc, "route", es.serverName, es.namespace)
	_ = DeleteResourceFromCluster(oc, "service", es.serverName, es.namespace)
	_ = DeleteResourceFromCluster(oc, "configmap", es.serverName, es.namespace)
	_ = DeleteResourceFromCluster(oc, "deployment", es.serverName, es.namespace)
	_ = DeleteResourceFromCluster(oc, "serviceaccount", es.serverName, es.namespace)
	if es.httpSSL || es.userAuth {
		_ = DeleteResourceFromCluster(oc, "secret", es.secretName, es.loggingNS)
	}
	if es.httpSSL {
		_ = DeleteResourceFromCluster(oc, "secret", es.serverName, es.namespace)
	}
}

func (es externalES) getPodName(oc *exutil.CLI) string {
	esPods, err := oc.AdminKubeClient().CoreV1().Pods(es.namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "app=" + es.serverName})
	o.Expect(err).NotTo(o.HaveOccurred())
	var names []string
	for i := 0; i < len(esPods.Items); i++ {
		names = append(names, esPods.Items[i].Name)
	}
	return names[0]
}

func (es externalES) baseCurlString() string {
	curlString := "curl -H \"Content-Type: application/json\""
	if es.userAuth {
		curlString += " -u " + es.username + ":" + es.password
	}
	if es.httpSSL {
		if es.clientAuth {
			curlString += " --cert /usr/share/elasticsearch/config/secret/elasticsearch.crt --key /usr/share/elasticsearch/config/secret/elasticsearch.key"
		}
		curlString += " --cacert /usr/share/elasticsearch/config/secret/admin-ca -s https://localhost:9200/"
	} else {
		curlString += " -s http://localhost:9200/"
	}
	return curlString
}

func (es externalES) getIndices(oc *exutil.CLI) ([]ESIndex, error) {
	cmd := es.baseCurlString() + "_cat/indices?format=JSON"
	stdout, err := e2eoutput.RunHostCmdWithRetries(es.namespace, es.getPodName(oc), cmd, 3*time.Second, 9*time.Second)
	indices := []ESIndex{}
	json.Unmarshal([]byte(stdout), &indices)
	return indices, err
}

func (es externalES) waitForIndexAppear(oc *exutil.CLI, indexName string) {
	err := wait.PollUntilContextTimeout(context.Background(), 10*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
		indices, err := es.getIndices(oc)
		count := 0
		for _, index := range indices {
			if strings.Contains(index.Index, indexName) {
				if index.Health != "red" {
					docCount, _ := strconv.Atoi(index.DocsCount)
					count += docCount
				}
			}
		}
		if count > 0 && err == nil {
			return true, nil
		}
		return false, err
	})
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("Index %s didn't appear or the doc count is 0 in last 3 minutes.", indexName))
}

func (es externalES) getDocCount(oc *exutil.CLI, indexName string, queryString string) (int64, error) {
	cmd := es.baseCurlString() + indexName + "*/_count?format=JSON -d '" + queryString + "'"
	stdout, err := e2eoutput.RunHostCmdWithRetries(es.namespace, es.getPodName(oc), cmd, 5*time.Second, 30*time.Second)
	res := CountResult{}
	json.Unmarshal([]byte(stdout), &res)
	return res.Count, err
}

func (es externalES) waitForProjectLogsAppear(oc *exutil.CLI, projectName string, indexName string) {
	query := "{\"query\": {\"match_phrase\": {\"kubernetes.namespace_name\": \"" + projectName + "\"}}}"
	err := wait.PollUntilContextTimeout(context.Background(), 10*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
		logCount, err := es.getDocCount(oc, indexName, query)
		if err != nil {
			return false, err
		}
		if logCount > 0 {
			return true, nil
		}
		return false, nil
	})
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("The logs of project %s didn't collected to index %s in last 180 seconds.", projectName, indexName))
}

func (es externalES) searchDocByQuery(oc *exutil.CLI, indexName string, queryString string) SearchResult {
	cmd := es.baseCurlString() + indexName + "*/_search?format=JSON -d '" + queryString + "'"
	stdout, err := e2eoutput.RunHostCmdWithRetries(es.namespace, es.getPodName(oc), cmd, 3*time.Second, 30*time.Second)
	o.Expect(err).ShouldNot(o.HaveOccurred())
	res := SearchResult{}
	//data := bytes.NewReader([]byte(stdout))
	//_ = json.NewDecoder(data).Decode(&res)
	json.Unmarshal([]byte(stdout), &res)
	return res
}

func (es externalES) removeIndices(oc *exutil.CLI, indexName string) {
	cmd := es.baseCurlString() + indexName + " -X DELETE"
	_, err := e2eoutput.RunHostCmdWithRetries(es.namespace, es.getPodName(oc), cmd, 3*time.Second, 30*time.Second)
	o.Expect(err).ShouldNot(o.HaveOccurred())
}

/*

type rsyslog struct {
	serverName          string //the name of the rsyslog server, it's also used to name the svc/cm/sa/secret
	namespace           string //the namespace where the rsyslog server deployed in
	tls                 bool
	secretName          string //the name of the secret for the collector to use
	loggingNS           string //the namespace where the collector pods deployed in
	clientKeyPassphrase string //client private key passphrase
}

func (r rsyslog) createPipelineSecret(oc *exutil.CLI, keysPath string) {
	secret := resource{"secret", r.secretName, r.loggingNS}
	cmd := []string{"secret", "generic", secret.name, "-n", secret.namespace, "--from-file=ca-bundle.crt=" + keysPath + "/ca.crt"}
	if r.clientKeyPassphrase != "" {
		cmd = append(cmd, "--from-file=tls.key="+keysPath+"/client.key", "--from-file=tls.crt="+keysPath+"/client.crt", "--from-literal=passphrase="+r.clientKeyPassphrase)
	}

	err := oc.AsAdmin().WithoutNamespace().Run("create").Args(cmd...).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
	secret.WaitForResourceToAppear(oc)
}

func (r rsyslog) deploy(oc *exutil.CLI) {
	// create SA
	sa := resource{"serviceaccount", r.serverName, r.namespace}
	err := oc.WithoutNamespace().Run("create").Args("serviceaccount", sa.name, "-n", sa.namespace).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
	sa.WaitForResourceToAppear(oc)
	err = oc.AsAdmin().WithoutNamespace().Run("adm").Args("policy", "add-scc-to-user", "privileged", fmt.Sprintf("system:serviceaccount:%s:%s", r.namespace, r.serverName), "-n", r.namespace).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
	filePath := []string{"testdata", "logging", "external-log-stores", "rsyslog"}
	// create secrets if needed
	if r.tls {
		o.Expect(r.secretName).NotTo(o.BeEmpty())
		// create a temporary directory
		baseDir := compat_otp.FixturePath("testdata", "logging")
		keysPath := filepath.Join(baseDir, "temp"+getRandomString())
		defer exec.Command("rm", "-r", keysPath).Output()
		err = os.MkdirAll(keysPath, 0755)
		o.Expect(err).NotTo(o.HaveOccurred())

		cert := certsConf{r.serverName, r.namespace, r.clientKeyPassphrase}
		cert.generateCerts(oc, keysPath)
		// create pipelinesecret
		r.createPipelineSecret(oc, keysPath)
		// create secret for rsyslog server
		err = oc.AsAdmin().WithoutNamespace().Run("create").Args("secret", "generic", r.serverName, "-n", r.namespace, "--from-file=server.key="+keysPath+"/server.key", "--from-file=server.crt="+keysPath+"/server.crt", "--from-file=ca_bundle.crt="+keysPath+"/ca.crt").Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		filePath = append(filePath, "secure")
	} else {
		filePath = append(filePath, "insecure")
	}

	// create configmap/deployment/svc
	cm := resource{"configmap", r.serverName, r.namespace}
	cmFilePath := append(filePath, "configmap.yaml")
	cmFile := compat_otp.FixturePath(cmFilePath...)
	err = cm.applyFromTemplate(oc, "-f", cmFile, "-n", r.namespace, "-p", "NAMESPACE="+r.namespace, "-p", "NAME="+r.serverName)
	o.Expect(err).NotTo(o.HaveOccurred())

	deploy := resource{"deployment", r.serverName, r.namespace}
	deployFilePath := append(filePath, "deployment.yaml")
	deployFile := compat_otp.FixturePath(deployFilePath...)
	err = deploy.applyFromTemplate(oc, "-f", deployFile, "-n", r.namespace, "-p", "NAMESPACE="+r.namespace, "-p", "NAME="+r.serverName)
	o.Expect(err).NotTo(o.HaveOccurred())
	WaitForDeploymentPodsToBeReady(oc, r.namespace, r.serverName)

	svc := resource{"svc", r.serverName, r.namespace}
	svcFilePath := append(filePath, "svc.yaml")
	svcFile := compat_otp.FixturePath(svcFilePath...)
	err = svc.applyFromTemplate(oc, "-f", svcFile, "-n", r.namespace, "-p", "NAMESPACE="+r.namespace, "-p", "NAME="+r.serverName)
	o.Expect(err).NotTo(o.HaveOccurred())
}

func (r rsyslog) remove(oc *exutil.CLI) {
	resource{"serviceaccount", r.serverName, r.namespace}.clear(oc)
	if r.tls {
		resource{"secret", r.serverName, r.namespace}.clear(oc)
		resource{"secret", r.secretName, r.loggingNS}.clear(oc)
	}
	resource{"configmap", r.serverName, r.namespace}.clear(oc)
	resource{"deployment", r.serverName, r.namespace}.clear(oc)
	resource{"svc", r.serverName, r.namespace}.clear(oc)
}

func (r rsyslog) getPodName(oc *exutil.CLI) string {
	pods, err := oc.AdminKubeClient().CoreV1().Pods(r.namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "component=" + r.serverName})
	o.Expect(err).NotTo(o.HaveOccurred())
	var names []string
	for i := 0; i < len(pods.Items); i++ {
		names = append(names, pods.Items[i].Name)
	}
	return names[0]
}

func (r rsyslog) checkData(oc *exutil.CLI, expect bool, filename string) {
	cmd := "ls -l /var/log/clf/" + filename
	if expect {
		err := wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 60*time.Second, true, func(context.Context) (done bool, err error) {
			stdout, err := e2eoutput.RunHostCmdWithRetries(r.namespace, r.getPodName(oc), cmd, 3*time.Second, 15*time.Second)
			if err != nil {
				if strings.Contains(err.Error(), "No such file or directory") {
					return false, nil
				}
				return false, err
			}
			return strings.Contains(stdout, filename), nil
		})
		compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("The %s doesn't exist", filename))
	} else {
		err := wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 60*time.Second, true, func(context.Context) (done bool, err error) {
			stdout, err := e2eoutput.RunHostCmdWithRetries(r.namespace, r.getPodName(oc), cmd, 3*time.Second, 15*time.Second)
			if err != nil {
				if strings.Contains(err.Error(), "No such file or directory") {
					return true, nil
				}
				return false, err
			}
			return strings.Contains(stdout, "No such file or directory"), nil
		})
		compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("The %s exists", filename))
	}
}

func (r rsyslog) checkDataContent(oc *exutil.CLI, expect bool, filename string, expected string) {
	re := regexp.MustCompile(`.*` + expected + `.*`)
	cmd := "tail -10 /var/log/clf/" + filename
	err := wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 60*time.Second, true, func(context.Context) (done bool, err error) {
		stdout, err := e2eoutput.RunHostCmdWithRetries(r.namespace, r.getPodName(oc), cmd, 3*time.Second, 15*time.Second)
		if err != nil {
			return false, err
		}
		return re.MatchString(stdout), nil
	})
	if expect {
		compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("The file %s not exist or the file doesn't contain %s", filename, expected))
	} else {
		compat_otp.AssertWaitPollWithErr(err, fmt.Sprintf("The file %s shouldn't include %s", filename, expected))
	}
}

type fluentdServer struct {
	serverName                 string //the name of the fluentd server, it's also used to name the svc/cm/sa/secret
	namespace                  string //the namespace where the fluentd server deployed in
	serverAuth                 bool
	clientAuth                 bool   // only can be set when serverAuth is true
	clientPrivateKeyPassphrase string //only can be set when clientAuth is true
	sharedKey                  string //if it's not empty, means the shared_key is set, only works when serverAuth is true
	secretName                 string //the name of the secret for the collector to use
	loggingNS                  string //the namespace where the collector pods deployed in
	inPluginType               string //forward or http
}

func (f fluentdServer) createPipelineSecret(oc *exutil.CLI, keysPath string) {
	secret := resource{"secret", f.secretName, f.loggingNS}
	cmd := []string{"secret", "generic", secret.name, "-n", secret.namespace, "--from-file=ca-bundle.crt=" + keysPath + "/ca.crt"}
	if f.clientAuth {
		cmd = append(cmd, "--from-file=tls.key="+keysPath+"/client.key", "--from-file=tls.crt="+keysPath+"/client.crt")
	}
	if f.clientPrivateKeyPassphrase != "" {
		cmd = append(cmd, "--from-literal=passphrase="+f.clientPrivateKeyPassphrase)
	}
	if f.sharedKey != "" {
		cmd = append(cmd, "--from-literal=shared_key="+f.sharedKey)
	}
	err := oc.AsAdmin().WithoutNamespace().Run("create").Args(cmd...).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
	secret.WaitForResourceToAppear(oc)
}

func (f fluentdServer) deploy(oc *exutil.CLI) {
	// create SA
	sa := resource{"serviceaccount", f.serverName, f.namespace}
	err := oc.WithoutNamespace().Run("create").Args("serviceaccount", sa.name, "-n", sa.namespace).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
	sa.WaitForResourceToAppear(oc)
	//err = oc.AsAdmin().WithoutNamespace().Run("adm").Args("policy", "add-scc-to-user", "privileged", fmt.Sprintf("system:serviceaccount:%s:%s", f.namespace, f.serverName), "-n", f.namespace).Execute()
	//o.Expect(err).NotTo(o.HaveOccurred())
	filePath := []string{"testdata", "logging", "external-log-stores", "fluentd"}

	// create secrets if needed
	if f.serverAuth {
		o.Expect(f.secretName).NotTo(o.BeEmpty())
		filePath = append(filePath, "secure")
		// create a temporary directory
		baseDir := compat_otp.FixturePath("testdata", "logging")
		keysPath := filepath.Join(baseDir, "temp"+getRandomString())
		defer exec.Command("rm", "-r", keysPath).Output()
		err = os.MkdirAll(keysPath, 0755)
		o.Expect(err).NotTo(o.HaveOccurred())
		//generate certs
		cert := certsConf{f.serverName, f.namespace, f.clientPrivateKeyPassphrase}
		cert.generateCerts(oc, keysPath)
		//create pipelinesecret
		f.createPipelineSecret(oc, keysPath)
		//create secret for fluentd server
		err = oc.AsAdmin().WithoutNamespace().Run("create").Args("secret", "generic", f.serverName, "-n", f.namespace, "--from-file=ca-bundle.crt="+keysPath+"/ca.crt", "--from-file=tls.key="+keysPath+"/server.key", "--from-file=tls.crt="+keysPath+"/server.crt", "--from-file=ca.key="+keysPath+"/ca.key").Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

	} else {
		filePath = append(filePath, "insecure")
	}

	// create configmap/deployment/svc
	cm := resource{"configmap", f.serverName, f.namespace}
	//when prefix is http-, the fluentdserver using http inplugin.
	cmFilePrefix := ""
	if f.inPluginType == "http" {
		cmFilePrefix = "http-"
	}

	var cmFileName string
	if !f.serverAuth {
		cmFileName = cmFilePrefix + "configmap.yaml"
	} else {
		if f.clientAuth {
			if f.sharedKey != "" {
				cmFileName = "cm-mtls-share.yaml"
			} else {
				cmFileName = cmFilePrefix + "cm-mtls.yaml"
			}
		} else {
			if f.sharedKey != "" {
				cmFileName = "cm-serverauth-share.yaml"
			} else {
				cmFileName = cmFilePrefix + "cm-serverauth.yaml"
			}
		}
	}
	cmFilePath := append(filePath, cmFileName)
	cmFile := compat_otp.FixturePath(cmFilePath...)
	cCmCmd := []string{"-f", cmFile, "-n", f.namespace, "-p", "NAMESPACE=" + f.namespace, "-p", "NAME=" + f.serverName}
	if f.sharedKey != "" {
		cCmCmd = append(cCmCmd, "-p", "SHARED_KEY="+f.sharedKey)
	}
	err = cm.applyFromTemplate(oc, cCmCmd...)
	o.Expect(err).NotTo(o.HaveOccurred())

	deploy := resource{"deployment", f.serverName, f.namespace}
	deployFilePath := append(filePath, "deployment.yaml")
	deployFile := compat_otp.FixturePath(deployFilePath...)
	err = deploy.applyFromTemplate(oc, "-f", deployFile, "-n", f.namespace, "-p", "NAMESPACE="+f.namespace, "-p", "NAME="+f.serverName)
	o.Expect(err).NotTo(o.HaveOccurred())
	WaitForDeploymentPodsToBeReady(oc, f.namespace, f.serverName)

	err = oc.AsAdmin().WithoutNamespace().Run("expose").Args("-n", f.namespace, "deployment", f.serverName, "--name="+f.serverName).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
}

func (f fluentdServer) remove(oc *exutil.CLI) {
	resource{"serviceaccount", f.serverName, f.namespace}.clear(oc)
	if f.serverAuth {
		resource{"secret", f.serverName, f.namespace}.clear(oc)
		resource{"secret", f.secretName, f.loggingNS}.clear(oc)
	}
	resource{"configmap", f.serverName, f.namespace}.clear(oc)
	resource{"deployment", f.serverName, f.namespace}.clear(oc)
	resource{"svc", f.serverName, f.namespace}.clear(oc)
}

func (f fluentdServer) getPodName(oc *exutil.CLI) string {
	pods, err := oc.AdminKubeClient().CoreV1().Pods(f.namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "component=" + f.serverName})
	o.Expect(err).NotTo(o.HaveOccurred())
	var names []string
	for i := 0; i < len(pods.Items); i++ {
		names = append(names, pods.Items[i].Name)
	}
	return names[0]
}

// check the data in fluentd server
// filename is the name of a file you want to check
// expect true means you expect the file to exist, false means the file is not expected to exist
func (f fluentdServer) checkData(oc *exutil.CLI, expect bool, filename string) {
	cmd := "ls -l /fluentd/log/" + filename
	if expect {
		err := wait.PollUntilContextTimeout(context.Background(), 20*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
			stdout, err := e2eoutput.RunHostCmdWithRetries(f.namespace, f.getPodName(oc), cmd, 3*time.Second, 15*time.Second)
			if err != nil {
				if strings.Contains(err.Error(), "No such file or directory") {
					return false, nil
				}
				return false, err
			}
			return strings.Contains(stdout, filename), nil
		})
		compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("The %s doesn't exist", filename))
	} else {
		err := wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 60*time.Second, true, func(context.Context) (done bool, err error) {
			stdout, err := e2eoutput.RunHostCmdWithRetries(f.namespace, f.getPodName(oc), cmd, 3*time.Second, 15*time.Second)
			if err != nil {
				if strings.Contains(err.Error(), "No such file or directory") {
					return true, nil
				}
				return false, err
			}
			return strings.Contains(stdout, "No such file or directory"), nil
		})
		compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("The %s exists", filename))
	}

}

func GetDataFromKafkaConsumerPod(oc *exutil.CLI, kafkaNS, consumerPod string) ([]LogEntity, error) {
	e2e.Logf("get logs from kakfa consumerPod %s", consumerPod)
	var logs []LogEntity
	//wait up to 5 minutes for logs appear in consumer pod
	err := wait.PollUntilContextTimeout(context.Background(), 10*time.Second, 300*time.Second, true, func(context.Context) (done bool, err error) {
		output, err := oc.AsAdmin().WithoutNamespace().Run("logs").Args("-n", kafkaNS, consumerPod, "--since=30s", "--tail=30").Output()
		if err != nil {
			e2e.Logf("error when oc logs consumer pod, continue")
			return false, nil
		}
		for _, line := range strings.Split(strings.TrimSuffix(output, "\n"), "\n") {
			//exclude those kafka-consumer logs, for exampe:
			//[2024-11-09 07:25:47,953] WARN [Consumer clientId=consumer-console-consumer-99163-1, groupId=console-consumer-99163] Error while fetching metadata with correlation id 165
			//: {topic-logging-app=UNKNOWN_TOPIC_OR_PARTITION} (org.apache.kafka.clients.NetworkClient)
			r, _ := regexp.Compile(`^{"@timestamp":.*}`)
			if r.MatchString(line) {
				var log LogEntity
				err = json.Unmarshal([]byte(line), &log)
				if err != nil {
					continue
				}
				logs = append(logs, log)
			} else {
				continue
			}
		}
		if len(logs) > 0 {
			return true, nil
		} else {
			e2e.Logf("can not find logs in consumerPod %s, continue", consumerPod)
			return false, nil
		}
	})
	if err != nil {
		return logs, fmt.Errorf("can not find consumer logs in 3 minutes")
	}
	return logs, nil
}

func GetDataFromKafkaByNamespace(oc *exutil.CLI, kafkaNS, consumerPod, namespace string) ([]LogEntity, error) {
	data, err := GetDataFromKafkaConsumerPod(oc, kafkaNS, consumerPod)
	if err != nil {
		return nil, err
	}
	var logs []LogEntity
	for _, log := range data {
		if log.Kubernetes.NamespaceName == namespace {
			logs = append(logs, log)
		}
	}
	return logs, nil
}

type kafka struct {
	namespace      string
	kafkasvcName   string
	zoosvcName     string
	authtype       string //Name the kafka folders under testdata same as the authtype (options: plaintext-ssl, sasl-ssl, sasl-plaintext)
	pipelineSecret string //the name of the secret for collectors to use
	collectorType  string //must be specified when auth type is sasl-ssl/sasl-plaintext
	loggingNS      string //the namespace where the collector pods are deployed in
}

func (k kafka) deployZookeeper(oc *exutil.CLI) {
	zookeeperFilePath := compat_otp.FixturePath("testdata", "logging", "external-log-stores", "kafka", "zookeeper")
	//create zookeeper configmap/svc/StatefulSet
	configTemplate := filepath.Join(zookeeperFilePath, "configmap.yaml")
	if k.authtype == "plaintext-ssl" {
		configTemplate = filepath.Join(zookeeperFilePath, "configmap-ssl.yaml")
	}
	err := resource{"configmap", k.zoosvcName, k.namespace}.applyFromTemplate(oc, "-n", k.namespace, "-f", configTemplate, "-p", "NAME="+k.zoosvcName, "NAMESPACE="+k.namespace)
	o.Expect(err).NotTo(o.HaveOccurred())

	zoosvcFile := filepath.Join(zookeeperFilePath, "zookeeper-svc.yaml")
	zoosvc := resource{"Service", k.zoosvcName, k.namespace}
	err = zoosvc.applyFromTemplate(oc, "-n", k.namespace, "-f", zoosvcFile, "-p", "NAME="+k.zoosvcName, "-p", "NAMESPACE="+k.namespace)
	o.Expect(err).NotTo(o.HaveOccurred())

	zoosfsFile := filepath.Join(zookeeperFilePath, "zookeeper-statefulset.yaml")
	zoosfs := resource{"StatefulSet", k.zoosvcName, k.namespace}
	err = zoosfs.applyFromTemplate(oc, "-n", k.namespace, "-f", zoosfsFile, "-p", "NAME="+k.zoosvcName, "-p", "NAMESPACE="+k.namespace, "-p", "SERVICENAME="+zoosvc.name, "-p", "CM_NAME="+k.zoosvcName)
	o.Expect(err).NotTo(o.HaveOccurred())
	WaitForPodReadyByLabel(oc, k.namespace, "app="+k.zoosvcName)
}

func (k kafka) deployKafka(oc *exutil.CLI) {
	kafkaFilePath := compat_otp.FixturePath("testdata", "logging", "external-log-stores", "kafka")
	kafkaConfigmapTemplate := filepath.Join(kafkaFilePath, k.authtype, "kafka-configmap.yaml")
	consumerConfigmapTemplate := filepath.Join(kafkaFilePath, k.authtype, "consumer-configmap.yaml")

	var keysPath string
	if k.authtype == "sasl-ssl" || k.authtype == "plaintext-ssl" {
		baseDir := compat_otp.FixturePath("testdata", "logging")
		keysPath = filepath.Join(baseDir, "temp"+getRandomString())
		defer exec.Command("rm", "-r", keysPath).Output()
		err := os.MkdirAll(keysPath, 0755)
		o.Expect(err).NotTo(o.HaveOccurred())
		generateCertsSH := filepath.Join(kafkaFilePath, "cert_generation.sh")
		stdout, err := exec.Command("sh", generateCertsSH, keysPath, k.namespace).Output()
		if err != nil {
			e2e.Logf("error generating certs: %s", string(stdout))
			e2e.Failf("error generating certs: %v", err)
		}
		err = oc.AsAdmin().WithoutNamespace().Run("create").Args("secret", "generic", "kafka-cluster-cert", "-n", k.namespace, "--from-file=ca_bundle.jks="+keysPath+"/ca/ca_bundle.jks", "--from-file=cluster.jks="+keysPath+"/cluster/cluster.jks").Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
	}

	pipelineSecret := resource{"secret", k.pipelineSecret, k.loggingNS}
	kafkaClientCert := resource{"secret", "kafka-client-cert", k.namespace}
	//create kafka secrets and confimap
	cmdPipeline := []string{"secret", "generic", pipelineSecret.name, "-n", pipelineSecret.namespace}
	cmdClient := []string{"secret", "generic", kafkaClientCert.name, "-n", kafkaClientCert.namespace}
	switch k.authtype {
	case "sasl-plaintext":
		{
			cmdClient = append(cmdClient, "--from-literal=username=admin", "--from-literal=password=admin-secret")
			cmdPipeline = append(cmdPipeline, "--from-literal=username=admin", "--from-literal=password=admin-secret")
			if k.collectorType == "vector" {
				cmdPipeline = append(cmdPipeline, "--from-literal=sasl.enable=True", "--from-literal=sasl.mechanisms=PLAIN")
			}
		}
	case "sasl-ssl":
		{
			cmdClient = append(cmdClient, "--from-file=ca-bundle.jks="+keysPath+"/ca/ca_bundle.jks", "--from-file=ca-bundle.crt="+keysPath+"/ca/ca_bundle.crt", "--from-file=tls.crt="+keysPath+"/client/client.crt", "--from-file=tls.key="+keysPath+"/client/client.key", "--from-literal=username=admin", "--from-literal=password=admin-secret")
			cmdPipeline = append(cmdPipeline, "--from-file=ca-bundle.crt="+keysPath+"/ca/ca_bundle.crt", "--from-literal=username=admin", "--from-literal=password=admin-secret")
			switch k.collectorType {
			case "fluentd":
				{
					cmdPipeline = append(cmdPipeline, "--from-literal=sasl_over_ssl=true")
				}
			case "vector":
				{
					cmdPipeline = append(cmdPipeline, "--from-literal=sasl.enable=True", "--from-literal=sasl.mechanisms=PLAIN", "--from-file=tls.crt="+keysPath+"/client/client.crt", "--from-file=tls.key="+keysPath+"/client/client.key")
				}
			}
		}
	case "plaintext-ssl":
		{
			cmdClient = append(cmdClient, "--from-file=ca-bundle.jks="+keysPath+"/ca/ca_bundle.jks", "--from-file=ca-bundle.crt="+keysPath+"/ca/ca_bundle.crt", "--from-file=tls.crt="+keysPath+"/client/client.crt", "--from-file=tls.key="+keysPath+"/client/client.key")
			cmdPipeline = append(cmdPipeline, "--from-file=ca-bundle.crt="+keysPath+"/ca/ca_bundle.crt", "--from-file=tls.crt="+keysPath+"/client/client.crt", "--from-file=tls.key="+keysPath+"/client/client.key")
		}
	}
	err := oc.AsAdmin().WithoutNamespace().Run("create").Args(cmdClient...).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
	kafkaClientCert.WaitForResourceToAppear(oc)
	err = oc.AsAdmin().WithoutNamespace().Run("create").Args(cmdPipeline...).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
	pipelineSecret.WaitForResourceToAppear(oc)

	consumerConfigmap := resource{"configmap", "kafka-client", k.namespace}
	err = consumerConfigmap.applyFromTemplate(oc, "-n", k.namespace, "-f", consumerConfigmapTemplate, "-p", "NAME="+consumerConfigmap.name, "NAMESPACE="+consumerConfigmap.namespace)
	o.Expect(err).NotTo(o.HaveOccurred())

	kafkaConfigmap := resource{"configmap", k.kafkasvcName, k.namespace}
	err = kafkaConfigmap.applyFromTemplate(oc, "-n", k.namespace, "-f", kafkaConfigmapTemplate, "-p", "NAME="+kafkaConfigmap.name, "NAMESPACE="+kafkaConfigmap.namespace)
	o.Expect(err).NotTo(o.HaveOccurred())

	//create ClusterRole and ClusterRoleBinding
	rbacFile := filepath.Join(kafkaFilePath, "kafka-rbac.yaml")
	output, err := oc.AsAdmin().WithoutNamespace().Run("process").Args("-n", k.namespace, "-f", rbacFile, "-p", "NAMESPACE="+k.namespace).OutputToFile(getRandomString() + ".json")
	o.Expect(err).NotTo(o.HaveOccurred())
	oc.AsAdmin().WithoutNamespace().Run("apply").Args("-f", output, "-n", k.namespace).Execute()

	//create kafka svc
	svcFile := filepath.Join(kafkaFilePath, "kafka-svc.yaml")
	svc := resource{"Service", k.kafkasvcName, k.namespace}
	err = svc.applyFromTemplate(oc, "-f", svcFile, "-n", svc.namespace, "-p", "NAME="+svc.name, "NAMESPACE="+svc.namespace)
	o.Expect(err).NotTo(o.HaveOccurred())

	//create kafka StatefulSet
	sfsFile := filepath.Join(kafkaFilePath, k.authtype, "kafka-statefulset.yaml")
	sfs := resource{"StatefulSet", k.kafkasvcName, k.namespace}
	err = sfs.applyFromTemplate(oc, "-f", sfsFile, "-n", k.namespace, "-p", "NAME="+sfs.name, "-p", "NAMESPACE="+sfs.namespace, "-p", "CM_NAME="+k.kafkasvcName)
	o.Expect(err).NotTo(o.HaveOccurred())
	WaitForStatefulsetReady(oc, sfs.namespace, sfs.name)

	//create kafka-consumer deployment
	deployFile := filepath.Join(kafkaFilePath, k.authtype, "kafka-consumer-deployment.yaml")
	deploy := resource{"deployment", "kafka-consumer-" + k.authtype, k.namespace}
	err = deploy.applyFromTemplate(oc, "-f", deployFile, "-n", deploy.namespace, "-p", "NAMESPACE="+deploy.namespace, "NAME="+deploy.name, "CM_NAME=kafka-client")
	o.Expect(err).NotTo(o.HaveOccurred())
	WaitForDeploymentPodsToBeReady(oc, deploy.namespace, deploy.name)
}

func (k kafka) removeZookeeper(oc *exutil.CLI) {
	resource{"configmap", k.zoosvcName, k.namespace}.clear(oc)
	resource{"svc", k.zoosvcName, k.namespace}.clear(oc)
	resource{"statefulset", k.zoosvcName, k.namespace}.clear(oc)
}

func (k kafka) removeKafka(oc *exutil.CLI) {
	resource{"secret", "kafka-client-cert", k.namespace}.clear(oc)
	resource{"secret", "kafka-cluster-cert", k.namespace}.clear(oc)
	resource{"secret", k.pipelineSecret, k.loggingNS}.clear(oc)
	oc.AsAdmin().WithoutNamespace().Run("delete").Args("clusterrole/kafka-node-reader").Execute()
	oc.AsAdmin().WithoutNamespace().Run("delete").Args("clusterrolebinding/kafka-node-reader").Execute()
	resource{"configmap", k.kafkasvcName, k.namespace}.clear(oc)
	resource{"svc", k.kafkasvcName, k.namespace}.clear(oc)
	resource{"statefulset", k.kafkasvcName, k.namespace}.clear(oc)
	resource{"configmap", "kafka-client", k.namespace}.clear(oc)
	resource{"deployment", "kafka-consumer-" + k.authtype, k.namespace}.clear(oc)
}

// deploy amqstream instance, kafka user for predefined topics
// if amqstreams absent, deploy amqstream operator
func (amqi *amqInstance) deploy(oc *exutil.CLI) {
	e2e.Logf("deploy amq instance")
	//initialize kakfa vars
	if amqi.name == "" {
		amqi.name = "my-cluster"
	}
	if amqi.namespace == "" {
		e2e.Failf("error, please define a namespace for amqstream instance")
	}
	if amqi.user == "" {
		amqi.user = "my-user"
	}
	if amqi.topicPrefix == "" {
		amqi.topicPrefix = "topic-logging"
	}
	if amqi.instanceType == "" {
		amqi.instanceType = "kafka-sasl-cluster"
	}

	loggingBaseDir := compat_otp.FixturePath("testdata", "logging")
	operatorDeployed := false
	// Wait csv appears up to 3 minutes
	wait.PollUntilContextTimeout(context.Background(), 30*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
		output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "-n", "openshift-operators").Output()
		if err != nil {
			return false, err
		}
		if strings.Contains(output, "amqstreams") {
			operatorDeployed = true
			return true, nil
		}
		return false, nil
	})
	if !operatorDeployed {
		e2e.Logf("deploy amqstream operator")
		output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("operatorhub/cluster", `-ojsonpath='{.status.sources[?(@.name=="redhat-operators")].disabled}'`).Output()
		if err != nil {
			g.Skip("Can not detect the catalog source/redhat-operators status")
		}
		if output == "true" {
			g.Skip("catalog source/redhat-operators is disabled")
		}
		catsrc := CatalogSourceObjects{"amq-streams-2.x", "redhat-operators", "openshift-marketplace"}
		amqs := SubscriptionObjects{
			OperatorName:  "amq-streams-cluster-operator",
			Namespace:     amqi.namespace,
			PackageName:   "amq-streams",
			Subscription:  filepath.Join(loggingBaseDir, "subscription", "sub-template.yaml"),
			OperatorGroup: filepath.Join(loggingBaseDir, "subscription", "singlenamespace-og.yaml"),
			CatalogSource: catsrc,
		}
		amqs.SubscribeOperator(oc)
		if IsFipsEnabled(oc) {
			//disable FIPS_MODE due to "java.io.IOException: getPBEAlgorithmParameters failed: PBEWithHmacSHA256AndAES_256 AlgorithmParameters not available"
			err = oc.AsAdmin().WithoutNamespace().Run("patch").Args("sub/"+amqs.PackageName, "-n", amqs.Namespace, "-p", "{\"spec\": {\"config\": {\"env\": [{\"name\": \"FIPS_MODE\", \"value\": \"disabled\"}]}}}", "--type=merge").Execute()
			o.Expect(err).NotTo(o.HaveOccurred())
		}
	}
	// before creating kafka, check the existence of crd kafkas.kafka.strimzi.io
	checkResource(oc, true, true, "kafka.strimzi.io", []string{"crd", "kafkas.kafka.strimzi.io", "-ojsonpath={.spec.group}"})
	kafka := resource{"kafka", amqi.name, amqi.namespace}
	kafkaTemplate := filepath.Join(loggingBaseDir, "external-log-stores", "kafka", "amqstreams", amqi.instanceType+".yaml")
	kafka.applyFromTemplate(oc, "-n", kafka.namespace, "-f", kafkaTemplate, "-p", "NAME="+kafka.name)
	// wait for kafka cluster to be ready
	WaitForPodReadyByLabel(oc, kafka.namespace, "app.kubernetes.io/instance="+kafka.name)
	if amqi.instanceType == "kafka-sasl-cluster" {
		e2e.Logf("deploy kafka user")
		kafkaUser := resource{"kafkauser", amqi.user, amqi.namespace}
		kafkaUserTemplate := filepath.Join(loggingBaseDir, "external-log-stores", "kafka", "amqstreams", "kafka-sasl-user.yaml")
		kafkaUser.applyFromTemplate(oc, "-n", kafkaUser.namespace, "-f", kafkaUserTemplate, "-p", "NAME="+amqi.user, "-p", "KAFKA_NAME="+amqi.name, "-p", "TOPIC_PREFIX="+amqi.topicPrefix)
		// get user password from secret my-user
		err := wait.PollUntilContextTimeout(context.Background(), 30*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
			secrets, err := oc.AdminKubeClient().CoreV1().Secrets(kafkaUser.namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "app.kubernetes.io/instance=" + kafkaUser.name})
			if err != nil {
				e2e.Logf("failed to list secret, continue")
				return false, nil
			}
			count := len(secrets.Items)
			if count == 0 {
				e2e.Logf("canot not find the secret %s, continues", kafkaUser.name)
				return false, nil
			}
			return true, nil
		})
		compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("Can not find the kafka user Secret %s", amqi.user))

		e2e.Logf("set kafka user password")
		amqi.password, err = oc.NotShowInfo().AsAdmin().WithoutNamespace().Run("get").Args("secret", amqi.user, "-n", amqi.namespace, "-o", "jsonpath={.data.password}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		temp, err := base64.StdEncoding.DecodeString(amqi.password)
		o.Expect(err).NotTo(o.HaveOccurred())
		amqi.password = string(temp)

		// get extranal route of amqstream kafka
		e2e.Logf("get kafka route")
		amqi.route = getRouteAddress(oc, amqi.namespace, amqi.name+"-kafka-external-bootstrap")
		amqi.route = amqi.route + ":443"

		// get ca for route
		e2e.Logf("get kafka routeCA")
		amqi.routeCA, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("secret", amqi.name+"-cluster-ca-cert", "-n", amqi.namespace, "-o", `jsonpath={.data.ca\.crt}`).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		temp, err = base64.StdEncoding.DecodeString(amqi.routeCA)
		o.Expect(err).NotTo(o.HaveOccurred())
		amqi.routeCA = string(temp)
	}
	// get internal service URL of amqstream kafka
	amqi.service = amqi.name + "-kafka-bootstrap." + amqi.namespace + ".svc:9092"
	e2e.Logf("amqstream deployed")
}

// try best to delete resources which will block normal deletion
func (amqi *amqInstance) destroy(oc *exutil.CLI) {
	e2e.Logf("delete kakfa resources")
	oc.AsAdmin().WithoutNamespace().Run("delete").Args("job", "--all", "-n", amqi.namespace).Execute()
	oc.AsAdmin().WithoutNamespace().Run("delete").Args("kafkauser", "--all", "-n", amqi.namespace).Execute()
	oc.AsAdmin().WithoutNamespace().Run("delete").Args("kafkatopic", "--all", "-n", amqi.namespace).Execute()
	oc.AsAdmin().WithoutNamespace().Run("delete").Args("kafka", amqi.name, "-n", amqi.namespace).Execute()
}

// Create kafka topic, create consumer pod and return consumer pod name
// Note: the topic name must match the amq.topicPrefix
func (amqi amqInstance) createTopicAndConsumber(oc *exutil.CLI, topicName string) string {
	e2e.Logf("create kakfa topic %s and consume pod", topicName)
	if !strings.HasPrefix(topicName, amqi.topicPrefix) {
		e2e.Failf("error, the topic %s must has prefix %s", topicName, amqi.topicPrefix)
	}
	var (
		consumerPodName string
		loggingBaseDir  = compat_otp.FixturePath("testdata", "logging")
		topicTemplate   = filepath.Join(loggingBaseDir, "external-log-stores", "kafka", "amqstreams", "kafka-topic.yaml")
		topic           = resource{"Kafkatopic", topicName, amqi.namespace}
	)
	err := topic.applyFromTemplate(oc, "-n", topic.namespace, "-f", topicTemplate, "-p", "NAMESPACE="+topic.namespace, "-p", "NAME="+topic.name, "CLUSTER_NAME="+amqi.name)
	o.Expect(err).NotTo(o.HaveOccurred())

	if amqi.instanceType == "kafka-sasl-cluster" {
		//create consumers sasl.client.property
		truststorePassword, err := oc.NotShowInfo().AsAdmin().WithoutNamespace().Run("get").Args("secret", amqi.name+"-cluster-ca-cert", "-n", amqi.namespace, "-o", `jsonpath={.data.ca\.password}`).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		temp, err := base64.StdEncoding.DecodeString(truststorePassword)
		o.Expect(err).NotTo(o.HaveOccurred())
		truststorePassword = string(temp)
		consumerConfigTemplate := filepath.Join(loggingBaseDir, "external-log-stores", "kafka", "amqstreams", "kafka-sasl-consumers-config.yaml")
		consumerConfig := resource{"configmap", "client-property-" + amqi.user, amqi.namespace}
		err = consumerConfig.applyFromTemplate(oc.NotShowInfo(), "-n", consumerConfig.namespace, "-f", consumerConfigTemplate, "-p", "NAME="+consumerConfig.name, "-p", "USER="+amqi.user, "-p", "PASSWORD="+amqi.password, "-p", "TRUSTSTORE_PASSWORD="+truststorePassword, "-p", "KAFKA_NAME="+amqi.name)
		o.Expect(err).NotTo(o.HaveOccurred())

		//create consumer pod
		consumerTemplate := filepath.Join(loggingBaseDir, "external-log-stores", "kafka", "amqstreams", "kafka-sasl-consumer-job.yaml")
		consumer := resource{"job", topicName + "-consumer", amqi.namespace}
		err = consumer.applyFromTemplate(oc, "-n", consumer.namespace, "-f", consumerTemplate, "-p", "NAME="+consumer.name, "-p", "CLUSTER_NAME="+amqi.name, "-p", "TOPIC_NAME="+topicName, "-p", "CLIENT_CONFIGMAP_NAME="+consumerConfig.name, "-p", "CA_SECRET_NAME="+amqi.name+"-cluster-ca-cert")
		o.Expect(err).NotTo(o.HaveOccurred())
		WaitForPodReadyByLabel(oc, amqi.namespace, "job-name="+consumer.name)

		consumerPods, err := oc.AdminKubeClient().CoreV1().Pods(amqi.namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "job-name=" + topicName + "-consumer"})
		o.Expect(err).NotTo(o.HaveOccurred())
		consumerPodName = consumerPods.Items[0].Name

	}
	if amqi.instanceType == "kafka-no-auth-cluster" {
		//create consumer pod
		consumerTemplate := filepath.Join(loggingBaseDir, "external-log-stores", "kafka", "amqstreams", "kafka-no-auth-consumer-job.yaml")
		consumer := resource{"job", topicName + "-consumer", amqi.namespace}
		err = consumer.applyFromTemplate(oc, "-n", consumer.namespace, "-f", consumerTemplate, "-p", "NAME="+consumer.name, "-p", "CLUSTER_NAME="+amqi.name, "-p", "TOPIC_NAME="+topicName)
		o.Expect(err).NotTo(o.HaveOccurred())
		WaitForPodReadyByLabel(oc, amqi.namespace, "job-name="+consumer.name)

		consumerPods, err := oc.AdminKubeClient().CoreV1().Pods(amqi.namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "job-name=" + topicName + "-consumer"})
		o.Expect(err).NotTo(o.HaveOccurred())
		consumerPodName = consumerPods.Items[0].Name
	}
	if consumerPodName == "" {
		e2e.Logf("can not get comsumer pod for the topic %s", topicName)
	} else {
		e2e.Logf("found the comsumer pod %s ", consumerPodName)
	}
	return consumerPodName
}

func (s *splunkPodServer) checkLogs(query string) bool {
	e2e.Logf("Splunk query: %s", query)
	err := wait.PollUntilContextTimeout(context.Background(), 30*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
		searchID, err := s.requestSearchTask(query)
		if err != nil {
			e2e.Logf("error getting search ID: %v", err)
			return false, nil
		}
		searchResult, err := s.getSearchResult(searchID)
		if err != nil {
			e2e.Logf("hit error when querying logs with %s: %v, try next round", query, err)
			return false, nil
		}
		if searchResult == nil || len(searchResult.Results) == 0 {
			e2e.Logf("no logs found for the query: %s, try next round", query)
			return false, nil
		}
		e2e.Logf("found records for the query: %s", query)
		return true, nil
	})

	return err == nil
}

func (s *splunkPodServer) auditLogFound() bool {
	return s.checkLogs("log_type=audit|head 1")
}

func (s *splunkPodServer) anyLogFound() bool {
	for _, logType := range []string{"infrastructure", "application", "audit"} {
		if s.checkLogs("log_type=" + logType + "|head 1") {
			return true
		}
	}
	return false
}

func (s *splunkPodServer) allQueryFound(queries []string) bool {
	if len(queries) == 0 {
		queries = []string{
			"log_type=application|head 1",
			"log_type=\"infrastructure\" SYSTEMD_INVOCATION_ID |head 1",
			"log_type=\"infrastructure\" container_image|head 1",
			"log_type=\"audit\" .linux-audit.log|head 1",
			"log_type=\"audit\" .ovn-audit.log|head 1",
			"log_type=\"audit\" .k8s-audit.log|head 1",
			"log_type=\"audit\" .openshift-audit.log|head 1",
		}
	}
	//return false if any query fail
	foundAll := true
	for _, query := range queries {
		if !s.checkLogs(query) {
			foundAll = false
		}
	}
	return foundAll
}

func (s *splunkPodServer) allTypeLogsFound() bool {
	queries := []string{
		"log_type=\"infrastructure\" SYSTEMD_INVOCATION_ID |head 1",
		"log_type=\"infrastructure\" container_image|head 1",
		"log_type=application|head 1",
		"log_type=audit|head 1",
	}
	return s.allQueryFound(queries)
}

func (s *splunkPodServer) getSearchResult(searchID string) (*splunkSearchResult, error) {
	h := make(http.Header)
	h.Add("Content-Type", "application/json")
	h.Add(
		"Authorization",
		"Basic "+base64.StdEncoding.EncodeToString([]byte(s.adminUser+":"+s.adminPassword)),
	)
	params := url.Values{}
	params.Add("output_mode", "json")

	var searchResult *splunkSearchResult

	resp, err1 := doHTTPRequest(h, "https://"+s.splunkdRoute, "/services/search/jobs/"+searchID+"/results", params.Encode(), "GET", true, 5, nil, 200)
	if err1 != nil {
		return nil, fmt.Errorf("failed to get response: %v", err1)
	}

	err2 := json.Unmarshal(resp, &searchResult)
	if err2 != nil {
		return nil, fmt.Errorf("failed to unmarshal splunk response: %v", err2)
	}
	return searchResult, nil
}

func (s *splunkPodServer) searchLogs(query string) (*splunkSearchResult, error) {
	searchID, err := s.requestSearchTask(query)
	if err != nil {
		return nil, fmt.Errorf("error getting search ID: %v", err)
	}
	return s.getSearchResult(searchID)
}

func (s *splunkPodServer) requestSearchTask(query string) (string, error) {
	h := make(http.Header)
	h.Add("Content-Type", "application/json")
	h.Add(
		"Authorization",
		"Basic "+base64.StdEncoding.EncodeToString([]byte(s.adminUser+":"+s.adminPassword)),
	)
	params := url.Values{}
	params.Set("search", "search "+query)

	resp, err := doHTTPRequest(h, "https://"+s.splunkdRoute, "/services/search/jobs", "", "POST", true, 2, strings.NewReader(params.Encode()), 201)
	if err != nil {
		return "", err
	}

	resmap := splunkSearchResp{}
	err = xml.Unmarshal(resp, &resmap)
	if err != nil {
		return "", err
	}
	return resmap.Sid, nil
}

// Set the default values to the splunkPodServer Object
func (s *splunkPodServer) init() {
	s.adminUser = "admin"
	s.adminPassword = getRandomString()
	s.hecToken = uuid.New().String()
	//https://idelta.co.uk/generate-hec-tokens-with-python/,https://docs.splunk.com/Documentation/SplunkCloud/9.0.2209/Security/Passwordbestpracticesforadministrators
	s.serviceName = s.name + "-0"
	s.serviceURL = s.serviceName + "." + s.namespace + ".svc"
	if s.name == "" {
		s.name = "splunk-default"
	}
	//authType must be one of "http|tls_serveronly|tls_mutual"
	//Note: when authType==http, you can still access splunk via https://${splunk_route}
	if s.authType == "" {
		s.authType = "http"
	}
	if s.version == "" {
		s.version = "9.0"
	}

	//Exit if anyone of caFile, keyFile,CertFile is null
	if s.authType == "tls_clientauth" || s.authType == "tls_serveronly" {
		o.Expect(s.caFile == "").To(o.BeFalse())
		o.Expect(s.keyFile == "").To(o.BeFalse())
		o.Expect(s.certFile == "").To(o.BeFalse())
	}
}

func (s *splunkPodServer) deploy(oc *exutil.CLI) {
	// Get route URL of splunk service
	appDomain, err := getAppDomain(oc)
	o.Expect(err).NotTo(o.HaveOccurred())
	//splunkd route URL
	s.splunkdRoute = s.name + "-splunkd-" + s.namespace + "." + appDomain
	//splunkd hec URL
	s.hecRoute = s.name + "-hec-" + s.namespace + "." + appDomain
	s.webRoute = s.name + "-web-" + s.namespace + "." + appDomain

	err = oc.AsAdmin().WithoutNamespace().Run("adm").Args("policy", "add-scc-to-user", "nonroot", "-z", "default", "-n", s.namespace).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())

	// Create secret used by splunk
	switch s.authType {
	case "http":
		s.deployHTTPSplunk(oc)
	case "tls_clientauth":
		s.deployCustomCertClientForceSplunk(oc)
	case "tls_serveronly":
		s.deployCustomCertSplunk(oc)
	default:
		s.deployHTTPSplunk(oc)
	}
	//In general, it take 1 minutes to be started, here wait 30second before call  waitForStatefulsetReady
	time.Sleep(30 * time.Second)
	waitForStatefulsetReady(oc, s.namespace, s.name)
}

func (s *splunkPodServer) deployHTTPSplunk(oc *exutil.CLI) {
	filePath := compat_otp.FixturePath("testdata", "logging", "external-log-stores", "splunk")
	//Create secret for splunk Statefulset
	secretTemplate := filepath.Join(filePath, "secret_splunk_template.yaml")
	secret := resource{"secret", s.name, s.namespace}
	err := secret.applyFromTemplate(oc, "-f", secretTemplate, "-p", "NAME="+secret.name, "-p", "HEC_TOKEN="+s.hecToken, "-p", "PASSWORD="+s.adminPassword)
	o.Expect(err).NotTo(o.HaveOccurred())

	//create splunk StatefulSet
	statefulsetTemplate := filepath.Join(filePath, "statefulset_splunk-"+s.version+"_template.yaml")
	splunkSfs := resource{"StatefulSet", s.name, s.namespace}
	err = splunkSfs.applyFromTemplate(oc, "-f", statefulsetTemplate, "-p", "NAME="+s.name)
	o.Expect(err).NotTo(o.HaveOccurred())

	//create route for splunk service
	routeHecTemplate := filepath.Join(filePath, "route-edge_splunk_template.yaml")
	routeHec := resource{"route", s.name + "-hec", s.namespace}
	err = routeHec.applyFromTemplate(oc, "-f", routeHecTemplate, "-p", "NAME="+routeHec.name, "-p", "SERVICE_NAME="+s.serviceName, "-p", "PORT_NAME=http-hec", "-p", "ROUTE_HOST="+s.hecRoute)
	o.Expect(err).NotTo(o.HaveOccurred())

	routeSplunkdTemplate := filepath.Join(filePath, "route-passthrough_splunk_template.yaml")
	routeSplunkd := resource{"route", s.name + "-splunkd", s.namespace}
	err = routeSplunkd.applyFromTemplate(oc, "-f", routeSplunkdTemplate, "-p", "NAME="+routeSplunkd.name, "-p", "SERVICE_NAME="+s.serviceName, "-p", "PORT_NAME=https-splunkd", "-p", "ROUTE_HOST="+s.splunkdRoute)
	o.Expect(err).NotTo(o.HaveOccurred())
}

func (s *splunkPodServer) genHecPemFile(hecFile string) error {
	dat1, err := os.ReadFile(s.certFile)
	if err != nil {
		e2e.Logf("Can not find the certFile %s", s.certFile)
		return err
	}
	dat2, err := os.ReadFile(s.keyFile)
	if err != nil {
		e2e.Logf("Can not find the keyFile %s", s.keyFile)
		return err
	}
	dat3, err := os.ReadFile(s.caFile)
	if err != nil {
		e2e.Logf("Can not find the caFile %s", s.caFile)
		return err
	}

	buf := []byte{}
	buf = append(buf, dat1...)
	buf = append(buf, dat2...)
	buf = append(buf, dat3...)
	err = os.WriteFile(hecFile, buf, 0644)
	return err
}

func (s *splunkPodServer) deployCustomCertSplunk(oc *exutil.CLI) {
	//Create basic secret content for splunk Statefulset
	filePath := compat_otp.FixturePath("testdata", "logging", "external-log-stores", "splunk")
	secretTemplate := filepath.Join(filePath, "secret_tls_splunk_template.yaml")
	if s.passphrase != "" {
		secretTemplate = filepath.Join(filePath, "secret_tls_passphrase_splunk_template.yaml")
	}
	secret := resource{"secret", s.name, s.namespace}
	if s.passphrase != "" {
		err := secret.applyFromTemplate(oc, "-f", secretTemplate, "-p", "NAME="+secret.name, "-p", "HEC_TOKEN="+s.hecToken, "-p", "PASSWORD="+s.adminPassword, "-p", "PASSPHASE="+s.passphrase)
		o.Expect(err).NotTo(o.HaveOccurred())
	} else {
		err := secret.applyFromTemplate(oc, "-f", secretTemplate, "-p", "NAME="+secret.name, "-p", "HEC_TOKEN="+s.hecToken, "-p", "PASSWORD="+s.adminPassword)
		o.Expect(err).NotTo(o.HaveOccurred())
	}

	//HEC need all in one PEM file.
	hecPemFile := "/tmp/" + getRandomString() + "hecAllKeys.crt"
	defer os.Remove(hecPemFile)
	err := s.genHecPemFile(hecPemFile)
	o.Expect(err).NotTo(o.HaveOccurred())

	//The secret will be mounted into splunk pods and used in server.conf,inputs.conf
	args := []string{"data", "secret/" + secret.name, "-n", secret.namespace}
	args = append(args, "--from-file=hec.pem="+hecPemFile)
	args = append(args, "--from-file=ca.pem="+s.caFile)
	args = append(args, "--from-file=key.pem="+s.keyFile)
	args = append(args, "--from-file=cert.pem="+s.certFile)
	if s.passphrase != "" {
		args = append(args, "--from-literal=passphrase="+s.passphrase)
	}
	err = oc.AsAdmin().WithoutNamespace().Run("set").Args(args...).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())

	//create splunk StatefulSet
	statefulsetTemplate := filepath.Join(filePath, "statefulset_splunk-"+s.version+"_template.yaml")
	splunkSfs := resource{"StatefulSet", s.name, s.namespace}
	err = splunkSfs.applyFromTemplate(oc, "-f", statefulsetTemplate, "-p", "NAME="+splunkSfs.name)
	o.Expect(err).NotTo(o.HaveOccurred())

	//create route for splunk service
	routeHecTemplate := filepath.Join(filePath, "route-passthrough_splunk_template.yaml")
	routeHec := resource{"route", s.name + "-hec", s.namespace}
	err = routeHec.applyFromTemplate(oc, "-f", routeHecTemplate, "-p", "NAME="+routeHec.name, "-p", "SERVICE_NAME="+s.serviceName, "-p", "PORT_NAME=http-hec", "-p", "ROUTE_HOST="+s.hecRoute)
	o.Expect(err).NotTo(o.HaveOccurred())

	routeSplunkdTemplate := filepath.Join(filePath, "route-passthrough_splunk_template.yaml")
	routeSplunkd := resource{"route", s.name + "-splunkd", s.namespace}
	err = routeSplunkd.applyFromTemplate(oc, "-f", routeSplunkdTemplate, "-p", "NAME="+routeSplunkd.name, "-p", "SERVICE_NAME="+s.serviceName, "-p", "PORT_NAME=https-splunkd", "-p", "ROUTE_HOST="+s.splunkdRoute)
	o.Expect(err).NotTo(o.HaveOccurred())
}

func (s *splunkPodServer) deployCustomCertClientForceSplunk(oc *exutil.CLI) {
	//Create secret for splunk Statefulset
	filePath := compat_otp.FixturePath("testdata", "logging", "external-log-stores", "splunk")
	secretTemplate := filepath.Join(filePath, "secret_tls_splunk_template.yaml")
	if s.passphrase != "" {
		secretTemplate = filepath.Join(filePath, "secret_tls_passphrase_splunk_template.yaml")
	}
	secret := resource{"secret", s.name, s.namespace}
	if s.passphrase != "" {
		err := secret.applyFromTemplate(oc, "-f", secretTemplate, "-p", "NAME="+secret.name, "-p", "HEC_TOKEN="+s.hecToken, "-p", "PASSWORD="+s.adminPassword, "-p", "HEC_CLIENTAUTH=True", "-p", "PASSPHASE="+s.passphrase)
		o.Expect(err).NotTo(o.HaveOccurred())
	} else {
		err := secret.applyFromTemplate(oc, "-f", secretTemplate, "-p", "NAME="+secret.name, "-p", "HEC_TOKEN="+s.hecToken, "-p", "PASSWORD="+s.adminPassword, "-p", "HEC_CLIENTAUTH=True")
		o.Expect(err).NotTo(o.HaveOccurred())
	}

	//HEC need all in one PEM file.
	hecPemFile := "/tmp/" + getRandomString() + "hecAllKeys.crt"
	defer os.Remove(hecPemFile)
	err := s.genHecPemFile(hecPemFile)
	o.Expect(err).NotTo(o.HaveOccurred())

	//The secret will be mounted into splunk pods and used in server.conf,inputs.conf
	secretArgs := []string{"data", "secret/" + secret.name, "-n", secret.namespace}
	secretArgs = append(secretArgs, "--from-file=hec.pem="+hecPemFile)
	secretArgs = append(secretArgs, "--from-file=ca.pem="+s.caFile)
	secretArgs = append(secretArgs, "--from-file=key.pem="+s.keyFile)
	secretArgs = append(secretArgs, "--from-file=cert.pem="+s.certFile)
	if s.passphrase != "" {
		secretArgs = append(secretArgs, "--from-literal=passphrase="+s.passphrase)
	}
	err = oc.AsAdmin().WithoutNamespace().Run("set").Args(secretArgs...).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())

	//create splunk StatefulSet
	statefulsetTemplate := filepath.Join(filePath, "statefulset_splunk-"+s.version+"_template.yaml")
	splunkSfs := resource{"StatefulSet", s.name, s.namespace}
	err = splunkSfs.applyFromTemplate(oc, "-f", statefulsetTemplate, "-p", "NAME="+splunkSfs.name)
	o.Expect(err).NotTo(o.HaveOccurred())

	//create route for splunk service
	routeHecTemplate := filepath.Join(filePath, "route-passthrough_splunk_template.yaml")
	routeHec := resource{"route", s.name + "-hec", s.namespace}
	err = routeHec.applyFromTemplate(oc, "-f", routeHecTemplate, "-p", "NAME="+routeHec.name, "-p", "SERVICE_NAME="+s.serviceName, "-p", "PORT_NAME=http-hec", "-p", "ROUTE_HOST="+s.hecRoute)
	o.Expect(err).NotTo(o.HaveOccurred())

	routeSplunkdTemplate := filepath.Join(filePath, "route-passthrough_splunk_template.yaml")
	routeSplunkd := resource{"route", s.name + "-splunkd", s.namespace}
	err = routeSplunkd.applyFromTemplate(oc, "-f", routeSplunkdTemplate, "-p", "NAME="+routeSplunkd.name, "-p", "SERVICE_NAME="+s.serviceName, "-p", "PORT_NAME=https-splunkd", "-p", "ROUTE_HOST="+s.splunkdRoute)
	o.Expect(err).NotTo(o.HaveOccurred())
}

func (s *splunkPodServer) destroy(oc *exutil.CLI) {
	oc.AsAdmin().WithoutNamespace().Run("delete").Args("route", s.name+"-hec", "-n", s.namespace).Execute()
	oc.AsAdmin().WithoutNamespace().Run("delete").Args("route", s.name+"-splunkd", "-n", s.namespace).Execute()
	oc.AsAdmin().WithoutNamespace().Run("delete").Args("statefulset", s.name, "-n", s.namespace).Execute()
	oc.AsAdmin().WithoutNamespace().Run("delete").Args("secret", s.name, "-n", s.namespace).Execute()
	oc.AsAdmin().WithoutNamespace().Run("adm").Args("policy", "remove-scc-from-user", "nonroot", "-z", "default", "-n", s.namespace).Execute()
}

// createIndexes adds custom index(es) into splunk
func (s *splunkPodServer) createIndexes(oc *exutil.CLI, indexes ...string) error {
	splunkPod, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", s.namespace, "pod", "-l", "app.kubernetes.io/instance="+s.name, "-ojsonpath={.items[0].metadata.name}").Output()
	if err != nil {
		return fmt.Errorf("error getting splunk pod: %v", err)
	}
	for _, index := range indexes {
		// curl -k -u admin:gjc2t9jx  https://localhost:8089/servicesNS/admin/search/data/indexes -d name=devtutorial
		cmd := "curl -k -u admin:" + s.adminPassword + " https://localhost:8089/servicesNS/admin/search/data/indexes -d name=" + index
		stdout, err := oc.NotShowInfo().AsAdmin().WithoutNamespace().Run("exec").Args("-n", s.namespace, splunkPod, "--", "/bin/sh", "-x", "-c", cmd).Output()
		if err != nil {
			e2e.Logf("query output: %v", stdout)
			return fmt.Errorf("can't create index %s, error: %v", index, err)
		}
	}
	return nil
}

// Create the secret which is used in CLF
func (toSp *toSplunkSecret) create(oc *exutil.CLI) {
	secretArgs := []string{"secret", "generic", toSp.name, "-n", toSp.namespace}

	if toSp.hecToken != "" {
		secretArgs = append(secretArgs, "--from-literal=hecToken="+toSp.hecToken)
	}
	if toSp.caFile != "" {
		secretArgs = append(secretArgs, "--from-file=ca-bundle.crt="+toSp.caFile)
	}
	if toSp.keyFile != "" {
		secretArgs = append(secretArgs, "--from-file=tls.key="+toSp.keyFile)
	}
	if toSp.certFile != "" {
		secretArgs = append(secretArgs, "--from-file=tls.crt="+toSp.certFile)
	}
	if toSp.passphrase != "" {
		secretArgs = append(secretArgs, "--from-literal=passphrase="+toSp.passphrase)
	}
	err := oc.AsAdmin().WithoutNamespace().Run("create").Args(secretArgs...).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
}

func (toSp *toSplunkSecret) delete(oc *exutil.CLI) {
	s := resource{"secret", toSp.name, toSp.namespace}
	s.clear(oc)
}

type externalLoki struct {
	name      string
	namespace string
}

func (l externalLoki) deployLoki(oc *exutil.CLI) {
	//Create configmap for Loki
	cmTemplate := compat_otp.FixturePath("testdata", "logging", "external-log-stores", "loki", "loki-configmap.yaml")
	lokiCM := resource{"configmap", l.name, l.namespace}
	err := lokiCM.applyFromTemplate(oc, "-n", l.namespace, "-f", cmTemplate, "-p", "LOKINAMESPACE="+l.namespace, "-p", "LOKICMNAME="+l.name)
	o.Expect(err).NotTo(o.HaveOccurred())

	//Create Deployment for Loki
	deployTemplate := compat_otp.FixturePath("testdata", "logging", "external-log-stores", "loki", "loki-deployment.yaml")
	lokiDeploy := resource{"deployment", l.name, l.namespace}
	err = lokiDeploy.applyFromTemplate(oc, "-n", l.namespace, "-f", deployTemplate, "-p", "LOKISERVERNAME="+l.name, "-p", "LOKINAMESPACE="+l.namespace, "-p", "LOKICMNAME="+l.name)
	o.Expect(err).NotTo(o.HaveOccurred())

	//Expose Loki as a Service
	WaitForDeploymentPodsToBeReady(oc, l.namespace, l.name)
	err = oc.AsAdmin().WithoutNamespace().Run("expose").Args("-n", l.namespace, "deployment", l.name).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())

	// expose loki route
	err = oc.AsAdmin().WithoutNamespace().Run("expose").Args("-n", l.namespace, "svc", l.name).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
}

func (l externalLoki) remove(oc *exutil.CLI) {
	resource{"configmap", l.name, l.namespace}.clear(oc)
	resource{"deployment", l.name, l.namespace}.clear(oc)
	resource{"svc", l.name, l.namespace}.clear(oc)
	resource{"route", l.name, l.namespace}.clear(oc)
}

func GetExtLokiSecret() (string, string, error) {
	glokiUser := os.Getenv("GLOKIUSER")
	glokiPwd := os.Getenv("GLOKIPWD")
	if glokiUser == "" || glokiPwd == "" {
		return "", "", fmt.Errorf("GLOKIUSER or GLOKIPWD environment variable is not set")
	}
	return glokiUser, glokiPwd, nil
}
*/
