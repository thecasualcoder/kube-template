package cmd

import (
	"bytes"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/thecasualcoder/kube-template/mock"
	v1 "k8s.io/api/core/v1"
	apiV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestParseTemplateFlag(t *testing.T) {
	type output struct {
		firstRet  string
		secondRet string
		err       error
	}

	type testcase struct {
		name string
		in   string
		out  output
	}

	tt := []testcase{
		{name: "happy case", in: "input.tmpl:input.conf", out: output{"input.tmpl", "input.conf", nil}},
		{name: "wrong format with more colons", in: "input.tmpl:input.conf:", out: output{"", "", fmt.Errorf("template flag format is wrong")}},
		{name: "wrong format with less colons", in: "input.tmpl", out: output{"", "", fmt.Errorf("template flag format is wrong")}},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			out1, out2, err := parseTemplateFlag(testCase.in)

			assert.Equal(t, testCase.out.firstRet, out1)
			assert.Equal(t, testCase.out.secondRet, out2)
			assert.Equal(t, testCase.out.err, err)
		})
	}
}

func TestRenderTemplate(t *testing.T) {
	t.Run("should render template with endpoints", func(t *testing.T) {
		source := `
{{- with endpoints "default" "haproxy" }}
{{ .TypeMeta.Kind }}
{{ .TypeMeta.APIVersion }}
{{- range .Subsets }}
{{- $ports := .Ports }}
{{- range .Addresses }}
{{ $ip := .IP }}
{{- range $ports }}
{{ $ip }}:{{ .Port }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}
`
		target := &bytes.Buffer{}
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockManager(ctrl)
		expected := v1.Endpoints{
			TypeMeta: apiV1.TypeMeta{
				Kind:       "Endpoints",
				APIVersion: "v1",
			},
			ObjectMeta: apiV1.ObjectMeta{},
			Subsets: []v1.EndpointSubset{
				{
					Addresses: []v1.EndpointAddress{
						{IP: "10.0.0.100"},
						{IP: "10.0.0.101"},
					},
					Ports: []v1.EndpointPort{
						{
							Name:     "http",
							Port:     8080,
							Protocol: v1.ProtocolTCP,
						},
						{
							Name:     "metrics",
							Port:     9100,
							Protocol: v1.ProtocolTCP,
						},
					},
				},
			},
		}
		m.
			EXPECT().
			Endpoints("default", "haproxy").
			Return(&expected, nil)

		err := renderTemplate(m, source, target)

		assert.NoError(t, err)
		assert.Equal(t, "\nEndpoints\nv1\n\n10.0.0.100:8080\n10.0.0.100:9100\n\n10.0.0.101:8080\n10.0.0.101:9100\n", target.String())
	})

	t.Run("should return error if endpoints gives error", func(t *testing.T) {
		source := `
{{- range endpoints "default" "haproxy" }}
{{ . }}
{{- end }}
`
		target := &bytes.Buffer{}
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockManager(ctrl)
		m.
			EXPECT().
			Endpoints("default", "haproxy").
			Return(nil, fmt.Errorf("endpoints not found"))

		err := renderTemplate(m, source, target)

		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "error calling endpoints: endpoints not found")
		}

		assert.Equal(t, "", target.String())
	})
}
