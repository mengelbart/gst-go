# Gstreamer Wrapper for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/mengelbart/gst-go.svg)](https://pkg.go.dev/github.com/mengelbart/gst-go)

*gst-go* is a very simple CGO wrapper for using Gstreamer in Go.

An example application:

```golang
package main

import (
	"log"

	"github.com/mengelbart/gst-go"
)

func main() {
	gst.GstInit()
	defer gst.GstDeinit()

	ml := gst.NewMainLoop()
	p, err := gst.NewPipeline("videotestsrc ! autovideosink")
	if err != nil {
		log.Fatal(err)
	}
	p.Start()
	p.SetEOSHandler(func() {
		p.Stop()
		ml.Stop()
	})
	p.SetErrorHandler(func(err error) {
		log.Println(err)
		p.Stop()
		ml.Stop()
	})
	ml.Run()
}

```
