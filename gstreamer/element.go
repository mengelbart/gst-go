package gstreamer

import (
	"fmt"
	"strings"
)

type Elements []*Element

func (ee Elements) Build() string {
	res := make([]string, len(ee))
	for i, e := range ee {
		res[i] = e.String()
	}
	return strings.Join(res, " ! ")
}

type Element struct {
	name    string
	options map[string]interface{}
}

type ElementOption func(*Element)

func Set(key string, value interface{}) ElementOption {
	return func(e *Element) {
		e.options[key] = value
	}
}

func NewElement(name string, opts ...ElementOption) *Element {
	e := &Element{
		name:    name,
		options: map[string]interface{}{},
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Element) String() string {
	compiledOptions := make([]string, len(e.options))
	i := 0
	for k, v := range e.options {
		compiledOptions[i] = fmt.Sprintf("%v=%v", k, v)
		i++
	}
	return fmt.Sprintf("%v %v", e.name, strings.Join(compiledOptions, " "))
}
