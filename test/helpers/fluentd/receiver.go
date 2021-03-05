// package fluentd provides a fluentd receiver for use in e2e forwarding tests.
package fluentd

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/helpers/certificate"
	"github.com/openshift/cluster-logging-operator/test/helpers/cmd"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
)

const (
	appName   = "fluentd-receiver"
	dataDir   = "/opt/app-root/data"
	configDir = "/opt/app-root/etc"
)

// Receiver is a service running fluentd, listening on on one or more source ports.
type Receiver struct {
	Name      string
	Sources   map[string]*Source
	Pod       *corev1.Pod
	ConfigMap *corev1.ConfigMap

	service *corev1.Service
	config  strings.Builder
	ready   chan struct{}
}

// Source represents a fluentd listening port and output file.
type Source struct {
	// Name for the source in the Receiver's Source map, must be unique per receiver.
	Name string
	// Type is the Fluentd source @type parameter, normally "forward".
	Type string
	// Port is the listening port for this source.
	Port int
	// Cert enables sources as TLS servers.
	Cert *certificate.CertKey
	// CA enables client verification using this CA.
	CA *certificate.CertKey
	// SharedKey enables fluentd's shared_key authentication.
	SharedKey string

	receiver *Receiver
}

// Host is the host name of the receiver service.
func (s *Source) Host() string { return s.receiver.Host() }

// OutFile is the path of the output file for this source on the fluentd Pod.
func (s *Source) OutFile() string { return filepath.Join(dataDir, s.Name) }

// TailReader returns a CmdReader that tails the source's log stream.
// Will fail if called before Receiver.Create() has returned.
//
// It waits for the file to exist, and will tail the output file forever,
// so it won't normally return io.EOF.
func (s *Source) TailReader() *cmd.Reader {
	r, err := cmd.TailReader(s.receiver.Pod, s.OutFile())
	test.Must(err)
	return r
}

// HasOutput returns true if the source's output file exists and is non empty.
func (s *Source) HasOutput() (bool, error) {
	script := fmt.Sprintf("if test -s %q; then echo yes; fi", s.OutFile())
	out, err := runtime.Exec(s.receiver.Pod, "sh", "-c", script).Output()
	if err != nil {
		return false, utils.WrapError(err)
	}
	return len(out) > 0, nil
}

// NewReceiver returns a receiver with no sources. Use AddSource() before calling Create().
func NewReceiver(ns, name string) *Receiver {
	r := &Receiver{
		Name:      name,
		Sources:   map[string]*Source{},
		ConfigMap: runtime.NewConfigMap(ns, name, nil),
		Pod:       runtime.NewPod(ns, name),
		service:   runtime.NewService(ns, name),
		ready:     make(chan struct{}),
	}
	r.Pod.Spec.RestartPolicy = corev1.RestartPolicyNever // Don't restart if fluentd fails.
	runtime.Labels(r.Pod)[appName] = name
	r.Pod.Spec.Containers = []corev1.Container{{
		Name:  name,
		Image: "quay.io/openshift/origin-logging-fluentd:latest",
		Args:  []string{"fluentd", "-c", filepath.Join(configDir, "fluent.conf")},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "config",
				ReadOnly:  true,
				MountPath: configDir,
			},
			{
				Name:      "data",
				MountPath: dataDir,
			},
		},
	}}
	r.Pod.Spec.Volumes = []corev1.Volume{
		{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: name,
					},
				},
			},
		},
		{
			Name: "data",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
	r.service.Spec = corev1.ServiceSpec{
		Selector: map[string]string{appName: name},
	}
	// Start of fluent config: ignore fluent's internal events
	r.config.WriteString(`
<label @FLUENT_LOG>
  <match **>
    @type null
  </match>
</label>
`)
	return r
}

func (r *Receiver) AddSource(s *Source) {
	s.receiver = r
	r.Sources[s.Name] = s
	r.service.Spec.Ports = append(r.service.Spec.Ports,
		corev1.ServicePort{Name: s.Name, Port: int32(s.Port)},
	)
}

func (s *Source) config() {
	if s.Cert != nil {
		s.receiver.ConfigMap.Data[s.Name+"-cert.pem"] = string(s.Cert.CertificatePEM())
		s.receiver.ConfigMap.Data[s.Name+"-key.pem"] = string(s.Cert.PrivateKeyPEM())
	}
	if s.CA != nil {
		s.receiver.ConfigMap.Data[s.Name+"-ca.pem"] = string(s.CA.CertificatePEM())
		if s.Cert == nil {
			panic("Cannot set Source.CA without setting Source.Cert")
		}
	}
	// Note output @type exec rather than file, to send everything to a single file.
	t := template.Must(template.New("config").Funcs(template.FuncMap{
		"ToUpper":    strings.ToUpper,
		"ConfigPath": func(name string) string { return s.receiver.ConfigPath(name) },
	}).Parse(`
<source>
  @type {{.Type}}
  port {{.Port}}
  @label @{{.Name | ToUpper}}
{{- if .Cert}}
  <transport tls>
      cert_path {{ConfigPath .Name}}-cert.pem
      private_key_path {{ConfigPath .Name}}-key.pem
{{- if .CA}}
      ca_path {{ConfigPath .Name}}-ca.pem
      client_cert_auth true
{{- end}}
  </transport>
{{- end}}
{{- if .SharedKey}}
  <security>
    shared_key "{{.SharedKey}}"
    self_hostname "#{ENV['NODE_NAME']}"
  </security>
{{end}}
</source>
<label @{{.Name | ToUpper}}>
  <match **>
    @type exec
    command bash -c 'cat < $0 >> {{.OutFile}}'
    <format>
      @type json
    </format>
    <buffer>
      @type memory
      flush_mode immediate
    </buffer>
  </match>
</label>
`))
	test.Must(t.Execute(&s.receiver.config, s))
}

// Create the receiver's resources. Blocks till created.
func (r *Receiver) Create(c *client.Client) error {
	for _, s := range r.Sources {
		s.config()
	}
	r.ConfigMap.Data["fluent.conf"] = r.config.String()
	// Crate and wait for the pod concurrently.
	g := errgroup.Group{}
	g.Go(func() error { return c.Create(r.ConfigMap) })
	g.Go(func() error { return c.Create(r.service) })
	g.Go(func() error {
		if err := c.Create(r.Pod); err != nil {
			return err
		}
		return c.WaitFor(r.Pod, client.PodRunning)
	})
	return g.Wait()
}

// Host name for the receiver.
func (r *Receiver) Host() string { return runtime.ServiceDomainName(r.service) }

// ConfigPath returns a path relative to the configuration dir.
func (r *Receiver) ConfigPath(path ...string) string {
	return filepath.Join(append([]string{configDir}, path...)...)
}
