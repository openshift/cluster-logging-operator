package comparators

import (
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// CompareVersions will return one of:
// -1 : if lhs > rhs
// 0  : if lhs == rhs
// 1  : if rhs > lhs
func CompareVersions(lhs, rhs string) int {
	lVersions := buildVersionArray(lhs)
	rVersions := buildVersionArray(rhs)

	lLen := len(lVersions)
	rLen := len(rVersions)

	for i := 0; i < lLen && i < rLen; i++ {
		if lVersions[i] > rVersions[i] {
			return -1
		}

		if lVersions[i] < rVersions[i] {
			return 1
		}
	}

	// check if lhs is a more specific version number (aka newer)
	if lLen > rLen {
		return -1
	}

	// check if rhs is a more specific version number
	if lLen < rLen {
		return 1
	}

	// versions are exactly the same
	return 0
}

func buildVersionArray(version string) []int {

	versions := []int{}
	for _, v := range strings.Split(version, ".") {
		i, err := strconv.Atoi(v)
		if err != nil {
			logrus.Warnf("unable to build version array: %v", err)
			break
		}

		versions = append(versions, i)
	}

	return versions
}
