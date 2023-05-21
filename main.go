package main

import (
	"flag"
	"io/ioutil"
	"log"
	"time"

	"github.com/jasontconnell/scexport/conf"
	"github.com/jasontconnell/scexport/process"
	"github.com/jasontconnell/sitecore/data"
)

func main() {
	start := time.Now()
	c := flag.String("c", "config.json", "config filename")
	es := flag.String("settings", "", "export settings file")
	q := flag.Bool("q", false, "quiet mode")
	flag.Parse()

	if *q {
		log.SetOutput(ioutil.Discard)
	}

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

	pkg, err := process.ReadAll(cfg.ConnectionString, psettings, lang)
	if err != nil || pkg == nil {
		log.Fatal("reading items. ", err)
	}

	groups, err := process.Resolve(pkg, psettings, lang)
	if err != nil {
		log.Fatal("resolving items. ", err)
	}

	ws := process.WriteSettings{
		ContentFormat:   settings.Output.ContentFormat,
		ContentLocation: settings.Output.ContentLocation,
		BlobLocation:    settings.Output.BlobLocation,
		WriteBlobs:      settings.Output.WriteBlobs,
	}

	if ws.WriteBlobs {
		log.Println("processing blobs in parallel")
		process.ProcessBlobs(cfg.ConnectionString, groups, ws)
	}
	err = process.WriteContent(groups, ws)
	if err != nil {
		log.Fatal("writing contents. ", err)
	}

	log.Println("Time:", time.Since(start))
}
