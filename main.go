package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/jasontconnell/scexport/conf"
	"github.com/jasontconnell/scexport/process"
	"github.com/jasontconnell/sitecore/data"
)

func main() {
	start := time.Now()
	c := flag.String("c", "config.json", "config filename")
	es := flag.String("settings", "", "export settings file")
	flag.Parse()

	cfg, err := conf.LoadConfig(*c)
	if err != nil {
		log.Fatal("couldn't load config. ", err)
	}
	settings, err := conf.LoadExportSettings(*es)
	if err != nil {
		log.Fatal("couldn't load export settings. ", err)
	}

	psettings, err := process.GetSettings(settings)
	if err != nil {
		log.Fatal("problem with settings. ", err)
	}

	lang := data.Language(settings.FilterLanguage)

	items, itemMap, err := process.ReadAll(cfg.ConnectionString, psettings, cfg.GlobalTemplateFilter, lang)
	if err != nil {
		log.Fatal("reading items. ", err)
	}

	groups, err := process.Resolve(items, itemMap, psettings, lang)
	if err != nil {
		log.Fatal("resolving items. ", err)
	}

	log.Println(len(groups), "groups")
	for _, g := range groups {
		log.Println(g.Name, len(g.Items))
	}

	wd, _ := os.Getwd()
	ws := process.WriteSettings{
		ContentFormat:   settings.Output.ContentFormat,
		ContentLocation: settings.Output.ContentLocation,
		BlobLocation:    settings.Output.BlobLocation,
	}

	err = process.Write(wd, groups, ws)
	if err != nil {
		log.Fatal("writing contents. ", err)
	}

	log.Println("Time:", time.Since(start))
}
