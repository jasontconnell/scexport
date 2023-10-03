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

const defaulthandler string = "default"

var mediaReg *regexp.Regexp = regexp.MustCompile(`<(?:image|file) .*?mediaid="\{([A-F0-9\-]+)\}" ?.*?/>`)
var mediaRteReg *regexp.Regexp = regexp.MustCompile(`src="-\/media\/([a-f0-9]{32})\.ashx`)

type FieldHandler func(
	fv data.FieldValueNode,
	item data.ItemNode,
	pkg *DataPackage,
	fsetting FieldSettings,
	bsetting BlobSettings,
	lang data.Language) (HandlerResult, error)

var fieldHandlers map[string]FieldHandler

func init() {
	fieldHandlers = map[string]FieldHandler{
		"Single-Line Text":      handleString,
		"text":                  handleString,
		"Droplink":              handleReference,
		"Droptree":              handleReference,
		"Treelist":              handleReferenceList,
		"MultiRoot Treelist":    handleReferenceList,
		"Multilist with Search": handleReferenceList,
		"Datetime":              handleString,
		"Rich Text":             handleRichText,
		"Multi-Line Text":       handleRichText,
		"Image":                 handleMedia,
		"File":                  handleMedia,
		"attachment":            handleAttachment,
		defaulthandler:          handleString,
	}
}

func ResolveField(
	fv data.FieldValueNode,
	tfld data.TemplateFieldNode,
	item data.ItemNode,
	pkg *DataPackage,
	fsetting FieldSettings,
	bsetting BlobSettings,
	lang data.Language) (HandlerResult, error) {

	fh, ok := fieldHandlers[tfld.GetType()]
	if !ok {
		fh = fieldHandlers[defaulthandler]
	}

	return fh(fv, item, pkg, fsetting, bsetting, lang)
}

func handleString(
	fv data.FieldValueNode,
	item data.ItemNode,
	pkg *DataPackage,
	fsetting FieldSettings,
	bsetting BlobSettings,
	lang data.Language) (HandlerResult, error) {

	return handlerResult{value: fv.GetValue()}, nil
}

func handleRichText(
	fv data.FieldValueNode,
	item data.ItemNode,
	pkg *DataPackage,
	fsetting FieldSettings,
	bsetting BlobSettings,
	lang data.Language) (HandlerResult, error) {

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
			return handlerResult{}, fmt.Errorf("couldn't parse media id in rte field %s %s. %w", fv.GetName(), fv.GetItemId(), err)
		}

		b, err := extractBlob(id, pkg, bsetting, lang)
		if err != nil {
			return handlerResult{}, fmt.Errorf("handleRichText: extracting blob. item %v. %w", fv.GetItemId(), err)
		}
		hr.blobs = append(hr.blobs, b)

		str = strings.ReplaceAll(str, "-/media/"+idval+".ashx", "blobref:"+b.GetId().String())
	}

	hr.value = formatText(str)

	return hr, nil
}

func handleReferenceList(
	fv data.FieldValueNode,
	item data.ItemNode,
	pkg *DataPackage,
	fsetting FieldSettings,
	bsetting BlobSettings,
	lang data.Language) (HandlerResult, error) {

	refval := fv.GetValue()
	ids := strings.Split(refval, "|")

	if len(ids) == 0 {
		return nil, nil
	}

	list := []Item{}
	for _, id := range ids {
		if id == "" {
			continue
		}

		uid, err := api.TryParseUUID(id)
		if err != nil {
			log.Printf("couldn't parse uuid in item %v field %v value %s. skipping\n", item.GetId(), fv.GetName(), id)
			continue
		}

		refitem, ok := pkg.RefItems[uid]
		if !ok {
			log.Printf("ref item not found in item %v field %v value %s. skipping\n", item.GetId(), fv.GetName(), id)
			continue
		}

		ref, referr := resolveReferenceItem(refitem, pkg, fsetting.RefField, bsetting, lang)
		if referr != nil {
			log.Printf("couldn't get referenced item in list. item %v field %v value %s. skipping\n", item.GetId(), fv.GetName(), id)
			continue
		}

		list = append(list, ref)
	}

	hr := handlerResult{refs: list}

	return hr, nil
}

func handleReference(
	fv data.FieldValueNode,
	item data.ItemNode,
	pkg *DataPackage,
	fsetting FieldSettings,
	bsetting BlobSettings,
	lang data.Language) (HandlerResult, error) {

	return handleReferenceList(fv, item, pkg, fsetting, bsetting, lang)
}

func handleMedia(
	fv data.FieldValueNode,
	item data.ItemNode,
	pkg *DataPackage,
	fsetting FieldSettings,
	bsetting BlobSettings,
	lang data.Language) (HandlerResult, error) {

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
	b, err := extractBlob(id, pkg, bsetting, lang)
	if err != nil {
		return handlerResult{}, fmt.Errorf("handleImage: extracting blob. %w", err)
	}

	hr := handlerResult{id: b.GetId().String(), path: b.GetPath(), value: "blobref:" + b.GetId().String()}
	hr.blobs = append(hr.blobs, b)

	return hr, nil
}

func getRefItemResult(
	id uuid.UUID,
	fv data.FieldValueNode,
	item data.ItemNode,
	pkg *DataPackage,
	fsetting FieldSettings,
	bsetting BlobSettings,
	lang data.Language) (HandlerResult, error) {

	refitem, ok := pkg.RefItems[id]
	if !ok {
		return handlerResult{}, fmt.Errorf("reference not found in map %v. referenced by item %v", id, item.GetId())
	}

	reft := refitem.GetTemplate()
	if reft == nil {
		log.Println("reference template null", refitem.GetTemplateId(), refitem.GetId())
		return handlerResult{}, nil
	}

	if fsetting.RefField != ItemNameOutputField {
		fld := reft.FindField(fsetting.RefField)
		if fld == nil {
			return handlerResult{}, fmt.Errorf("couldn't find field %s on template %s %v", fsetting.RefField, reft.GetName(), reft.GetId())
		}

		reffv := refitem.GetFieldValue(fld.GetId(), lang)
		if reffv == nil {
			log.Printf("no value for ref field %s (referencing %s id: %v) on item %s (id: %v)", reft.GetName(), fld.GetName(), fld.GetId(), refitem.GetName(), refitem.GetId())
			return handlerResult{}, nil
		}
		hr, err := ResolveField(reffv, fld, refitem, pkg, fsetting, bsetting, lang)
		return hr, err
	} else {
		return handlerResult{id: refitem.GetId().String(), path: refitem.GetPath(), value: refitem.GetName()}, nil
	}
}

// image id points to the media library
// will find out the blob data and return that
func extractBlob(mediaId uuid.UUID, pkg *DataPackage, bsetting BlobSettings, lang data.Language) (BlobResult, error) {
	media, ok := pkg.RefItems[mediaId]
	if !ok {
		return blobResult{}, fmt.Errorf("referenced blob not found %v", mediaId)
	}

	blobfld := media.GetTemplate().FindField("Blob")
	extfld := media.GetTemplate().FindField("Extension")

	if blobfld == nil || extfld == nil {
		return blobResult{}, fmt.Errorf("blob fields not found. ID: %v Blob: %v  Extension: %v", mediaId, blobfld == nil, extfld == nil)
	}

	blobidfv := media.GetFieldValue(blobfld.GetId(), lang)
	extfv := media.GetFieldValue(extfld.GetId(), lang)

	attrs := []Attr{}
	for _, cfld := range bsetting.CustomFields {
		fld := media.GetTemplate().GetField(cfld)
		if fld != nil {
			fldval := media.GetFieldValue(fld.GetId(), lang)
			if fldval != nil {
				attrs = append(attrs, Attr{Name: fld.GetName(), Value: fldval.GetValue()})
			}
		}
	}

	// no actual blob
	if blobidfv == nil {
		return blobResult{}, fmt.Errorf("no blob id field value exists on media item %v, field id is %v", mediaId, blobfld.GetId())
	}

	blobId, err := uuid.Parse(blobidfv.GetValue())
	if err != nil {
		return blobResult{}, fmt.Errorf("blob field is invalid format %s %w", blobidfv.GetValue(), err)
	}

	b := blobResult{blobId: blobId, name: media.GetName(), ext: extfv.GetValue(), path: media.GetPath(), attrs: attrs}

	return b, nil
}

func handleAttachment(
	fv data.FieldValueNode,
	item data.ItemNode,
	pkg *DataPackage,
	fsetting FieldSettings,
	bsetting BlobSettings,
	lang data.Language) (HandlerResult, error) {

	val := fv.GetValue()
	// attachment field just has a blob id
	blobId, err := api.TryParseUUID(val)

	if val == "" || err != nil {
		return nil, fmt.Errorf("parsing blob id from attachment. item id: %v field id: %v field value %s. %w", item.GetId(), fv.GetFieldId(), val, err)
	}

	blobfld := item.GetTemplate().FindField("Blob")
	extfld := item.GetTemplate().FindField("Extension")
	if blobfld == nil || extfld == nil {
		return nil, fmt.Errorf("blob fields not found. ID: %v Blob: %v  Extension: %v ", item.GetId(), blobfld == nil, extfld == nil)
	}
	extfv := item.GetFieldValue(extfld.GetId(), lang)
	ext := ""
	if extfv != nil {
		ext = extfv.GetValue()
	}

	b := blobResult{blobId: blobId, name: item.GetName(), ext: ext, path: item.GetPath()}

	hr := handlerResult{value: "blobref:" + b.blobId.String()}
	hr.blobs = append(hr.blobs, b)

	return hr, nil
}
