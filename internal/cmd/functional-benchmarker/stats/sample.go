package stats

import (
	"math"
	"regexp"
	"strconv"

	"github.com/ViaQ/logerr/v2/log"
)

var (
	re = regexp.MustCompile("(?P<value>[0-9]*)(?P<unit>[mEPTGMKi])?")
)

type Sample struct {
	Time        int64
	CPUCores    string
	MemoryBytes string
}

func (s *Sample) CPUCoresAsFloat() float64 {
	if matches := re.FindStringSubmatch(s.CPUCores); len(matches) > 0 {
		value := "0"
		if index := re.SubexpIndex("value"); index > 0 {
			value = matches[index]
		}
		intVar, err := strconv.Atoi(value)
		if err != nil {
			log.NewLogger("stats-testing").Error(err, "Error converting value", "value", value)
			return 0.0
		}
		if index := re.SubexpIndex("unit"); index > 0 {
			if matches[index] == "m" {
				return float64(intVar) / 1000.0
			}
		}
		return float64(intVar)
	}
	return 0.0
}
func (s *Sample) MemoryBytesAsFloat() float64 {
	if matches := re.FindStringSubmatch(s.MemoryBytes); len(matches) > 0 {
		value := "0"
		if index := re.SubexpIndex("value"); index > 0 {
			value = matches[index]
		}
		intVar, err := strconv.Atoi(value)
		if err != nil {
			log.NewLogger("stats-testing").Error(err, "Error converting value", "value", value)
			return 0.0
		}
		if index := re.SubexpIndex("unit"); index > 0 {
			switch unit := matches[index]; unit {
			case "Ki":
				return float64(intVar) * math.Pow(2.0, 10)
			case "Gi":
				return float64(intVar) * math.Pow(2.0, 30)
			case "Mi":
				return float64(intVar) * math.Pow(2.0, 20)
			case "K":
				return float64(intVar) * 1000.0
			case "G":
				return float64(intVar) * 1000000000.0
			case "M":
				return float64(intVar) * 1000000.0
			}
		}
		return float64(intVar)
	}
	return 0.0
}

type ResourceMetrics struct {
	Samples []Sample
}

func NewResourceMetrics() *ResourceMetrics {
	return &ResourceMetrics{
		Samples: []Sample{},
	}
}

func (rm *ResourceMetrics) AddSample(sample *Sample) {
	log.NewLogger("stats-testing").V(3).Info("Adding resource metric", "sample", sample)
	if sample == nil {
		return
	}
	rm.Samples = append(rm.Samples, *sample)
}
