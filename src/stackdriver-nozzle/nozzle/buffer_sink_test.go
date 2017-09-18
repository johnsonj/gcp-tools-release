package nozzle_test

import (
	. "github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/nozzle"

	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/mocks"
	"github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BufferSink", func() {

	var (
		dest    *mocks.Sink
		subject Sink
	)

	BeforeEach(func() {
		dest = &mocks.Sink{}
		subject = NewSinkBuffer(dest)
	})

	It("delivers messages", func() {
		eventType := events.Envelope_HttpStartStop
		event := events.Envelope{EventType: &eventType}
		err := subject.Receive(&event)
		Expect(err).NotTo(HaveOccurred())

		Eventually(dest.HandledEnvelopes).Should(ContainElement(event))
	})
})
