package specs

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	exutil "github.com/openshift/cluster-logging-operator/test/test-extension/util"
	"github.com/openshift/origin/test/extended/util/compat_otp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

var _ = g.Describe("[sig-openshift-logging] LOGGING Logging", func() {
	defer g.GinkgoRecover()
	var (
		oc      = compat_otp.NewCLI("log-accept", compat_otp.KubeConfigPath())
		CLO, LO exutil.SubscriptionObjects
	)

	g.BeforeEach(func() {
		compat_otp.SkipBaselineCaps(oc, "None")
		subTemplate := exutil.FixturePath("testdata", "subscription", "sub-template.yaml")
		CLO = exutil.SubscriptionObjects{
			OperatorName:       "cluster-logging-operator",
			Namespace:          exutil.CloNS,
			PackageName:        "cluster-logging",
			Subscription:       subTemplate,
			OperatorGroup:      exutil.FixturePath("testdata", "subscription", "allnamespace-og.yaml"),
			SkipCaseWhenFailed: true,
		}
		LO = exutil.SubscriptionObjects{
			OperatorName:       "loki-operator-controller-manager",
			Namespace:          exutil.LoNS,
			PackageName:        "loki-operator",
			Subscription:       subTemplate,
			OperatorGroup:      exutil.FixturePath("testdata", "subscription", "allnamespace-og.yaml"),
			SkipCaseWhenFailed: true,
		}

		g.By("deploy CLO")
		CLO.SubscribeOperator(oc)
		oc.SetupProject()
	})

	// author qitang@redhat.com
	g.It("Author:qitang-CPaasrunBoth-Critical-74397-[InterOps] Forward logs to LokiStack.[Slow][Serial]", func() {
		g.By("deploy LO")
		LO.SubscribeOperator(oc)
		s := exutil.GetStorageType(oc)
		sc, err := exutil.GetStorageClassName(oc)
		if err != nil || len(sc) == 0 {
			g.Skip("can't get storageclass from cluster, skip this case")
		}

		appProj := oc.Namespace()
		jsonLogFile := exutil.FixturePath("testdata", "generatelog", "container_json_log_template.json")
		err = oc.WithoutNamespace().Run("new-app").Args("-n", appProj, "-f", jsonLogFile).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		if !exutil.HasMaster(oc) {
			nodeName, err := exutil.GenLinuxAuditLogsOnWorker(oc)
			o.Expect(err).NotTo(o.HaveOccurred())
			defer exutil.DeleteLinuxAuditPolicyFromNode(oc, nodeName)
		}

		g.By("deploy loki stack")
		lokiStackTemplate := exutil.FixturePath("testdata", "lokistack", "lokistack-simple.yaml")
		ls := exutil.LokiStack{
			Name:          "loki-74397",
			Namespace:     exutil.LoggingNS,
			TSize:         "1x.demo",
			StorageType:   s,
			StorageSecret: "storage-secret-74397",
			StorageClass:  sc,
			BucketName:    "logging-loki-74397-" + exutil.GetInfrastructureName(oc),
			Template:      lokiStackTemplate,
		}
		defer ls.RemoveObjectStorage(oc)
		err = ls.PrepareResourcesForLokiStack(oc)
		o.Expect(err).NotTo(o.HaveOccurred())
		defer ls.RemoveLokiStack(oc)
		err = ls.DeployLokiStack(oc)
		o.Expect(err).NotTo(o.HaveOccurred())
		ls.WaitForLokiStackToBeReady(oc)

		compat_otp.By("deploy logfilesmetricexporter")
		lfme := exutil.LogFileMetricExporter{
			Name:          "instance",
			Namespace:     exutil.LoggingNS,
			Template:      exutil.FixturePath("testdata", "logfilemetricexporter", "lfme.yaml"),
			WaitPodsReady: true,
		}
		defer lfme.Delete(oc)
		lfme.Create(oc)

		compat_otp.By("create a CLF to test forward to lokistack")
		clf := exutil.Clusterlogforwarder{
			Name:                      "clf-74397",
			Namespace:                 exutil.LoggingNS,
			ServiceAccountName:        "logcollector-74397",
			TemplateFile:              exutil.FixturePath("testdata", "observability.openshift.io_clusterlogforwarder", "lokistack.yaml"),
			SecretName:                "lokistack-secret-74397",
			CollectApplicationLogs:    true,
			CollectAuditLogs:          true,
			CollectInfrastructureLogs: true,
			WaitForPodReady:           true,
			EnableMonitoring:          true,
		}
		clf.CreateServiceAccount(oc)
		defer exutil.RemoveClusterRoleFromServiceAccount(oc, clf.Namespace, clf.ServiceAccountName, "logging-collector-logs-writer")
		err = exutil.AddClusterRoleToServiceAccount(oc, clf.Namespace, clf.ServiceAccountName, "logging-collector-logs-writer")
		o.Expect(err).NotTo(o.HaveOccurred())
		defer exutil.DeleteResourceFromCluster(oc, "secret", clf.SecretName, clf.Namespace)
		ls.CreateSecretFromGateway(oc, clf.SecretName, clf.Namespace, "")
		defer clf.Delete(oc)
		clf.Create(oc, "LOKISTACK_NAME="+ls.Name, "LOKISTACK_NAMESPACE="+ls.Namespace)

		//check logs in loki stack
		g.By("check logs in loki")
		defer exutil.RemoveClusterRoleFromServiceAccount(oc, oc.Namespace(), "default", "cluster-admin")
		err = exutil.AddClusterRoleToServiceAccount(oc, oc.Namespace(), "default", "cluster-admin")
		o.Expect(err).NotTo(o.HaveOccurred())
		bearerToken := exutil.GetSAToken(oc, "default", oc.Namespace())
		route := "https://" + exutil.GetRouteAddress(oc, ls.Namespace, ls.Name)
		lc := exutil.NewLokiClient(route).WithToken(bearerToken).Retry(5)
		for _, logType := range []string{"application", "infrastructure", "audit"} {
			lc.WaitForLogsAppearByKey(logType, "log_type", logType)
			labels, err := lc.ListLabels(logType, "")
			o.Expect(err).NotTo(o.HaveOccurred(), "got error when checking %s log labels", logType)
			e2e.Logf("\nthe %s log labels are: %v\n", logType, labels)
		}
		journalLog, err := lc.SearchLogsInLoki("infrastructure", `{log_type = "infrastructure", kubernetes_namespace_name !~ ".+"}`)
		o.Expect(err).NotTo(o.HaveOccurred())
		journalLogs := exutil.ExtractLogEntities(journalLog)
		o.Expect(len(journalLogs) > 0).Should(o.BeTrue(), "can't find journal logs in lokistack")
		e2e.Logf("find journal logs")
		lc.WaitForLogsAppearByProject("application", appProj)

		compat_otp.By("Checking for stream labels following OTEL Semantic Conventions as a forward compatibility")
		// For App tenant
		lc.WaitForLogsAppearByKey("application", "k8s_namespace_name", appProj)
		lc.WaitForLogsAppearByKey("application", "k8s_container_name", "logging-centos-logtest")
		// For Infra tenant
		logs, err := lc.SearchLogsInLoki("infrastructure", `{openshift_log_type = "infrastructure", k8s_namespace_name=~".+"}`)
		o.Expect(err).NotTo(o.HaveOccurred())
		extractedLogs := exutil.ExtractLogEntities(logs)
		o.Expect(len(extractedLogs) > 0).Should(o.BeTrue())
		journalLog, err = lc.SearchLogsInLoki("infrastructure", `{openshift_log_type = "infrastructure", k8s_namespace_name !~ ".+"}`)
		o.Expect(err).NotTo(o.HaveOccurred())
		journalLogs = exutil.ExtractLogEntities(journalLog)
		o.Expect(len(journalLogs) > 0).Should(o.BeTrue())
		// For Audit tenant
		lc.WaitForLogsAppearByKey("audit", "openshift_log_type", "audit")
		e2e.Logf("Otel semantic convention labels found!")

		g.By("Check if the ServiceMonitor object for Vector is created.")
		o.Expect(exutil.WaitForResourceToAppear(oc, "servicemonitor", clf.Name, clf.Namespace)).NotTo(o.HaveOccurred())

		promToken := exutil.GetSAToken(oc, "prometheus-k8s", "openshift-monitoring")
		g.By("check metrics exposed by collector")
		for _, job := range []string{clf.Name, "logfilesmetricexporter"} {
			exutil.CheckMetric(oc, promToken, "{job=\""+job+"\"}", 3)
		}
		for _, metric := range []string{"log_logged_bytes_total", "vector_component_received_events_total"} {
			exutil.CheckMetric(oc, promToken, metric, 3)
		}

		g.By("check metrics exposed by loki")
		svcs, err := oc.AdminKubeClient().CoreV1().Services(ls.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "app.kubernetes.io/created-by=lokistack-controller"})
		o.Expect(err).NotTo(o.HaveOccurred())
		for _, svc := range svcs.Items {
			if !strings.Contains(svc.Name, "grpc") && !strings.Contains(svc.Name, "ring") {
				exutil.CheckMetric(oc, promToken, "{job=\""+svc.Name+"\"}", 3)
			}
		}
		for _, metric := range []string{"loki_boltdb_shipper_compactor_running", "loki_distributor_bytes_received_total", "loki_inflight_requests", "workqueue_work_duration_seconds_bucket{namespace=\"" + exutil.LoNS + "\", job=\"loki-operator-controller-manager-metrics-service\"}", "loki_build_info", "loki_ingester_streams_created_total"} {
			exutil.CheckMetric(oc, promToken, metric, 3)
		}
		compat_otp.By("Validate log streams are pushed to external storage bucket/container")
		ls.ValidateExternalObjectStorageForLogs(oc, []string{"application", "audit", "infrastructure"})
	})

	g.It("Author:qitang-CPaasrunBoth-ConnectedOnly-Critical-74926-[InterOps] Forward logs to Cloudwatch.", func() {
		clfNS := oc.Namespace()
		cw := exutil.CloudwatchSpec{
			CollectorSAName: "cloudwatch-" + exutil.GetRandomString(),
			GroupName:       "logging-74926-" + exutil.GetInfrastructureName(oc) + `.{.log_type||"none-typed-logs"}`,
			LogTypes:        []string{"infrastructure", "application", "audit"},
			SecretNamespace: clfNS,
			SecretName:      "logging-74926-" + exutil.GetRandomString(),
		}
		cw.Init(oc)
		defer cw.DeleteResources(oc)

		g.By("Create log producer")
		appProj := oc.Namespace()
		jsonLogFile := exutil.FixturePath("testdata", "generatelog", "container_json_log_template.json")
		err := oc.WithoutNamespace().Run("new-app").Args("-n", appProj, "-f", jsonLogFile).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		if !cw.HasMaster {
			nodeName, err := exutil.GenLinuxAuditLogsOnWorker(oc)
			o.Expect(err).NotTo(o.HaveOccurred())
			defer exutil.DeleteLinuxAuditPolicyFromNode(oc, nodeName)
		}

		g.By("Create clusterlogforwarder")
		var template string
		if cw.StsEnabled {
			template = exutil.FixturePath("testdata", "observability.openshift.io_clusterlogforwarder", "cloudwatch-iamRole.yaml")
		} else {
			template = exutil.FixturePath("testdata", "observability.openshift.io_clusterlogforwarder", "cloudwatch-accessKey.yaml")
		}
		clf := exutil.Clusterlogforwarder{
			Name:                      "clf-74926",
			Namespace:                 clfNS,
			SecretName:                cw.SecretName,
			TemplateFile:              template,
			WaitForPodReady:           true,
			CollectApplicationLogs:    true,
			CollectAuditLogs:          true,
			CollectInfrastructureLogs: true,
			EnableMonitoring:          true,
			ServiceAccountName:        cw.CollectorSAName,
		}
		defer clf.Delete(oc)
		clf.CreateServiceAccount(oc)
		cw.CreateClfSecret(oc)
		clf.Create(oc, "REGION="+cw.Region, "GROUP_NAME="+cw.GroupName, `TUNING={"compression": "snappy", "deliveryMode": "AtMostOnce", "maxRetryDuration": 20, "maxWrite": "10M", "minRetryDuration": 5}`)

		nodes, err := clf.GetCollectorNodeNames(oc)
		o.Expect(err).NotTo(o.HaveOccurred())
		cw.Nodes = append(cw.Nodes, nodes...)

		g.By("Check logs in Cloudwatch")
		o.Expect(cw.LogsFound()).To(o.BeTrue())

		compat_otp.By("check tuning in collector configurations")
		expectedConfigs := []string{
			`compression = "snappy"`,
			`[sinks.output_cloudwatch.batch]
max_bytes = 10000000`,
			`[sinks.output_cloudwatch.buffer]
when_full = "drop_newest"`,
			`[sinks.output_cloudwatch.request]
retry_initial_backoff_secs = 5
retry_max_duration_secs = 20`,
		}
		result, err := exutil.CheckCollectorConfiguration(oc, clf.Namespace, clf.Name+"-config", expectedConfigs...)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(result).Should(o.BeTrue())
	})

	//author qitang@redhat.com
	g.It("Author:qitang-CPaasrunBoth-ConnectedOnly-Critical-74924-Forward logs to GCL", func() {
		projectID, err := exutil.GetGCPProjectID(oc)
		o.Expect(err).NotTo(o.HaveOccurred())
		gcl := exutil.GoogleCloudLogging{
			ProjectID: projectID,
			LogName:   exutil.GetInfrastructureName(oc) + "-74924",
		}
		defer gcl.RemoveLogs()

		g.By("Create log producer")
		appProj := oc.Namespace()
		jsonLogFile := exutil.FixturePath("testdata", "generatelog", "container_json_log_template.json")
		err = oc.WithoutNamespace().Run("new-app").Args("-n", appProj, "-f", jsonLogFile).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		oc.SetupProject()
		clfNS := oc.Namespace()
		defer exutil.DeleteResourceFromCluster(oc, "secret", "gcp-secret-74924", clfNS)
		err = exutil.CreateSecretForGCL(oc, "gcp-secret-74924", clfNS)
		o.Expect(err).NotTo(o.HaveOccurred())

		clf := exutil.Clusterlogforwarder{
			Name:                      "clf-74924",
			Namespace:                 clfNS,
			SecretName:                "gcp-secret-74924",
			TemplateFile:              exutil.FixturePath("testdata", "observability.openshift.io_clusterlogforwarder", "googleCloudLogging.yaml"),
			WaitForPodReady:           true,
			CollectApplicationLogs:    true,
			CollectAuditLogs:          true,
			CollectInfrastructureLogs: true,
			ServiceAccountName:        "test-clf-" + exutil.GetRandomString(),
		}
		defer clf.Delete(oc)
		clf.Create(oc, "ID_TYPE=project", "ID_VALUE="+gcl.ProjectID, "LOG_ID="+gcl.LogName)

		for _, logType := range []string{"infrastructure", "audit", "application"} {
			err = wait.PollUntilContextTimeout(context.Background(), 30*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
				logs, err := gcl.GetLogByType(logType)
				if err != nil {
					return false, err
				}
				return len(logs) > 0, nil
			})
			compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("%s logs are not found", logType))
		}
		err = gcl.WaitForLogsAppearByNamespace(appProj)
		compat_otp.AssertWaitPollNoErr(err, "can't find app logs from project/"+appProj)

		// Check tuning options for GCL under collector configMap
		expectedConfigs := []string{"[sinks.output_gcp_logging.batch]", "[sinks.output_gcp_logging.buffer]", "[sinks.output_gcp_logging.request]", "retry_initial_backoff_secs = 10", "retry_max_duration_secs = 20"}
		result, err := exutil.CheckCollectorConfiguration(oc, clf.Namespace, clf.Name+"-config", expectedConfigs...)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(result).Should(o.BeTrue())
	})

	//author anli@redhat.com
	g.It("Author:anli-CPaasrunBoth-ConnectedOnly-Critical-71772-Forward logs to AZMonitor -- full options", func() {
		platform := compat_otp.CheckPlatform(oc)
		if platform == "azure" && compat_otp.IsWorkloadIdentityCluster(oc) {
			g.Skip("Skip on the workload identity enabled cluster!")
		}
		var (
			resourceGroupName string
			location          string
		)
		infraName := exutil.GetInfrastructureName(oc)
		if platform != "azure" {
			if !exutil.ReadAzureCredentials() {
				g.Skip("Skip for the platform is not Azure and can't get credentials from env vars.")
			}
			resourceGroupName = infraName + "-logging-71772-rg"
			azureSubscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
			cred := exutil.CreateNewDefaultAzureCredential()
			location = "westus" //TODO: define default location

			_, err := exutil.CreateAzureResourceGroup(resourceGroupName, azureSubscriptionID, location, cred)
			defer exutil.DeleteAzureResourceGroup(resourceGroupName, azureSubscriptionID, cred)
			if err != nil {
				g.Skip("Failed to create azure resource group: " + err.Error() + ", skip the case.")
			}
			e2e.Logf("Successfully created resource group %s", resourceGroupName)
		} else {
			cloudName := exutil.GetAzureCloudName(oc)
			if !(cloudName == "azurepubliccloud" || cloudName == "azureusgovernmentcloud") {
				g.Skip("The case can only be running on Azure Public and Azure US Goverment now!")
			}
			resourceGroupName, _ = compat_otp.GetAzureCredentialFromCluster(oc)
		}

		g.By("Prepre Azure Log Storage Env")
		workSpaceName := infraName + "case71772"
		azLog, err := exutil.NewAzureLog(oc, location, resourceGroupName, workSpaceName, "case71772")
		defer azLog.DeleteWorkspace()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Create log producer")
		clfNS := oc.Namespace()
		jsonLogFile := exutil.FixturePath("testdata", "generatelog", "container_json_log_template.json")
		err = oc.WithoutNamespace().Run("new-app").Args("-n", clfNS, "-f", jsonLogFile).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Deploy CLF to send logs to Log Analytics")
		defer exutil.DeleteResourceFromCluster(oc, "secret", "azure-secret-71772", clfNS)
		err = azLog.CreateSecret(oc, "azure-secret-71772", clfNS)
		o.Expect(err).NotTo(o.HaveOccurred())
		clf := exutil.Clusterlogforwarder{
			Name:                      "clf-71772",
			Namespace:                 clfNS,
			SecretName:                "azure-secret-71772",
			TemplateFile:              exutil.FixturePath("testdata", "observability.openshift.io_clusterlogforwarder", "azureMonitor.yaml"),
			WaitForPodReady:           true,
			CollectApplicationLogs:    true,
			CollectAuditLogs:          true,
			CollectInfrastructureLogs: true,
			ServiceAccountName:        "test-clf-" + exutil.GetRandomString(),
		}
		defer clf.Delete(oc)
		clf.Create(oc, "PREFIX_OR_NAME="+azLog.PrefixOrName, "CUSTOMER_ID="+azLog.CustomerID, "RESOURCE_ID="+azLog.WorkspaceID, "AZURE_HOST="+azLog.Host)

		g.By("Verify the test result")
		for _, tableName := range []string{azLog.PrefixOrName + "infra_log_CL", azLog.PrefixOrName + "audit_log_CL", azLog.PrefixOrName + "app_log_CL"} {
			_, err := azLog.GetLogByTable(tableName)
			compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("can't find logs from %s in AzureLogWorkspace", tableName))
		}
	})

})
