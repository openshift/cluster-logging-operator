package syslog

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("#MapSyslog", func() {
	var (
		url = "0.0.0.0:9200"
	)
	It("should map logging.Syslog to obs.Syslog", func() {
		loggingOutSpec := logging.OutputSpec{
			URL: url,
			OutputTypeSpec: logging.OutputTypeSpec{
				Syslog: &logging.Syslog{
					RFC:        "RFC3164",
					Severity:   "error",
					Facility:   "foo",
					PayloadKey: "bar",
					AppName:    "app",
					ProcID:     "123",
					MsgID:      "12345",
				},
			},
		}

		expObsSyslog := &obs.Syslog{
			URL:        url,
			RFC:        "RFC3164",
			Severity:   "error",
			Facility:   "foo",
			PayloadKey: "{.bar}",
			AppName:    "app",
			ProcID:     "123",
			MsgID:      "12345",
		}

		Expect(MapSyslog(loggingOutSpec)).To(Equal(expObsSyslog))
	})

	It("should map logging.Syslog to obs.Syslog when RFC is not specified", func() {
		loggingOutSpec := logging.OutputSpec{
			URL: url,
			OutputTypeSpec: logging.OutputTypeSpec{
				Syslog: &logging.Syslog{
					Severity:   "error",
					Facility:   "foo",
					PayloadKey: "bar",
					AppName:    "app",
					ProcID:     "123",
					MsgID:      "12345",
				},
			},
		}

		expObsSyslog := &obs.Syslog{
			URL:        url,
			RFC:        obs.SyslogRFC5424,
			Severity:   "error",
			Facility:   "foo",
			PayloadKey: "{.bar}",
			AppName:    "app",
			ProcID:     "123",
			MsgID:      "12345",
		}

		Expect(MapSyslog(loggingOutSpec)).To(Equal(expObsSyslog))
	})

	It("should map logging.Syslog to obs.Syslog and default values when facility & severity are not spec'd", func() {
		loggingOutSpec := logging.OutputSpec{
			URL: url,
			OutputTypeSpec: logging.OutputTypeSpec{
				Syslog: &logging.Syslog{
					PayloadKey: "bar",
					AppName:    "app",
					ProcID:     "123",
					MsgID:      "12345",
				},
			},
		}

		expObsSyslog := &obs.Syslog{
			URL:        url,
			RFC:        obs.SyslogRFC5424,
			Severity:   "informational",
			Facility:   "user",
			PayloadKey: "{.bar}",
			AppName:    "app",
			ProcID:     "123",
			MsgID:      "12345",
		}

		Expect(MapSyslog(loggingOutSpec)).To(Equal(expObsSyslog))
	})
})
