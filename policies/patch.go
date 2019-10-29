package policies

// PatchOperation is used for specifying mutating patches on resources.
// It follows the JSONPatch format (http://jsonpatch.com/)
// This is the format that MutatingWebhookConfigurations require.
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}
