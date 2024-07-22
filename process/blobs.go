package process

import (
	"encoding/base64"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jasontconnell/sitecore/api"
)

const parallelWriteProcesses int = 8

func ProcessBlobs(connstr string, groups []Group, ws WriteSettings) {
	allblobs := []Blob{}
	dedup := make(map[string]bool)
	for _, g := range groups {
		for _, b := range g.Blobs {
			if _, ok := dedup[b.Filename]; ok {
				continue
			}
			allblobs = append(allblobs, b)
			dedup[b.Filename] = true
		}
	}

	if len(allblobs) == 0 {
		log.Println("no blobs to process")
		return
	}

	bchan := make(chan BlobData, 50000)
	echan := make(chan error, 50000)

	go errListener(echan)
	var wg sync.WaitGroup

	log.Println("reading", len(allblobs), "blobs")
	if len(allblobs) > 100 {
		size := len(allblobs) / parallelWriteProcesses
		batches := int(math.Ceil(float64(len(allblobs)) / float64(size)))

		log.Println("processing", batches, "batches of", size, "blobs each")

		wg.Add(batches)
		for i := 0; i < batches; i++ {
			go func(s, idx int) {
				start := s * idx
				end := idx*s + s
				if end >= len(allblobs) {
					end = len(allblobs) - 1
				}
				batch := allblobs[start:end]
				readBlobs(connstr, batch, ws, bchan, echan)
				wg.Done()
			}(size, i)
		}
		wg.Wait()
	} else {
		readBlobs(connstr, allblobs, ws, bchan, echan)
	}

	log.Println("writing", len(bchan), "blobs")
	wg.Add(parallelWriteProcesses)
	for i := 0; i < parallelWriteProcesses; i++ {
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

func readBlobs(connstr string, blobs []Blob, ws WriteSettings, blobchan chan BlobData, echan chan error) {
	existing := make(map[string]bool)
	files, _ := os.ReadDir(ws.BlobLocation)
	for _, f := range files {
		nm := strings.TrimSuffix(f.Name(), "."+ws.ContentFormat)
		existing[nm] = true
	}

	for _, b := range blobs {
		if _, ok := existing[b.Filename]; ok {
			continue
		}

		blob, err := api.LoadBlob(connstr, b.BlobId)
		if err != nil {
			echan <- fmt.Errorf("couldn't load blob %v %w", b.BlobId, err)
		}

		bdata := BlobData{ItemId: b.ItemId, BlobId: b.BlobId, Path: b.Path, Data: blob.GetData(), Attrs: b.Attrs, Filename: b.Filename}
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
				bxml := BlobXml{ItemId: b.ItemId.String(), BlobId: b.BlobId.String(), Path: b.Path, Filename: b.Filename, Length: len(b.Data), Fields: bfields, Data: BlobDataXml{Data: base64.StdEncoding.EncodeToString(b.Data)}}
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
