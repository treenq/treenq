package tqsdk

import (
	"strconv"
	"strings"
)

type SizeSlug string

type ComputationResource struct {
	CpuUnits   int
	MemoryMibs int
	DiskGibs   int
}

func (s SizeSlug) ToComputationResource() ComputationResource {
	parts := strings.Split(string(s), "-")
	if len(parts) != 3 {
		panic("expected 3 parts in SizeSlug value, given=" + s)
	}

	cpuStr, memStr, diskStr := parts[0], parts[1], parts[2]
	cpu, err := strconv.Atoi(cpuStr)
	if err != nil {
		panic(err)
	}

	mem, err := strconv.Atoi(memStr)
	if err != nil {
		panic(err)
	}

	disk, err := strconv.Atoi(diskStr)
	if err != nil {
		panic(err)
	}

	return ComputationResource{
		CpuUnits:   cpu,
		MemoryMibs: mem,
		DiskGibs:   disk,
	}
}

const (
	SizeSlugS SizeSlug = "500-1024-2"
)
