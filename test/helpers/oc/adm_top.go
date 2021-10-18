package oc

import (
	"fmt"
	"strings"
)

// Execer is interface for collecting arguments for Exec command
type AdmTopComman interface {
	Command

	//ForContainers to retrieve `oc adm top` of containers
	ForContainers() AdmTopComman

	//NoHeaders to not include noHeaders
	NoHeaders() AdmTopComman
}

type admTop struct {
	*runner
	namespace  string
	podname    string
	containers bool
	noHeaders  bool
}

// AdmTop creates an 'oc adm top' command
func AdmTop(namespace, name string) AdmTopComman {
	e := &admTop{
		runner:    &runner{},
		namespace: namespace,
		podname:   name,
	}
	e.collectArgsFunc = e.args
	return e
}

func (e *admTop) ForContainers() AdmTopComman {
	e.containers = true
	return e
}

func (e *admTop) NoHeaders() AdmTopComman {
	e.noHeaders = true
	return e
}

func (e *admTop) String() string {
	args := []string{}
	if e.containers {
		args = append(args, "--containers")
	}
	if e.noHeaders {
		args = append(args, "--no-headers")
	}
	return fmt.Sprintf("adm -n %s top pod %s %s", e.namespace, e.podname, strings.Join(args, " "))
}

// creates command args to be used by runner
func (e *admTop) args() []string {
	ocargs := sanitizeArgs(e.String())
	return ocargs
}
