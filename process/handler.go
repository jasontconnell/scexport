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

var mediaReg *regexp.Regexp = regexp.MustCompile(`<image .*?mediaid="\{([A-F0-9\-]+)\}" ?.*?/>`)
var mediaRteReg *regexp.Regexp = regexp.MustCompile(`src="-\/media\/([a-f0-9]{32})\.ashx`)

type FieldHandler func(
	fv data.FieldValueNode,
	item data.ItemNode,
	pkg *DataPackage,
	fsetting FieldSettings,
	lang data.Language) (HandlerResult, error)

var fieldHandlers map[string]FieldHandler

func init() {
	fieldHandlers = map[string]FieldHandler{
		"Single-Line Text":      handleString,
		"Droplink":              handleReference,
		"Treelist":              handleReferenceList,
		"MultiRoot Treelist":    handleReferenceList,
		"Multilist with Search": handleReferenceList,
		"Datetime":              handleString,
		"Rich Text":             handleRichText,
		"Multi-Line Text":       handleRichText,
		"Image":                 handleImage,
		"attachment":            handleAttachment,
	}
}

func ResolveField(
	fv data.FieldValueNode,
	tfld data.TemplateFieldNode,
	item data.ItemNode,
	pkg *DataPackage,
	fsetting FieldSettings,
	lang data.Language) (HandlerResult, error) {

	fh, ok := fieldHandlers[tfld.GetType()]
	if !ok {
		return handlerResult{}, fmt.Errorf("no handler for %s", tfld.GetType())
	}
	return fh(fv, item, pkg, fsetting, lang)
}

func handleString(
	fv data.FieldValueNode,
	item data.ItemNode,
	pkg *DataPackage,
	fsetting FieldSettings,
	lang data.Language) (HandlerResult, error) {

	return handlerResult{value: fv.GetValue()}, nil
}

func handleRichText(
	fv data.FieldValueNode,
	item data.ItemNode,
	pkg *DataPackage,
	fsetting FieldSettings,
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

		b, err := extractBlob(id, pkg, lang)
		if err != nil {
			return handlerResult{}, fmt.Errorf("handleRichText: extracting blob. item %v. %w", fv.GetItemId(), err)
		}
		hr.blobs = append(hr.blobs, b)

		str = strings.ReplaceAll(str, "-/media/"+idval+".ashx", "blobref:"+b.GetId().String())
	}

	hr.value = str

	return hr, nil
}

func handleReferenceList(
	fv data.FieldValueNode,
	item data.ItemNode,
	pkg *DataPackage,
	fsetting FieldSettings,
	lang data.Language) (HandlerResult, error) {

	refval := fv.GetValue()
	ids := strings.Split(refval, "|")

	if len(ids) == 0 {
		return nil, nil
	}

	if v, ok := fsetting.Properties["blob"].(bool); ok && v {
		return handleReference(fv, item, pkg, fsetting, lang)
	}

	usename := fsetting.RefField == ItemNameField

	finter, ok := fsetting.Properties["fields"].([]interface{})
	if !ok && !usename && fsetting.RefField == "" {
		return nil, fmt.Errorf("can't get fields from properties properties: %v  properties.fields: %v", fsetting.Properties, fsetting.Properties["fields"])
	}

	fields := []string{}
	for _, f := range finter {
		fields = append(fields, f.(string))
	}

	if fsetting.RefField != "" && !usename {
		fields = append(fields, fsetting.RefField)
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

		var ref Item
		var referr error
		if !usename {
			ref, referr = resolveReferenceItem(refitem, pkg, fields, lang)
		} else {
			log.Println("use name", refitem.GetId(), refitem.GetPath())
			ref = Item{ID: refitem.GetId().String(), Name: refitem.GetName(), Path: refitem.GetPath()}
		}
		if referr != nil {
			log.Printf("couldn't get referenced item in list. item %v field %v value %s. skipping\n", item.GetId(), fv.GetName(), id)
			continue // return nil, fmt.Errorf("couldn't get referenced item in list. item %v field %v value %s. %w", item.GetId(), fv.GetName(), id, referr)
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
	lang data.Language) (HandlerResult, error) {

	refval := fv.GetValue()

	if len(refval) == 0 {
		return handlerResult{}, nil
	}

	uid, err := api.TryParseUUID(refval)
	if err != nil {
		return handlerResult{}, fmt.Errorf("couldn't parse ref id %s %w", refval, err)
	}

	return getRefItemResult(uid, fv, item, pkg, fsetting, lang)
}

func handleImage(
	fv data.FieldValueNode,
	item data.ItemNode,
	pkg *DataPackage,
	fsetting FieldSettings,
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
	b, err := extractBlob(id, pkg, lang)
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

	isblobrefval, ok := fsetting.Properties["blob"]
	var blobref bool
	if ok {
		blobref, ok = isblobrefval.(bool)
		if !ok {
			return handlerResult{}, fmt.Errorf("invalid value for blob in properties. %v. expecting bool", isblobrefval)
		}
	}

	if !blobref {
		if fsetting.RefField != ItemNameField {
			fld := reft.FindField(fsetting.RefField)
			if fld == nil {
				return handlerResult{}, fmt.Errorf("couldn't find field %s on template %s %v", fsetting.RefField, reft.GetName(), reft.GetId())
			}

			reffv := refitem.GetFieldValue(fld.GetId(), lang)
			if reffv == nil {
				log.Printf("no value for ref field %s (referencing %s id: %v) on item %s (id: %v)", reft.GetName(), fld.GetName(), fld.GetId(), refitem.GetName(), refitem.GetId())
				return handlerResult{}, nil
			}
			hr, err := ResolveField(reffv, fld, refitem, pkg, fsetting, lang)
			return hr, err
		} else {
			return handlerResult{id: refitem.GetId().String(), path: refitem.GetPath(), value: refitem.GetName()}, nil
		}
	} else {
		// blob ref is for when a treelist or something references files or images
		b, err := extractBlob(id, pkg, lang)
		if err != nil {
			return handlerResult{}, fmt.Errorf("extract blob referenced by %v. %w", id, err)
		}
		return handlerResult{id: b.GetId().String(), path: b.GetPath(), value: "blobref:" + b.GetId().String(), blobs: []BlobResult{b}}, nil
	}
}

// image id points to the media library
// will find out the blob data and return that
func extractBlob(mediaId uuid.UUID, pkg *DataPackage, lang data.Language) (BlobResult, error) {
	media, ok := pkg.RefItems[mediaId]
	if !ok {
		return blobResult{}, fmt.Errorf("referenced blob not found %v", mediaId)
	}

	blobfld := media.GetTemplate().FindField("Blob")
	extfld := media.GetTemplate().FindField("Extension")
	altfld := media.GetTemplate().FindField("Alt")

	if blobfld == nil || extfld == nil {
		return blobResult{}, fmt.Errorf("blob fields not found. ID: %v Blob: %v  Extension: %v  Alt: %v", mediaId, blobfld == nil, extfld == nil, altfld == nil)
	}

	blobidfv := media.GetFieldValue(blobfld.GetId(), lang)
	extfv := media.GetFieldValue(extfld.GetId(), lang)

	alt := ""
	if altfld != nil {
		altfv := media.GetFieldValue(altfld.GetId(), lang)
		if altfv != nil {
			alt = altfv.GetValue()
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

	b := blobResult{blobId: blobId, name: media.GetName(), ext: extfv.GetValue(), path: media.GetPath()}
	b.attrs = append(b.attrs, Attr{Name: "alt", Value: alt})

	return b, nil
}

func handleAttachment(
	fv data.FieldValueNode,
	item data.ItemNode,
	pkg *DataPackage,
	fsetting FieldSettings,
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
