package process

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jasontconnell/sitecore/api"
)

func ProcessBlobs(connstr string, groups []Group, ws WriteSettings) {
	bchan := make(chan BlobData, 50000)
	echan := make(chan error, 50000)
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			readBlobs(connstr, groups, ws, bchan, echan)
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			writeBlobs(ws, bchan, echan)
			wg.Done()
		}()
	}

	go func() {
		for {
			select {
			case err := <-echan:
				log.Println("error occurred in process blobs", err.Error())
			}
		}
	}()

	wg.Wait()
}

func readBlobs(connstr string, groups []Group, ws WriteSettings, blobchan chan BlobData, errorchan chan error) {

	dedup := make(map[string]Blob)
	for _, g := range groups {
		for _, b := range g.Blobs {
			dedup[b.Filename] = b
		}
	}

	existing := make(map[string]bool)
	files, _ := os.ReadDir(ws.BlobLocation)
	for _, f := range files {
		nm := strings.TrimSuffix(f.Name(), "."+ws.ContentFormat)
		existing[nm] = true
	}

	for _, b := range dedup {
		if _, ok := existing[b.Filename]; ok {
			continue
		}

		blob, err := api.LoadBlob(connstr, b.Id)
		if err != nil {
			errorchan <- fmt.Errorf("couldn't load blob %v %w", b.Id, err)
		}

		bdata := BlobData{Id: blob.GetId(), Data: blob.GetData(), Filename: b.Filename}
		blobchan <- bdata
	}
}

func writeBlobs(settings WriteSettings, bchan chan BlobData, echan chan error) {
	fulldir := settings.BlobLocation
	err := os.MkdirAll(fulldir, os.ModePerm)
	if err != nil {
		echan <- fmt.Errorf("couldn't create dir structure %s. %w", fulldir, err)
	}
	isxml := settings.ContentFormat == "xml"
	for {
		select {
		case b := <-bchan:
			log.Println("writing blob xml", b.Filename)
			if isxml {
				bxml := BlobXml{Id: b.Id.String(), Filename: b.Filename, Length: len(b.Data), Data: base64.StdEncoding.EncodeToString(b.Data)}
				path := filepath.Join(fulldir, bxml.Filename+".xml")
				err = writeBlobXml(path, bxml)
				if err != nil {
					echan <- fmt.Errorf("writing file contents for %s, path: %s. %w", b.Filename, path, err)
				}
			}
		}
	}
}
