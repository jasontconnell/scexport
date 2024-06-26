package main

import (
	"flag"
	"io"
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
	q := flag.Bool("q", false, "quiet mode")
	out := flag.String("output", "", "log output to filename")
	blobs := flag.Bool("blobs", false, "process blobs")
	flast := flag.Bool("lastmod", false, "track with lastmod")
	flag.Parse()

	if *q {
		log.SetOutput(io.Discard)
	}

	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			log.Fatalf("couldn't create log file %s. %v", *out, err)
		}
		log.SetOutput(f)
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

	var since time.Time = process.DefaultModTime
	if *flast {
		since = process.ReadLastMod(*es)
	}

	lang := data.Language(settings.FilterLanguage)

	pkg, err := process.ReadAll(cfg.ConnectionString, cfg.ProtobufLocation, psettings, lang, since)
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
		WriteBlobs:      *blobs,
	}

	if ws.WriteBlobs {
		log.Println("processing blobs in parallel")
		process.ProcessBlobs(cfg.ConnectionString, groups, ws)
	}
	err = process.WriteContent(groups, ws)
	if err != nil {
		log.Fatal("writing contents. ", err)
	}

	if *flast {
		process.WriteLastMod(*es)
	}

	log.Println("Time:", time.Since(start))
}
