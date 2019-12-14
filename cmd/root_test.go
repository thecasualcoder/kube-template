package cmd

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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
			out1, out2, err := parseTempalateFlag(testCase.in)

			assert.Equal(t, testCase.out.firstRet, out1)
			assert.Equal(t, testCase.out.secondRet, out2)
			assert.Equal(t, testCase.out.err, err)
		})
	}
}
