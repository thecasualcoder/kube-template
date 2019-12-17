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

func TestRenderTemplate(t *testing.T) {
	t.Run("should render template with endpoints", func(t *testing.T) {
		source := `{{- with endpoints "default" "haproxy" -}}
endpoints:{{ range .Subsets -}}
{{- $ports := .Ports -}}
{{- range .Addresses -}}
{{ $ip := .IP -}}
{{- range $ports }}
- {{ $ip }}:{{ .Port }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}
`
		expected := `endpoints:
- 10.0.0.100:8080
- 10.0.0.100:9100
- 10.0.0.101:8080
- 10.0.0.101:9100
`
		target := &bytes.Buffer{}
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockManager(ctrl)
		endpoints := v1.Endpoints{
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
			Return(&endpoints, nil)

		err := renderTemplate(m, source, target)

		assert.NoError(t, err)
		assert.Equal(t, expected, target.String())
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

	t.Run("should render template with pods", func(t *testing.T) {
		source := `{{- with pods "default" "foo=bar" -}}
pods:
{{- range .Items }}
  - {{ .Name }}:{{ .Status.PodIP }}
{{- end }}
{{- end }}
`
		expected := `pods:
  - pod-1:10.0.0.100
  - pod-2:10.0.0.101
`
		target := &bytes.Buffer{}
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockManager(ctrl)
		pods := v1.PodList{
			TypeMeta: apiV1.TypeMeta{
				Kind:       "Endpoints",
				APIVersion: "v1",
			},
			Items: []v1.Pod{
				{
					ObjectMeta: apiV1.ObjectMeta{
						Name: "pod-1",
					},
					Status: v1.PodStatus{
						PodIP: "10.0.0.100",
					},
				},
				{
					ObjectMeta: apiV1.ObjectMeta{
						Name: "pod-2",
					},
					Status: v1.PodStatus{
						PodIP: "10.0.0.101",
					},
				},
			},
		}
		m.
			EXPECT().
			PodsForLabels("default", "foo=bar").
			Return(&pods, nil)

		err := renderTemplate(m, source, target)

		assert.NoError(t, err)
		assert.Equal(t, expected, target.String())
	})
}
