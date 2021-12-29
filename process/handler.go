package process

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jasontconnell/sitecore/api"
	"github.com/jasontconnell/sitecore/data"
)

type blobAttr struct {
	name, value string
}

type blobResult struct {
	blobId uuid.UUID
	name   string
	attrs  []blobAttr
	ext    string
}

type handlerResult struct {
	value string
	blobs []blobResult
	html  bool
}

var mediaReg *regexp.Regexp = regexp.MustCompile(`<image mediaid="\{([A-F0-9\-]+)\}" />`)
var mediaRteReg *regexp.Regexp = regexp.MustCompile(`src="-\/media\/([a-f0-9]{32})\.ashx`)

type FieldHandler func(fv data.FieldValueNode, item data.ItemNode, items data.ItemMap, fsetting FieldSettings, lang data.Language) (handlerResult, error)

var fieldHandlers map[string]FieldHandler

func init() {
	fieldHandlers = map[string]FieldHandler{
		"Single-Line Text": handleString,
		"Droplink":         handleReference,
		"Treelist":         handleReference,
		"Datetime":         handleString,
		"Rich Text":        handleRichText,
		"Multi-Line Text":  handleRichText,
		"Image":            handleImage,
	}
}

func ResolveField(fv data.FieldValueNode, tfld data.TemplateFieldNode, item data.ItemNode, items data.ItemMap, fsetting FieldSettings, lang data.Language) (handlerResult, error) {
	fh, ok := fieldHandlers[tfld.GetType()]
	if !ok {
		return handlerResult{}, fmt.Errorf("no handler for %s", tfld.GetType())
	}
	return fh(fv, item, items, fsetting, lang)
}

func handleString(fv data.FieldValueNode, item data.ItemNode, items data.ItemMap, fsetting FieldSettings, lang data.Language) (handlerResult, error) {
	return handlerResult{value: fv.GetValue()}, nil
}

func handleRichText(fv data.FieldValueNode, item data.ItemNode, items data.ItemMap, fsetting FieldSettings, lang data.Language) (handlerResult, error) {
	str := fv.GetValue()
	allgroups := mediaRteReg.FindAllStringSubmatch(str, -1)

	hr := handlerResult{html: true}
	for _, m := range allgroups {
		if len(m) != 2 {
			continue
		}

		idval := m[1]
		id, err := uuid.Parse(idval)
		if err != nil {
			return handlerResult{}, fmt.Errorf("couldn't parse image id in rte field %s %s. %w", fv.GetName(), fv.GetItemId(), err)
		}

		b, err := extractBlob(id, items, lang)
		if err != nil {
			return handlerResult{}, fmt.Errorf("extracting blob. item %v. %w", fv.GetItemId(), err)
		}
		hr.blobs = append(hr.blobs, b)

		str = strings.ReplaceAll(str, "-/media/"+idval+".ashx", "blobref:"+b.blobId.String())
	}

	hr.value = str

	return hr, nil
}

func handleReference(fv data.FieldValueNode, item data.ItemNode, items data.ItemMap, fsetting FieldSettings, lang data.Language) (handlerResult, error) {
	refval := fv.GetValue()
	ids := strings.Split(refval, "|")

	if len(ids) > 1 {
		return handlerResult{}, fmt.Errorf("multiple values not supported for %s (id: %v) on item %s (id: %v)", fv.GetName(), fv.GetFieldId(), item.GetName(), item.GetId())
	}
	if len(ids) == 0 {
		return handlerResult{}, nil
	}

	id := ids[0]
	uid, err := api.TryParseUUID(id)
	if err != nil {
		return handlerResult{}, fmt.Errorf("couldn't parse ref id %s %w", id, err)
	}
	refitem, ok := items[uid]
	if !ok {
		return handlerResult{}, fmt.Errorf("reference not found in map %v. referenced by item %v", uid, item.GetId())
	}

	reft := refitem.GetTemplate()
	if reft == nil {
		log.Println("reference template null", refitem.GetTemplateId(), refitem.GetId())
		return handlerResult{}, nil
	}

	fld := reft.FindField(fsetting.RefField)
	if fld == nil {
		return handlerResult{}, fmt.Errorf("couldn't find field %s on template %s %v", fsetting.RefField, reft.GetName(), reft.GetId())
	}

	reffv := refitem.GetFieldValue(fld.GetId(), lang)
	if reffv == nil {
		log.Printf("no value for ref field %s (referencing %s id: %v) on item %s (id: %v)", reft.GetName(), fld.GetName(), fld.GetId(), refitem.GetName(), refitem.GetId())
		return handlerResult{}, nil
	}

	return ResolveField(reffv, fld, refitem, items, fsetting, lang)
}

func handleImage(fv data.FieldValueNode, item data.ItemNode, items data.ItemMap, fsetting FieldSettings, lang data.Language) (handlerResult, error) {
	val := fv.GetValue()
	if val == "" {
		return handlerResult{}, nil
	}
	g := mediaReg.FindStringSubmatch(val)
	if len(g) != 2 {
		return handlerResult{}, fmt.Errorf("image field not in expected format %s", val)
	}
	id, err := uuid.Parse(g[1])
	if err != nil {
		return handlerResult{}, fmt.Errorf("image field not in expected format %s parsed %s %w", fv.GetValue(), g[1], err)
	}
	b, err := extractBlob(id, items, lang)
	if err != nil {
		return handlerResult{}, fmt.Errorf("extracting blob. %w", err)
	}

	hr := handlerResult{value: "blobref:" + b.blobId.String()}
	hr.blobs = append(hr.blobs, b)

	return hr, nil
}

// image id points to the media library
// will find out the blob data and return that
func extractBlob(imageId uuid.UUID, items data.ItemMap, lang data.Language) (blobResult, error) {
	media, ok := items[imageId]
	if !ok {
		return blobResult{}, fmt.Errorf("referenced image not found %v", imageId)
	}

	blobfld := media.GetTemplate().FindField("Blob")
	extfld := media.GetTemplate().FindField("Extension")
	altfld := media.GetTemplate().FindField("Alt")

	if blobfld == nil || extfld == nil || altfld == nil {
		return blobResult{}, fmt.Errorf("image fields not found. ID: %v Blob: %v  Extension: %v  Alt: %v", imageId, blobfld == nil, extfld == nil, altfld == nil)
	}

	blobidfv := media.GetFieldValue(blobfld.GetId(), lang)
	extfv := media.GetFieldValue(extfld.GetId(), lang)
	altfv := media.GetFieldValue(altfld.GetId(), lang)

	// no actual blob
	if blobidfv == nil {
		return blobResult{}, fmt.Errorf("no blob id field value exists on image %v, field id is %v", imageId, blobfld.GetId())
	}

	blobId, err := uuid.Parse(blobidfv.GetValue())
	if err != nil {
		return blobResult{}, fmt.Errorf("blob field is invalid format %s %w", blobidfv.GetValue(), err)
	}

	b := blobResult{blobId: blobId, name: media.GetName(), ext: extfv.GetValue()}
	b.attrs = append(b.attrs, blobAttr{name: "alt", value: altfv.GetValue()})

	return b, nil
}
