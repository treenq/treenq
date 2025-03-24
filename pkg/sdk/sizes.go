package tqsdk

import (
	"errors"
	"strconv"
	"strings"
)

type SizeSlug string

type ComputationResource struct {
	CpuUnits   int
	MemoryMibs int
	DiskGibs   int
}

var ErrInvalidSizeSlug = errors.New("size slug expected as mCput-mMem-gDisk")

func (s SizeSlug) ToComputationResource() (ComputationResource, error) {
	parts := strings.Split(string(s), "-")
	if len(parts) != 3 {
		return ComputationResource{}, ErrInvalidSizeSlug
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
	}, nil
}

const (
	SizeSlugS SizeSlug = "500-1024-2"
)
