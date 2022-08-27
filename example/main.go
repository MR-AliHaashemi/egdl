package main

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/MR-AliHaashemi/fndl"
)

func main() {
	pv, _ := fndl.NewManifestProvider()
	inf, _ := pv.GetManifestInfo(fndl.Windows)
	manifest, _ := pv.Download(inf)

	downloader := fndl.NewDownloader(manifest, runtime.NumCPU()*2)

	for _, file := range downloader.Files() {
		path := filepath.Join("Fortnite", file.FileName)

		if downloader.VerifyFile(file.FileName, path) {
			fmt.Println("Verified: ", file.FileName)
			continue
		}

		downloader.AddFile(file.FileName, path)

		fmt.Println("Added: ", file.FileName)
	}

	downloader.Start()
}
