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

package policies

import "fmt"

// ResourceViolation contains information needed to report and track violations, as
// well as checking for exemptions
type ResourceViolation struct {
	ResourceName string
	ResourceKind string
	Namespace    string
	Violation    string
	Policy       string
	Error        error
}

func (r ResourceViolation) HumanString() string {
	return fmt.Sprintf("%s %s had violation: %s", r.ResourceKind, r.ResourceName, r.Violation)
}
