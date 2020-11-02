// Licensed to Red Hat, Inc under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Red Hat, Inc licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package fluentd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

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
	done    chan struct{}
	err     error
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

	r *Receiver
}

// Host is the host name of the receiver service.
func (s *Source) Host() string { return s.r.Host() }

// OutFile is the path of the output file for this source on the fluentd Pod.
func (s *Source) OutFile() string { return filepath.Join(dataDir, s.Name) }

// TailReader returns a CmdReader that tails the source's log stream.
//
// It waits for the file to exist, and will tail the output file forever,
// so it won't normally return io.EOF.
func (s *Source) TailReader() *cmd.Reader {
	test.Must(s.r.Wait()) // Make sure r.Pod is running before we exec.
	r, err := cmd.NewReader(runtime.Exec(s.r.Pod, "tail", "-F", s.OutFile()))
	test.Must(err)
	return r
}

// NewReceiver creates a receiver. Use AddSource() to add sources before calling Create().
func NewReceiver(ns, name string) *Receiver {
	r := &Receiver{
		Name:      name,
		Sources:   map[string]*Source{},
		ConfigMap: runtime.NewConfigMap(ns, name, nil),
		Pod:       runtime.NewPod(ns, name),
		service:   runtime.NewService(ns, name),
		done:      make(chan struct{}),
	}
	r.Pod.Spec.RestartPolicy = corev1.RestartPolicyNever // Don't restart if fluentd fails.
	runtime.Labels(r.Pod)[appName] = name
	r.Pod.Spec.Containers = []corev1.Container{
		{
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
		},
	}
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
<system>
  log_level trace
</system>

<label @FLUENT_LOG>
  <match **>
    @type null
  </match>
</label>
`)
	return r
}

func (r *Receiver) AddSource(name, sourceType string, port int, extra ...string) *Source {
	s := &Source{Name: name, Type: sourceType, Port: port, r: r}
	r.Sources[s.Name] = s
	r.service.Spec.Ports = append(r.service.Spec.Ports,
		corev1.ServicePort{Name: s.Name, Port: int32(s.Port)},
	)
	return s
}

func (s *Source) config() {
	if s.Cert != nil {
		s.r.ConfigMap.Data[s.Name+"-cert.pem"] = string(s.Cert.CertificatePEM())
		s.r.ConfigMap.Data[s.Name+"-key.pem"] = string(s.Cert.PrivateKeyPEM())
	}
	if s.CA != nil {
		s.r.ConfigMap.Data[s.Name+"-ca.pem"] = string(s.CA.CertificatePEM())
		if s.Cert == nil {
			panic("Cannot set Source.CA without setting Source.Cert")
		}
	}
	// Note output @type exec rather than file, to send everything to a single file.
	t := template.Must(template.New("config").Funcs(template.FuncMap{
		"ToUpper":    strings.ToUpper,
		"ConfigPath": func(name string) string { return s.r.ConfigPath(name) },
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
	test.Must(t.Execute(&s.r.config, s))
}

// Create the receiver's resources, wait for pod to be running.
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
		err := c.WaitFor(r.Pod, client.PodRunning)
		return err
	})
	if err := g.Wait(); err != nil {
		r.err = fmt.Errorf("%w\n%v", err, r.Logs())
	}
	close(r.done)
	return r.err
}

func (r *Receiver) Logs() string {
	ns, name := r.Pod.Namespace, r.Pod.Name
	cmd := exec.Command("oc", "logs", "-n", ns, name)
	b, _ := cmd.CombinedOutput()
	return string(b)
}

// Wait till r.Pod is running.
func (r *Receiver) Wait() error { <-r.done; return r.err }

// Host name for the receiver.
func (r *Receiver) Host() string { return runtime.ServiceDomainName(r.service) }

// ConfigPath returns a path relative to the configuration dir.
func (r *Receiver) ConfigPath(path ...string) string {
	return filepath.Join(append([]string{configDir}, path...)...)
}
