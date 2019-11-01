// Copyright 2019 Cruise LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    https://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package ingress

package resource

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
)

type object interface {
	metav1.Object
	runtime.Object
}

func decodeObject(raw []byte, object object) error {
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, object); err != nil {
		return err
	}
	return nil
}

// GetResourceName attempts to get the best name for a resource
func GetResourceName(meta metav1.ObjectMeta) (name string) {
	// Attempt to get the owner controller's resource name.
	// This name is the high level resource that the user is working with.
	for _, owner := range meta.OwnerReferences {
		if owner.Controller != nil && *owner.Controller == true {
			if len(owner.Name) > 0 {
				name = owner.Name
				return
			}
		}
	}

	// Attempt to get the object's name
	if len(meta.Name) > 0 {
		name = meta.Name
		return
	}

	// Attempt to get the name label
	if val, ok := meta.Labels["name"]; ok {
		name = val
		return
	}

	return
}
