package kibana

import (
	consolev1 "github.com/openshift/api/console/v1"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewConsoleLink(name, href string) *consolev1.ConsoleLink {
	return &consolev1.ConsoleLink{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: consolev1.ConsoleLinkSpec{
			Location: consolev1.ApplicationMenu,
			Link: consolev1.Link{
				Text: "Logging",
				Href: href,
			},
			ApplicationMenu: &consolev1.ApplicationMenuSpec{
				Section: "Monitoring",
			},
		},
	}
}

func consoleLinksEqual(current, desired *consolev1.ConsoleLink) bool {
	if current.Spec.Location != desired.Spec.Location {
		logrus.Debugf("Location change detected for ConsoleLink CR %q", current.Name)
		return false
	}

	if current.Spec.Link.Text != desired.Spec.Link.Text {
		logrus.Debugf("Link text change detected for ConsoleLink CR %q", current.Name)
		return false
	}

	if current.Spec.Link.Href != desired.Spec.Link.Href {
		logrus.Debugf("Link href change detected for ConsoleLink CR %q", current.Name)
		return false
	}

	if current.Spec.ApplicationMenu.Section != desired.Spec.ApplicationMenu.Section {
		logrus.Debugf("ApplicationMenu section change detected for ConsoleLink CR %q", current.Name)
		return false
	}

	return true
}
