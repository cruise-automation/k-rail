package policies

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestMutateEmptyDirSizeLimit(t *testing.T) {
	specs := map[string]struct {
		src string
		exp MutateEmptyDirSizeLimit
	}{

		"all good": {
			src: `
mutate_empty_dir_size_limit:
  maximum_size_limit: "1Gi"
  default_size_limit: "512Mi"
`,
			exp: MutateEmptyDirSizeLimit{
				MaximumSizeLimit: "1Gi",
				DefaultSizeLimit: "512Mi",
			},
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			var cfg Config
			err := yaml.Unmarshal([]byte(spec.src), &cfg)
			if err != nil {
				t.Fatalf("unexpected error: %+v", err)
			}
			if exp, got := spec.exp, cfg.MutateEmptyDirSizeLimit; !reflect.DeepEqual(exp, got) {
				t.Errorf("expected %v but got %v", exp, got)
			}
		})
	}

}
