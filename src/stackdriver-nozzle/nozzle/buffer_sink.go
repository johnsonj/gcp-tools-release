package nozzle

import "github.com/cloudfoundry/sonde-go/events"

type buffer struct {
	messages chan *events.Envelope
}

const bufferSize = 10000

func NewSinkBuffer(destination Sink) Sink {
	b := &buffer{make(chan *events.Envelope, bufferSize)}

	go func() {
		for envelope := range b.messages {
			destination.Receive(envelope)
		}
	}()

	return b
}

func (b *buffer) Receive(event *events.Envelope) error {
	b.messages <- event
	return nil
}
