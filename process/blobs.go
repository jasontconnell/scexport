package process

import (
	"fmt"

	"github.com/jasontconnell/sitecore/api"
)

func ReadBlobs(connstr string, groups []Group) ([]BlobData, error) {
	list := []BlobData{}
	for _, g := range groups {
		for i := 0; i < len(g.Blobs); i++ {
			b := g.Blobs[i]
			blob, err := api.LoadBlob(connstr, b.Id)
			if err != nil {
				return nil, fmt.Errorf("couldn't load blob %v %w", b.Id, err)
			}

			bdata := BlobData{Id: blob.GetId(), Data: blob.GetData(), Filename: b.Filename}
			list = append(list, bdata)
		}
	}
	return list, nil
}
