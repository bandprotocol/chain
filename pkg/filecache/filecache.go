package filecache

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/peterbourgon/diskv"
)

type Cache struct {
	fileCache *diskv.Diskv
}

// New creates and returns a new file-backed data caching instance.
// If cacheSizeBytes is 0, it defaults to 32MB.
func New(basePath string, cacheSizeBytes int64) Cache {
	if cacheSizeBytes <= 0 {
		cacheSizeBytes = 32 * 1024 * 1024 // Default to 32MB
	}
	return Cache{
		fileCache: diskv.New(diskv.Options{
			BasePath:     basePath,
			Transform:    func(s string) []string { return []string{} },
			CacheSizeMax: cacheSizeBytes,
		}),
	}
}

func GetFilename(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// AddFile saves the given data to a file in HOME/files directory using sha256 sum as filename.
func (c Cache) AddFile(data []byte) string {
	filename := GetFilename(data)
	if !c.fileCache.Has(filename) {
		_ = c.fileCache.Write(filename, data)
	}
	return filename
}

// GetFile loads the file from the file storage. Returns error if the file does not exist.
func (c Cache) GetFile(filename string) ([]byte, error) {
	data, err := c.fileCache.Read(filename)
	if err != nil {
		return nil, err
	}
	if GetFilename(data) != filename { // Perform integrity check for safety. NEVER EXPECT TO HIT.
		return nil, errors.New("inconsistent filecache content")
	}
	return data, nil
}

// MustGetFile loads the file from the file storage. Panics if the file does not exist.
func (c Cache) MustGetFile(filename string) []byte {
	data, err := c.GetFile(filename)
	if err != nil {
		panic(err)
	}
	return data
}
