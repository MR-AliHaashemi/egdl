package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/MR-AliHaashemi/egdl"
	"github.com/er-azh/egmanifest"
)

var url = flag.String("url", "latest", "url of the manifest that you want to download the game")
var dir = flag.String("dir", "Fortnite", "directory where you want to game stored")

func init() { flag.Parse() }

func main() {
	var err error
	var manifest *egmanifest.BinaryManifest
	if *url == "latest" {
		pv, _ := egdl.NewManifestProvider()
		inf, _ := pv.GetManifestInfo(egdl.Windows)
		manifest, err = pv.Download(inf)
		if err != nil {
			panic(err)
		}
	} else {
		manifest, err = egdl.DownloadManifest(*url)
		if err != nil {
			panic(err)
		}
	}

	downloader := egdl.NewDownloader(manifest, runtime.NumCPU()*2)

	for _, file := range downloader.Files() {
		path := filepath.Join(*dir, file.FileName)

		if downloader.VerifyFile(file.FileName, path) {
			fmt.Println("Verified: ", file.FileName)
			continue
		}

		downloader.AddFile(file.FileName, path)

		fmt.Println("Added: ", file.FileName)
	}

	downloader.Start()
}
