package lifecycle_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/gordon"
)

var _ = Describe("A container with a grace time", func() {
	var handle string

	BeforeEach(func() {
		err := runner.Stop()
		Expect(err).ToNot(HaveOccurred())

		err = runner.Start("--containerGraceTime", "5s")
		Expect(err).ToNot(HaveOccurred())

		res, err := client.Create(nil)
		Expect(err).ToNot(HaveOccurred())

		handle = res.GetHandle()
	})

	AfterEach(func() {
		err := runner.Stop()
		Expect(err).ToNot(HaveOccurred())

		err = runner.Start()
		Expect(err).ToNot(HaveOccurred())
	})

	Context("when a request takes longer than the grace time", func() {
		It("is not destroyed after the request is over", func() {
			_, _, err := client.Run(handle, "sleep 6", gordon.ResourceLimits{})
			Expect(err).ToNot(HaveOccurred())

			_, err = client.Info(handle)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("when no requests are made for longer than the grace time", func() {
		It("is destroyed", func() {
			time.Sleep(6 * time.Second)

			_, err := client.Info(handle)
			Expect(err).To(HaveOccurred())
		})
	})
})
