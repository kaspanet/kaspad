package bip32

import (
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type path struct {
	isPublic bool
	indexes  []uint32
}

func parsePath(pathString string) (*path, error) {
	parts := strings.Split(pathString, "/")
	isPublic := false
	switch parts[0] {
	case "m":
		isPublic = false
	case "M":
		isPublic = true
	default:
		return nil, errors.Errorf("%s is an invalid extended key type", parts[0])
	}

	indexParts := parts[1:]
	indexes := make([]uint32, len(indexParts))
	for i, part := range indexParts {
		var err error
		indexes[i], err = parseIndex(part)
		if err != nil {
			return nil, err
		}
	}

	return &path{
		isPublic: isPublic,
		indexes:  indexes,
	}, nil
}

func parseIndex(indexString string) (uint32, error) {
	const isHardenedSuffix = "'"
	isHardened := strings.HasSuffix(indexString, isHardenedSuffix)
	trimmedIndexString := strings.TrimSuffix(indexString, isHardenedSuffix)
	index, err := strconv.Atoi(trimmedIndexString)
	if err != nil {
		return 0, err
	}

	if index >= hardenedIndexStart {
		return 0, errors.Errorf("max index value is %d but got %d", hardenedIndexStart, index)
	}

	if isHardened {
		return uint32(index) + hardenedIndexStart, nil
	}

	return uint32(index), nil
}

func PathToPublic(path string) string {
	trimmed := strings.TrimPrefix(path, "m")
	return "M" + trimmed
}
