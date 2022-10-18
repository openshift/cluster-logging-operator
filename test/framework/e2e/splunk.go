package e2e

import (
	"context"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net/url"

	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SplunkStandalone = "splunk-standalone"
	Splunk           = "splunk"
	splunkImage      = "quay.io/openshift-logging/splunk:9.0.0"
	// #nosec G101
	splunkSecret     = "splunk-hec-secret"
	SplunkHecService = "splunk-hec-service"
	splunkHecPort    = 8088
)

var (
	labels = map[string]string{
		"app.kubernetes.io/component":  SplunkStandalone,
		"app.kubernetes.io/instance":   Splunk,
		"app.kubernetes.io/managed-by": constants.OpenshiftNS,
		"app.kubernetes.io/name":       SplunkStandalone,
		"app.kubernetes.io/part-of":    Splunk,
	}

	HecToken      = utils.GetRandomWord(16)
	AdminPassword = utils.GetRandomWord(16)

	SplunkEndpoint = url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", SplunkHecService, splunkHecPort),
	}
)

func (tc *E2ETestFramework) DeploySplunk() (*apps.StatefulSet, error) {
	if err := tc.createSplunkSecret(); err != nil {
		return nil, err
	}

	app, err := tc.createSplunkStatefulSet()
	if err != nil {
		return nil, err
	}

	if err = tc.createSplunkHecService(); err != nil {
		return nil, err
	}
	return app, nil
}

func (tc *E2ETestFramework) createSplunkStatefulSet() (*apps.StatefulSet, error) {
	//need for special Splunk user
	_, err := oc.Literal().From("oc adm policy add-scc-to-user nonroot -z default -n " + constants.OpenshiftNS).Run()
	if err != nil {
		return nil, err
	}
	app := newSplunkStatefulSet()
	tc.AddCleanup(func() error {
		var zerograce int64
		opts := metav1.DeleteOptions{
			GracePeriodSeconds: &zerograce,
		}
		return tc.KubeClient.AppsV1().StatefulSets(constants.OpenshiftNS).Delete(context.TODO(), app.GetName(), opts)
	})

	app, err = tc.KubeClient.AppsV1().StatefulSets(constants.OpenshiftNS).Create(context.TODO(), app, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	if err := tc.waitForStatefulSet(constants.OpenshiftNS, app.GetName(), defaultRetryInterval, defaultTimeout); err != nil {
		return nil, err
	}

	return app, nil
}

func (tc *E2ETestFramework) createSplunkHecService() error {
	svc := newSplunkHecService()

	tc.AddCleanup(func() error {
		var zerograce int64
		opts := metav1.DeleteOptions{
			GracePeriodSeconds: &zerograce,
		}
		return tc.KubeClient.CoreV1().Services(constants.OpenshiftNS).Delete(context.TODO(), svc.GetName(), opts)
	})

	if _, err := tc.KubeClient.CoreV1().Services(constants.OpenshiftNS).Create(context.TODO(), svc, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

func (tc *E2ETestFramework) createSplunkSecret() error {
	s := newSplunkSecret()

	tc.AddCleanup(func() error {
		var zerograce int64
		opts := metav1.DeleteOptions{
			GracePeriodSeconds: &zerograce,
		}
		return tc.KubeClient.CoreV1().Secrets(constants.OpenshiftNS).Delete(context.TODO(), s.GetName(), opts)
	})

	if _, err := tc.KubeClient.CoreV1().Secrets(constants.OpenshiftNS).Create(context.TODO(), s, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

func newSplunkStatefulSet() *apps.StatefulSet {
	var (
		replicas    int32 = 1
		termination int64 = 30
	)

	return &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Splunk,
			Namespace: constants.OpenshiftNS,
			Labels: map[string]string{
				"app": Splunk,
			},
		},
		Spec: apps.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Replicas:    &replicas,
			ServiceName: "splunk-hec-headless",
			UpdateStrategy: apps.StatefulSetUpdateStrategy{
				Type: apps.RollingUpdateStatefulSetStrategyType,
			},
			PodManagementPolicy: "Parallel",
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					TerminationGracePeriodSeconds: &termination,
					Volumes: []v1.Volume{
						{
							Name: "mnt-splunk-secrets",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: splunkSecret,
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:  Splunk,
							Image: splunkImage,
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									Exec: &v1.ExecAction{
										Command: []string{
											"/bin/grep",
											"started",
											"/opt/container_artifact/splunk-container.state",
										},
									},
								},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      5,
								PeriodSeconds:       5,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							TerminationMessagePath: "/dev/termination-log",
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									Exec: &v1.ExecAction{
										Command: []string{
											"/sbin/checkstate.sh",
										},
									},
								},
								InitialDelaySeconds: 300,
								TimeoutSeconds:      30,
								PeriodSeconds:       30,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							Env: []v1.EnvVar{
								{
									Name:  "SPLUNK_DECLARATIVE_ADMIN_PASSWORD",
									Value: "true",
								},
								{
									Name:  "SPLUNK_DEFAULTS_URL",
									Value: "/mnt/splunk-secrets/default.yml",
								},
								{
									Name:  "SPLUNK_HOME",
									Value: "/opt/splunk",
								},
								{
									Name:  "SPLUNK_HOME_OWNERSHIP_ENFORCEMENT",
									Value: "false",
								},
								{
									Name:  "SPLUNK_ROLE",
									Value: "splunk_standalone",
								},
								{
									Name:  "SPLUNK_START_ARGS",
									Value: "--accept-license",
								},
							},
							Ports: []v1.ContainerPort{
								{
									Name:          "http-splunkweb",
									ContainerPort: 8000,
								},
								{
									Name:          "http-hec",
									ContainerPort: splunkHecPort,
								},
								{
									Name:          "https-splunkd",
									ContainerPort: 8089,
								},
								{
									Name:          "tcp-s2s",
									ContainerPort: 9997,
								},
							},
							ImagePullPolicy: v1.PullIfNotPresent,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "mnt-splunk-secrets",
									MountPath: "/mnt/splunk-secrets",
								},
								{
									Name:      "pvc-etc",
									MountPath: "/opt/splunk/etc",
								},
								{
									Name:      "pvc-var",
									MountPath: "/opt/splunk/var",
								},
							},
						},
					},
					SecurityContext: &v1.PodSecurityContext{
						RunAsUser:    utils.GetInt64(41812),
						RunAsNonRoot: utils.GetBool(true),
						FSGroup:      utils.GetInt64(41812),
					},
				},
			},
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pvc-etc",
						Namespace: constants.OpenshiftNS,
						Labels:    labels,
					},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{
							v1.ReadWriteOnce,
						},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("10Gi"),
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pvc-var",
						Namespace: constants.OpenshiftNS,
						Labels:    labels,
					},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{
							v1.ReadWriteOnce,
						},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("10Gi"),
							},
						},
					},
				},
			},
		},
	}

}

func newSplunkHecService() *v1.Service {
	ports := []v1.ServicePort{
		{
			Name:       "http-splunkweb",
			Port:       8000,
			TargetPort: intstr.FromInt(8000),
		},
		{
			Name:       "http-hec",
			Port:       splunkHecPort,
			TargetPort: intstr.FromInt(splunkHecPort),
		},
		{
			Name:       "https-splunkd",
			Port:       8089,
			TargetPort: intstr.FromInt(8089),
		},
		{
			Name:       "tcp-s2s",
			Port:       9097,
			TargetPort: intstr.FromInt(9097),
		},
	}

	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      SplunkHecService,
			Namespace: constants.OpenshiftNS,
			Labels:    labels,
		},
		Spec: v1.ServiceSpec{
			Selector: labels,
			Ports:    ports,
		},
	}
}

func newSplunkSecret() *v1.Secret {
	data := map[string][]byte{
		"default.yml": []byte("" +
			"splunk:\n" +
			"  hec:\n" +
			"    ssl: false\n" +
			"    token: \"" + string(HecToken) + "\"\n" +
			"  password: \"" + string(AdminPassword) + "\"\n" +
			"  pass4SymmKey: \"o4a9itWyG1YECvxpyVV9faUO\"\n" +
			"  idxc_secret: \"5oPyAqIlod4sxH1Xk7fZpNe4\"\n" +
			"  shc_secret: \"77mwFNOSUzmQLG9EGa2ZVEFq\""),
		"hec_token":    HecToken,
		"idxc_secret":  []byte("5oPyAqIlod4sxH1Xk7fZpNe4"),
		"pass4SymmKey": []byte("o4a9itWyG1YECvxpyVV9faUO"),
		"shc_secret":   []byte("77mwFNOSUzmQLG9EGa2ZVEFq"),
		"password":     AdminPassword,
	}
	secret := k8shandler.NewSecret(
		splunkSecret,
		constants.OpenshiftNS,
		data,
	)
	return secret
}
