package worker

import (
	"context"
)

type OutputSet struct {
	outputs map[string]*Output
}

func NewOutputSet() *OutputSet {
	s := &OutputSet{}
	s.outputs = make(map[string]*Output)
	return s
}

func (o *OutputSet) Add(ctx context.Context, name string, endpoint string) error {
	if _, ok := o.outputs[name]; !ok {
		o.outputs[name] = &(Output{name: name})
	}

	return o.outputs[name].Add(ctx, endpoint)

}

func (o *OutputSet) Send(name string, msg []uint8) error {
	return o.outputs[name].Send(msg)
}
