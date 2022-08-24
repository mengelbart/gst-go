package gstreamer

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0

#include "gst.h"

*/
import "C"
import (
	"errors"
	"log"
	"strings"
	"sync"
	"unsafe"
)

type Buffer struct {
	Bytes    []byte
	Duration int
}

type BufferHandler func(Buffer)
type EOSHandler func()
type ErrorHandler func(error)

var pipelines = map[int]*Pipeline{}
var pipelinesLock sync.Mutex

type Pipeline struct {
	launch     string
	id         int
	gstElement *C.GstElement
	gMainLoop  *C.GMainLoop
	closed     chan struct{}
	bufferCB   BufferHandler
	eosCB      EOSHandler
	errCB      ErrorHandler
}

func NewPipeline(launch string) (*Pipeline, error) {
	launchStrC := C.CString(launch)
	defer C.free(unsafe.Pointer(launchStrC))

	pipelinesLock.Lock()
	defer pipelinesLock.Unlock()

	p := &Pipeline{
		launch:     launch,
		id:         len(pipelines),
		gstElement: C.create_pipeline(launchStrC),
		gMainLoop:  C.create_mainloop(),
		closed:     make(chan struct{}),
		bufferCB: func(Buffer) {
		},
		eosCB: func() {
		},
		errCB: func(err error) {
		},
	}
	pipelines[p.id] = p
	if strings.Contains(launch, "appsink") {
		p.linkAppsink()
	}
	return p, nil
}

func (p *Pipeline) SetBufferHandler(h BufferHandler) {
	p.bufferCB = h
}

func (p *Pipeline) SetEOSHandler(h EOSHandler) {
	p.eosCB = h
}

func (p *Pipeline) SetErrorHandler(h ErrorHandler) {
	p.errCB = h
}

func (p *Pipeline) Write(buf []byte) (int, error) {
	n := len(buf)
	b := C.CBytes(buf)
	defer C.free(b)
	C.push_buffer(p.gstElement, b, C.int(len(buf)))
	return n, nil
}

func (p *Pipeline) String() string {
	return p.launch
}

func (p *Pipeline) DumpPipelineGraph(filename string) {
	filenameC := C.CString(filename)
	defer C.free(unsafe.Pointer(filenameC))

	C.dump_pipeline_graph(p.gstElement, filenameC)
}

func (p *Pipeline) linkAppsink() {
	C.link_appsink(p.gstElement, C.int(p.id))
}

func (p *Pipeline) Start() {
	go C.start_mainloop(p.gMainLoop)
	C.start_pipeline(p.gstElement, C.int(p.id))
}

func (p *Pipeline) Close() error {
	close(p.closed)
	C.stop_pipeline(p.gstElement)
	C.destroy_pipeline(p.gstElement)
	C.stop_mainloop(p.gMainLoop)
	return nil
}

func (p *Pipeline) SetPropertyUint(name string, prop string, value uint) {
	cName := C.CString(name)
	cProp := C.CString(prop)
	cValue := C.uint(value)

	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cProp))

	C.set_property_uint(p.gstElement, cName, cProp, cValue)
}

func (p *Pipeline) GetPropertyUint(name string, prop string) uint {
	cName := C.CString(name)
	cProp := C.CString(prop)

	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cProp))

	return uint(C.get_property_uint(p.gstElement, cName, cProp))
}

//export goHandlePipelineBuffer
func goHandlePipelineBuffer(buffer unsafe.Pointer, bufferLen C.int, duration C.int, pipelineID C.int) {
	pipelinesLock.Lock()
	pipeline, ok := pipelines[int(pipelineID)]
	pipelinesLock.Unlock()
	defer C.free(buffer)

	if !ok {
		log.Printf("no pipeline with ID %v, discarding buffer", int(pipelineID))
		return
	}

	select {
	case <-pipeline.closed:
		return
	default:
	}

	pipeline.bufferCB(Buffer{
		Bytes:    C.GoBytes(buffer, bufferLen),
		Duration: int(duration),
	})
}

//export goHandleBusCall
func goHandleBusCall(pipelineID C.int, signal C.int, message *C.char) {
	pipelinesLock.Lock()
	pipeline, ok := pipelines[int(pipelineID)]
	pipelinesLock.Unlock()
	if !ok {
		log.Printf("no pipeline with ID %v, discarding EOS", int(pipelineID))
		return
	}
	switch signal {
	case 0:
		pipeline.eosCB()

	case 1:
		msg := C.GoString(message)
		pipeline.errCB(errors.New(msg))
	}
}
