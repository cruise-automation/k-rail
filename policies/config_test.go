package policies

import (
	"reflect"
	"testing"

	apiresource "k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/yaml"
)

func TestMutateEmptyDirSizeLimit(t *testing.T) {
	specs := map[string]struct {
		src    string
		exp    *MutateEmptyDirSizeLimit
		expErr bool
	}{

		"all good": {
			src: `
mutate_empty_dir_size_limit:
  maximum_size_limit: "1Gi"
  default_size_limit: "512Mi"
`,
			exp: &MutateEmptyDirSizeLimit{
				MaximumSizeLimit: apiresource.MustParse("1Gi"),
				DefaultSizeLimit: apiresource.MustParse("512Mi"),
			},
		},
		"default > max": {
			src: `
mutate_empty_dir_size_limit:
  maximum_size_limit: "1Gi"
  default_size_limit: "2Gi"
`,
			expErr: true,
		},
		"default not set": {
			src: `
mutate_empty_dir_size_limit:
  maximum_size_limit: "1Gi"
`,
			expErr: true,
		},
		"max not set": {
			src: `
mutate_empty_dir_size_limit:
  default_size_limit: "2Gi"
`,
			expErr: true,
		},
		"unsupported type": {
			src: `
mutate_empty_dir_size_limit:
  default_size_limit: "2ALX"
  maximum_size_limit: "2ALX"
`,
			expErr: true,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			var cfg Config
			switch err := yaml.Unmarshal([]byte(spec.src), &cfg); {
			case spec.expErr && err != nil:
				return
			case spec.expErr:
				t.Fatal("expected error")
			case !spec.expErr && err != nil:
				t.Fatalf("unexpected error: %+v", err)
			}
			if exp, got := *spec.exp, cfg.MutateEmptyDirSizeLimit; !reflect.DeepEqual(exp, got) {
				t.Errorf("expected %v but got %v", exp, got)
			}
		})
	}

}
