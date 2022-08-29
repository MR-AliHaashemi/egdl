package egdl

import (
	"os"

	"github.com/er-azh/egmanifest"
)

func HasAnyTags(file egmanifest.File, tags ...string) bool {
	for _, fileTag := range file.InstallTags {
		for _, validTag := range tags {
			if fileTag == validTag {
				return true
			}
		}
	}

	return false
}

func GetFileSize(file egmanifest.File) (size uint32) {
	for _, chunk := range file.ChunkParts {
		size += chunk.Size
	}
	return
}

func allocateFile(path string, size uint32) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		_, err = os.Create(path)
	}
	if err != nil {
		return err
	}

	return os.Truncate(path, int64(size))
}
