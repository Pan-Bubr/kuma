package v3_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("FaultInjectionConfigurer", func() {
	type testCase struct {
		input    *mesh_proto.FaultInjection
		expected string
	}
	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			filterChain, err := NewFilterChainBuilder(envoy.APIV3).
				Configure(HttpConnectionManager("stats", false)).
				Configure(FaultInjection(given.input)).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			actual, err := util_proto.ToYAML(filterChain)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic input", testCase{
			input: &mesh_proto.FaultInjection{
				Sources: []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"tag1": "value1",
							"tag2": "value2",
						},
					},
				},
				Conf: &mesh_proto.FaultInjection_Conf{
					Delay: &mesh_proto.FaultInjection_Conf_Delay{
						Percentage: &wrapperspb.DoubleValue{Value: 50},
						Value:      &durationpb.Duration{Seconds: 5},
					},
				},
			},

			expected: `
            filters:
            - name: envoy.filters.network.http_connection_manager
              typedConfig:
                '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                httpFilters:
                - name: envoy.filters.http.fault
                  typedConfig:
                    '@type': type.googleapis.com/envoy.extensions.filters.http.fault.v3.HTTPFault
                    delay:
                      fixedDelay: 5s
                      percentage:
                        numerator: 50
                    headers:
                    - name: x-kuma-tags
                      safeRegexMatch:
                        googleRe2: {}
                        regex: '.*&tag1=[^&]*value1[,&].*&tag2=[^&]*value2[,&].*'
                - name: envoy.filters.http.router
                statPrefix: stats`,
		}),
		Entry("2 policy selectors", testCase{
			input: &mesh_proto.FaultInjection{
				Sources: []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"tag1": "value1m1",
							"tag2": "value2m1",
						},
					},
					{
						Match: map[string]string{
							"tag1": "value1m2",
							"tag2": "value2m2",
						},
					},
				},
				Conf: &mesh_proto.FaultInjection_Conf{
					Delay: &mesh_proto.FaultInjection_Conf_Delay{
						Percentage: &wrapperspb.DoubleValue{Value: 50},
						Value:      &durationpb.Duration{Seconds: 5},
					},
				},
			},

			expected: `
            filters:
            - name: envoy.filters.network.http_connection_manager
              typedConfig:
                '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                httpFilters:
                - name: envoy.filters.http.fault
                  typedConfig:
                    '@type': type.googleapis.com/envoy.extensions.filters.http.fault.v3.HTTPFault
                    delay:
                      fixedDelay: 5s
                      percentage:
                        numerator: 50
                    headers:
                    - name: x-kuma-tags
                      safeRegexMatch:
                        googleRe2: {}
                        regex: '(.*&tag1=[^&]*value1m1[,&].*&tag2=[^&]*value2m1[,&].*|.*&tag1=[^&]*value1m2[,&].*&tag2=[^&]*value2m2[,&].*)'
                - name: envoy.filters.http.router
                statPrefix: stats`,
		}),
	)
})
