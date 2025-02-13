package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/app/kuma-prometheus-sd/cmd"
)

var _ = Describe("root", func() {
	It("should be possible to run `kuma-prometheus-sd` without a sub-command", func() {
		// given
		cmd := NewRootCmd()
		cmd.SetArgs([]string{})
		// when
		err := cmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())
	})
})
