package process

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jasontconnell/sitecore/api"
	"github.com/jasontconnell/sitecore/data"
)

type handlerResult struct {
	value string
	blobs []uuid.UUID
	html  bool
}

type FieldHandler func(fv data.FieldValueNode, item data.ItemNode, items data.ItemMap, fsetting FieldSettings, lang data.Language) (handlerResult, error)

var fieldHandlers map[string]FieldHandler

func init() {
	fieldHandlers = map[string]FieldHandler{
		"Single-Line Text": handleString,
		"Droplink":         handleReference,
		"Datetime":         handleString,
		"Rich Text":        handleRichText,
	}
}

func ResolveField(fv data.FieldValueNode, tfld data.TemplateFieldNode, item data.ItemNode, items data.ItemMap, fsetting FieldSettings, lang data.Language) (handlerResult, error) {
	fmt.Println(tfld.GetType())
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
	return handlerResult{value: fv.GetValue(), html: true}, nil
}

func handleReference(fv data.FieldValueNode, item data.ItemNode, items data.ItemMap, fsetting FieldSettings, lang data.Language) (handlerResult, error) {
	refval := fv.GetValue()
	ids := strings.Split(refval, "|")
	refitems := []data.ItemNode{}
	for _, id := range ids {
		uid, err := api.TryParseUUID(id)
		if err != nil {
			return handlerResult{}, fmt.Errorf("couldn't parse ref id %s %w", id, err)
		}

		if refitem, ok := items[uid]; ok {
			refitems = append(refitems, refitem)
		} else {
			fmt.Println("not found in map", uid)
		}
	}

	val := ""
	for _, refitem := range refitems {
		reft := refitem.GetTemplate()
		if reft == nil {
			fmt.Println("reference template null", refitem.GetTemplateId(), refitem.GetId())
			continue
		}

		fld := reft.FindField(fsetting.RefField)
		if fld == nil {
			return handlerResult{}, fmt.Errorf("couldn't find field %s on template %s %v", fsetting.RefField, reft.GetName(), reft.GetId())
		}

		refval := refitem.GetFieldValue(fld.GetId(), lang)

		if refval == nil {
			fmt.Println("no value for ref field", fld.GetId(), reft.GetName())
			continue
		}

		val += refval.GetValue() + ", "
	}

	return handlerResult{value: strings.TrimRight(val, ", ")}, nil
}
