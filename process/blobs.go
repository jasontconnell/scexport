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

	go errListener(echan)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		readBlobs(connstr, groups, ws, bchan, echan)
		wg.Done()
	}()
	wg.Wait()

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			writeBlobs(ws, bchan, echan)
			wg.Done()
		}()
	}

	wg.Wait()
}

func errListener(echan chan error) {
	for {
		select {
		case err := <-echan:
			log.Println("error occurred in process blobs", err.Error())
		}
	}
}

func readBlobs(connstr string, groups []Group, ws WriteSettings, blobchan chan BlobData, echan chan error) {
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
		// if _, ok := existing[b.Filename]; ok {
		// 	continue
		// }

		blob, err := api.LoadBlob(connstr, b.Id)
		if err != nil {
			echan <- fmt.Errorf("couldn't load blob %v %w", b.Id, err)
		}

		bdata := BlobData{Id: blob.GetId(), Data: blob.GetData(), Attrs: b.Attrs, Filename: b.Filename}
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
	done := false
	for !done {
		select {
		case b := <-bchan:
			log.Println("writing blob xml", b.Filename)
			if isxml {
				bfields := []BlobFieldXml{}
				for _, f := range b.Attrs {
					bfields = append(bfields, BlobFieldXml{Name: f.Name, Value: f.Value})
				}
				bxml := BlobXml{Id: b.Id.String(), Filename: b.Filename, Length: len(b.Data), Fields: bfields, Data: BlobDataXml{Data: base64.StdEncoding.EncodeToString(b.Data)}}
				path := filepath.Join(fulldir, bxml.Filename+".xml")
				err = writeBlobXml(path, bxml)
				if err != nil {
					echan <- fmt.Errorf("writing file contents for %s, path: %s. %w", b.Filename, path, err)
				}
			}
		default:
			done = true
		}
	}
}
