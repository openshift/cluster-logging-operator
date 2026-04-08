package oc

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	osexec "os/exec"
	"strings"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/onsi/ginkgo/v2"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// PortForwarder manages a port-forward connection to a Kubernetes pod or service.
type PortForwarder struct {
	localPort       uint16
	stopCh, readyCh chan struct{}
	cmd             *osexec.Cmd // used for service port-forward via oc CLI
}

// LocalPort returns the local port as a string.
func (pf *PortForwarder) LocalPort() string {
	return fmt.Sprintf("%d", pf.localPort)
}

// Stop terminates the port-forward connection.
func (pf *PortForwarder) Stop() {
	if pf.cmd != nil && pf.cmd.Process != nil {
		_ = pf.cmd.Process.Kill()
		_ = pf.cmd.Wait()
	} else {
		pf.stopCh <- struct{}{}
	}
}

// SetupPodPortForwarder sets up port-forwarding to a Kubernetes pod using the API.
// It forwards from a random local port to the specified pod port.
func SetupPodPortForwarder(cfg *rest.Config, host, namespace, podName string, podPort int32) (*PortForwarder, error) {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, podName)
	hostIP := strings.TrimPrefix(host, `https://`)

	transport, upgrader, err := spdy.RoundTripperFor(cfg)
	if err != nil {
		return nil, err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})

	pf := &PortForwarder{
		stopCh:  make(chan struct{}, 1),
		readyCh: make(chan struct{}),
	}

	fw, err := portforward.New(dialer, []string{fmt.Sprintf("0:%d", podPort)}, pf.stopCh, pf.readyCh, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	if err != nil {
		return nil, err
	}

	go func() {
		err = fw.ForwardPorts()
		if err != nil {
			panic(err)
		}
	}()
	<-pf.readyCh

	forwardedPorts, err := fw.GetPorts()
	if err != nil {
		return nil, err
	}
	if n := len(forwardedPorts); n != 1 {
		return nil, fmt.Errorf("SetupPodPortForwarder: expected one forwarded port, got %d", n)
	}
	pf.localPort = forwardedPorts[0].Local
	return pf, nil
}

// SetupServicePortForwarder sets up port-forwarding to a Kubernetes service using oc CLI.
// It forwards from a random local port to the specified service port.
func SetupServicePortForwarder(namespace, serviceName string, servicePort int32) (*PortForwarder, error) {
	// Find a free local port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("unable to find free port: %v", err)
	}
	addr := listener.Addr().(*net.TCPAddr)
	localPort := fmt.Sprintf("%d", addr.Port)
	_ = listener.Close()

	log.V(2).Info("Starting port-forward to service", "service", serviceName, "localPort", localPort)

	// Start oc port-forward command
	cmd := osexec.Command("oc", "port-forward", "-n", namespace, fmt.Sprintf("svc/%s", serviceName), fmt.Sprintf("%s:%d", localPort, servicePort))

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("unable to start port-forward: %v", err)
	}

	// Wait for port to be ready
	time.Sleep(1 * time.Second)

	// Verify port is accessible
	var conn net.Conn
	for i := 0; i < 10; i++ {
		conn, err = net.DialTimeout("tcp", "127.0.0.1:"+localPort, 1*time.Second)
		if err == nil {
			_ = conn.Close()
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("port-forward did not become ready: %v", err)
	}

	pf := &PortForwarder{
		localPort: uint16(addr.Port),
		stopCh:    make(chan struct{}, 1),
		readyCh:   make(chan struct{}),
		cmd:       cmd,
	}

	return pf, nil
}
