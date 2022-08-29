package egdl

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/er-azh/egmanifest"
	"github.com/er-azh/egmanifest/chunks"
)

type Downloader struct {
	wg *sync.WaitGroup
	// mut sync.RWMutex

	updates    chan *ChunkData
	manifest   *egmanifest.BinaryManifest
	goroutines int

	files  map[string]egmanifest.File
	chunks map[string]*ChunkData
}

type ChunkData struct {
	*egmanifest.Chunk
	Files []File
}

type File struct {
	Filename string
	DataSize uint32
	Offset   uint32
	Size     uint32
	CTR      uint32
}

type UpdateChan struct {
	Path string
}

func NewDownloader(manifest *egmanifest.BinaryManifest, goroutines int) *Downloader {
	downloader := &Downloader{
		wg:         &sync.WaitGroup{},
		updates:    make(chan *ChunkData, goroutines),
		manifest:   manifest,
		goroutines: goroutines,

		files:  map[string]egmanifest.File{},
		chunks: map[string]*ChunkData{},
	}

	for _, chunk := range manifest.ChunkDataList.Chunks {
		downloader.chunks[chunk.GUID.String()] = &ChunkData{Chunk: chunk}
	}

	for _, file := range manifest.FileManifestList.FileManifestList {
		downloader.files[file.FileName] = file
	}

	for i := 0; i < goroutines; i++ {
		go downloader.downloader(downloader.wg)
	}

	return downloader
}

func (dl *Downloader) Files() map[string]egmanifest.File {
	return dl.files
}

func (dl *Downloader) VerifyFile(filename, path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}

	file, ok := dl.files[filename]
	if !ok {
		return true
	}

	// Calculate checksum && Compare checksum
	hash := sha1.New()
	if _, err := io.Copy(hash, f); err != nil {
		return true
	}
	return bytes.Equal(hash.Sum(nil), file.SHAHash[:])
}

func (dl *Downloader) AddFile(filename, path string) (size uint32, ok bool) {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return 0, false
	}

	file, ok := dl.files[filename]
	if !ok {
		return 0, false
	}

	var ctr uint32
	for _, chunk := range file.ChunkParts {
		data, ok := dl.chunks[chunk.ParentGUID.String()]
		if !ok {
			return 0, false
		}

		data.Files = append(data.Files, File{
			Filename: path,
			DataSize: chunk.DataSize,
			Offset:   chunk.Offset,
			Size:     chunk.Size,
			CTR:      ctr,
		})
		ctr += chunk.Size
	}

	if err := allocateFile(path, ctr); err != nil {
		return 0, false
	}

	return ctr, true
}

func (dl *Downloader) DownloadFile(filename, path string) (uint32, error) {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return 0, err
	}

	file, ok := dl.files[filename]
	if !ok {
		return 0, fmt.Errorf("file %s not found", filename)
	}

	var ctr uint32
	for _, chunk := range file.ChunkParts {
		dl.wg.Add(1)
		dl.updates <- &ChunkData{Chunk: chunk.Chunk, Files: []File{{
			Filename: path,
			DataSize: chunk.DataSize,
			Offset:   chunk.Offset,
			Size:     chunk.Size,
			CTR:      ctr,
		}}}
	}

	dl.wg.Wait()

	return ctr, nil
}

func (dl *Downloader) Start() {
	for _, chunk := range dl.chunks {
		if len(chunk.Files) == 0 {
			continue
		}

		dl.wg.Add(1)
		dl.updates <- chunk
	}

	dl.wg.Wait()
}

func (m *Downloader) downloader(wg *sync.WaitGroup) {
	for update := range m.updates {
		err := download(update)
		if err != nil {
			fmt.Println(err)
			download(update)
		}
		wg.Done()
	}
}

func download(chunk *ChunkData) error {
	chunkGUID := strings.ToUpper(strings.ReplaceAll(chunk.GUID.String(), "-", ""))

	resp, err := http.Get(chunk.Chunk.GetURL("http://epicgames-download1.akamaized.net/Builds/Fortnite/CloudDir/ChunksV4"))
	if err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("failed to download %s: URL %s -> status %s", chunkGUID, resp.Request.URL, resp.Status)
	}

	chunkData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	for _, f := range chunk.Files {
		err := func() error {
			result, err := chunks.ParseChunk(bytes.NewReader(chunkData))
			if err != nil {
				return err
			}
			file, err := os.OpenFile(f.Filename, os.O_RDWR, os.ModePerm)
			if err != nil {
				return err
			}
			defer file.Close()

			file.Seek(int64(f.CTR), io.SeekStart)
			result.Seek(int64(f.Offset), io.SeekCurrent)
			_, err = io.CopyN(file, result, int64(f.Size))
			if err != nil {
				return err
			}

			return nil
		}()
		if err != nil {
			return err
		}
	}

	fmt.Println(chunk.GUID.String())
	return nil
}
