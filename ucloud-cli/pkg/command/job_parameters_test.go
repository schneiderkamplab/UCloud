package command

import (
	"testing"
	orcapi "ucloud.dk/shared/pkg/orchestrators"
)

func TestFileParameters(t *testing.T) {
	for _, kind := range []orcapi.ApplicationParameterType{orcapi.ApplicationParameterTypeInputFile, orcapi.ApplicationParameterTypeInputDirectory} {
		param := orcapi.ApplicationParameter{Name: "initScript", Type: kind}
		value, err := parseParameterValue(param, "/998037/.home.sh")
		if err != nil || value.Type != orcapi.AppParameterValueTypeFile || value.Path != "/998037/.home.sh" {
			t.Fatalf("invalid file parameter: %v", err)
		}
		for _, path := range []string{"", "/", "relative.sh"} {
			if _, err := parseParameterValue(param, path); err == nil {
				t.Errorf("accepted invalid file path %q", path)
			}
		}
	}
}
