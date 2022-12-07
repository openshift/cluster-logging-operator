package scc

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	security "github.com/openshift/api/security/v1"
	"reflect"
	"sort"
)

func AreSame(current, desired security.SecurityContextConstraints) (bool, string) {
	log.V(3).Info("Comparing SCC current to desired", "current", current, "desired", desired)

	if same, property := samePriority(current.Priority, desired.Priority); !same {
		return same, property
	}

	if current.AllowPrivilegedContainer != desired.AllowPrivilegedContainer {
		log.V(3).Info("SCC AllowPrivilegedContainer change", "current name", current.Name)
		return false, "allowPrivilegedContainer"
	}

	sort.Slice(current.RequiredDropCapabilities, func(i, j int) bool { return current.RequiredDropCapabilities[i] < current.RequiredDropCapabilities[j] })
	sort.Slice(desired.RequiredDropCapabilities, func(i, j int) bool { return desired.RequiredDropCapabilities[i] < desired.RequiredDropCapabilities[j] })
	if !reflect.DeepEqual(current.RequiredDropCapabilities, desired.RequiredDropCapabilities) {
		log.V(3).Info("SCC RequiredDropCapabilities change", "current name", current.Name)
		return false, "requiredDropCapabilities"
	}

	if current.AllowHostDirVolumePlugin != desired.AllowHostDirVolumePlugin {
		log.V(3).Info("SCC AllowHostDirVolumePlugin change", "current name", current.Name)
		return false, "allowHostDirVolumePlugin"
	}

	sort.Slice(current.Volumes, func(i, j int) bool { return current.Volumes[i] < current.Volumes[j] })
	sort.Slice(desired.Volumes, func(i, j int) bool { return desired.Volumes[i] < desired.Volumes[j] })
	if !reflect.DeepEqual(current.Volumes, desired.Volumes) {
		log.V(3).Info("SCC Volumes change", "current name", current.Name)
		return false, "volumes"
	}

	if !sameAllowPrivilegeEscalation(current.DefaultAllowPrivilegeEscalation, desired.DefaultAllowPrivilegeEscalation) {
		log.V(3).Info("SCC DefaultAllowPrivilegeEscalation change", "current", current.DefaultAllowPrivilegeEscalation, "desired", desired.DefaultAllowPrivilegeEscalation)
		return false, "defaultAllowPrivilegeEscalation"
	}

	if !sameAllowPrivilegeEscalation(current.AllowPrivilegeEscalation, desired.AllowPrivilegeEscalation) {
		log.V(3).Info("SCC AllowPrivilegeEscalation change", "current", current.AllowPrivilegeEscalation, "desired", desired.AllowPrivilegeEscalation)
		return false, "allowPrivilegeEscalation"
	}

	if !reflect.DeepEqual(current.RunAsUser, desired.RunAsUser) {
		log.V(3).Info("SCC RunAsUser change", "current name", current.Name)
		return false, "runAsUser"
	}
	if !reflect.DeepEqual(current.SELinuxContext, desired.SELinuxContext) {
		log.V(3).Info("SCC SELinuxContext change", "current name", current.Name)
		return false, "SELinuxContext"
	}

	if current.ReadOnlyRootFilesystem != desired.ReadOnlyRootFilesystem {
		log.V(3).Info("SCC ReadOnlyRootFilesystem change", "current name", current.Name)
		return false, "allowPrivilegeEscalation"
	}

	sort.Slice(current.ForbiddenSysctls, func(i, j int) bool { return current.ForbiddenSysctls[i] < current.ForbiddenSysctls[j] })
	sort.Slice(desired.ForbiddenSysctls, func(i, j int) bool { return desired.ForbiddenSysctls[i] < desired.ForbiddenSysctls[j] })
	if !reflect.DeepEqual(current.ForbiddenSysctls, desired.ForbiddenSysctls) {
		log.V(3).Info("SCC ForbiddenSysctls change", "current name", current.Name)
		return false, "ForbiddenSysctls"
	}

	sort.Slice(current.SeccompProfiles, func(i, j int) bool { return current.SeccompProfiles[i] < current.SeccompProfiles[j] })
	sort.Slice(desired.SeccompProfiles, func(i, j int) bool { return desired.SeccompProfiles[i] < desired.SeccompProfiles[j] })
	if !reflect.DeepEqual(current.SeccompProfiles, desired.SeccompProfiles) {
		log.V(3).Info("SCC SeccompProfiles change", "current name", current.Name)
		return false, "SeccompProfiles"
	}

	return true, ""
}

func samePriority(current, desired *int32) (bool, string) {
	if current == nil && desired != nil || current != nil && desired == nil {
		log.V(3).Info("SCC AllowPrivilegedContainer change")
		return false, "priority"
	}
	if current != nil && desired != nil && *current != *desired {
		log.V(3).Info("SCC AllowPrivilegedContainer change")
		return false, "priority"
	}
	return true, ""
}

func sameAllowPrivilegeEscalation(current, desired *bool) bool {
	if (current != nil && desired == nil) ||
		(current == nil && desired != nil) ||
		(current != nil && desired != nil &&
			*current != *desired) {
		return false
	}

	if (desired != nil && current == nil) ||
		(desired == nil && current != nil) ||
		(desired != nil && current != nil &&
			*desired != *current) {
		return false
	}
	return true
}
