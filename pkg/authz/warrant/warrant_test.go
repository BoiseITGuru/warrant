// Copyright 2023 Forerunner Labs, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authz

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestToWarrantDirectWarrantSpec(t *testing.T) {
	spec := WarrantSpec{
		ObjectType: "permission",
		ObjectId:   "test",
		Relation:   "member",
		Subject: &SubjectSpec{
			ObjectType: "user",
			ObjectId:   "user-A",
		},
	}
	expectedWarrant := &Warrant{
		ObjectType:  "permission",
		ObjectId:    "test",
		Relation:    "member",
		SubjectType: "user",
		SubjectId:   "user-A",
	}
	actualWarrant, err := spec.ToWarrant()
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(actualWarrant, expectedWarrant) {
		t.Fatalf("Expected warrant to be %v, but it was %v", expectedWarrant, actualWarrant)
	}
}

func TestToWarrantDirectWarrantSpecWithPolicy(t *testing.T) {
	policy := Policy("tenant == 101")
	spec := WarrantSpec{
		ObjectType: "permission",
		ObjectId:   "test",
		Relation:   "member",
		Subject: &SubjectSpec{
			ObjectType: "user",
			ObjectId:   "user-A",
		},
		Policy: policy,
	}
	expectedWarrant := &Warrant{
		ObjectType:  "permission",
		ObjectId:    "test",
		Relation:    "member",
		SubjectType: "user",
		SubjectId:   "user-A",
		Policy:      policy,
		PolicyHash:  policy.Hash(),
	}
	actualWarrant, err := spec.ToWarrant()
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(actualWarrant, expectedWarrant) {
		t.Fatalf("Expected warrant to be %v, but it was %v", expectedWarrant, actualWarrant)
	}
}

func TestToWarrantDirectWarrantSpecWithContext(t *testing.T) {
	policy := Policy(`tenant == "101"`)
	spec := WarrantSpec{
		ObjectType: "permission",
		ObjectId:   "test",
		Relation:   "member",
		Subject: &SubjectSpec{
			ObjectType: "user",
			ObjectId:   "user-A",
		},
		Context: map[string]string{
			"tenant": "101",
		},
	}
	expectedWarrant := &Warrant{
		ObjectType:  "permission",
		ObjectId:    "test",
		Relation:    "member",
		SubjectType: "user",
		SubjectId:   "user-A",
		Policy:      policy,
		PolicyHash:  policy.Hash(),
	}
	actualWarrant, err := spec.ToWarrant()
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(actualWarrant, expectedWarrant) {
		t.Fatalf("Expected warrant to be %v, but it was %v", expectedWarrant, actualWarrant)
	}
}

func TestToWarrantIndirectWarrantSpec(t *testing.T) {
	spec := WarrantSpec{
		ObjectType: "permission",
		ObjectId:   "test",
		Relation:   "member",
		Subject: &SubjectSpec{
			ObjectType: "role",
			ObjectId:   "admin",
			Relation:   "member",
		},
	}
	expectedWarrant := &Warrant{
		ObjectType:      "permission",
		ObjectId:        "test",
		Relation:        "member",
		SubjectType:     "role",
		SubjectId:       "admin",
		SubjectRelation: "member",
	}
	actualWarrant, err := spec.ToWarrant()
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(actualWarrant, expectedWarrant) {
		t.Fatalf("Expected warrant to be %v, but it was %v", expectedWarrant, actualWarrant)
	}
}

func TestToWarrantIndirectWarrantSpecWithPolicy(t *testing.T) {
	policy := Policy("tenant == 101")
	spec := WarrantSpec{
		ObjectType: "permission",
		ObjectId:   "test",
		Relation:   "member",
		Subject: &SubjectSpec{
			ObjectType: "role",
			ObjectId:   "admin",
			Relation:   "member",
		},
		Policy: policy,
	}
	expectedWarrant := &Warrant{
		ObjectType:      "permission",
		ObjectId:        "test",
		Relation:        "member",
		SubjectType:     "role",
		SubjectId:       "admin",
		SubjectRelation: "member",
		Policy:          policy,
		PolicyHash:      policy.Hash(),
	}
	actualWarrant, err := spec.ToWarrant()
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(actualWarrant, expectedWarrant) {
		t.Fatalf("Expected warrant to be %v, but it was %v", expectedWarrant, actualWarrant)
	}
}

func TestToWarrantIndirectWarrantSpecWithContext(t *testing.T) {
	policy := Policy(`tenant == "101"`)
	spec := WarrantSpec{
		ObjectType: "permission",
		ObjectId:   "test",
		Relation:   "member",
		Subject: &SubjectSpec{
			ObjectType: "role",
			ObjectId:   "admin",
			Relation:   "member",
		},
		Context: map[string]string{
			"tenant": "101",
		},
	}
	expectedWarrant := &Warrant{
		ObjectType:      "permission",
		ObjectId:        "test",
		Relation:        "member",
		SubjectType:     "role",
		SubjectId:       "admin",
		SubjectRelation: "member",
		Policy:          policy,
		PolicyHash:      policy.Hash(),
	}
	actualWarrant, err := spec.ToWarrant()
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(actualWarrant, expectedWarrant) {
		t.Fatalf("Expected warrant to be %v, but it was %v", expectedWarrant, actualWarrant)
	}
}

func TestToMapDirectWarrantSpec(t *testing.T) {
	spec := WarrantSpec{
		ObjectType: "permission",
		ObjectId:   "test",
		Relation:   "member",
		Subject: &SubjectSpec{
			ObjectType: "user",
			ObjectId:   "user-A",
		},
	}
	expectedMap := map[string]interface{}{
		"objectType": "permission",
		"objectId":   "test",
		"relation":   "member",
		"subject":    spec.Subject.ToMap(),
	}
	actualMap := spec.ToMap()
	if !cmp.Equal(actualMap, expectedMap) {
		t.Fatalf("Expected map to be %v, but it was %v", expectedMap, actualMap)
	}
}

func TestToMapDirectWarrantSpecWithPolicy(t *testing.T) {
	policy := Policy("tenant == 101")
	spec := WarrantSpec{
		ObjectType: "permission",
		ObjectId:   "test",
		Relation:   "member",
		Subject: &SubjectSpec{
			ObjectType: "user",
			ObjectId:   "user-A",
		},
		Policy: policy,
	}
	expectedMap := map[string]interface{}{
		"objectType": "permission",
		"objectId":   "test",
		"relation":   "member",
		"subject":    spec.Subject.ToMap(),
		"policy":     policy,
	}
	actualMap := spec.ToMap()
	if !cmp.Equal(actualMap, expectedMap) {
		t.Fatalf("Expected map to be %v, but it was %v", expectedMap, actualMap)
	}
}

func TestToMapIndirectWarrantSpec(t *testing.T) {
	spec := WarrantSpec{
		ObjectType: "permission",
		ObjectId:   "test",
		Relation:   "member",
		Subject: &SubjectSpec{
			ObjectType: "role",
			ObjectId:   "admin",
			Relation:   "member",
		},
	}
	expectedMap := map[string]interface{}{
		"objectType": "permission",
		"objectId":   "test",
		"relation":   "member",
		"subject":    spec.Subject.ToMap(),
	}
	actualMap := spec.ToMap()
	if !cmp.Equal(actualMap, expectedMap) {
		t.Fatalf("Expected map to be %v, but it was %v", expectedMap, actualMap)
	}
}

func TestToMapIndirectWarrantSpecWithPolicy(t *testing.T) {
	policy := Policy("tenant == 101")
	spec := WarrantSpec{
		ObjectType: "permission",
		ObjectId:   "test",
		Relation:   "member",
		Subject: &SubjectSpec{
			ObjectType: "role",
			ObjectId:   "admin",
			Relation:   "member",
		},
		Policy: policy,
	}
	expectedMap := map[string]interface{}{
		"objectType": "permission",
		"objectId":   "test",
		"relation":   "member",
		"subject":    spec.Subject.ToMap(),
		"policy":     policy,
	}
	actualMap := spec.ToMap()
	if !cmp.Equal(actualMap, expectedMap) {
		t.Fatalf("Expected map to be %v, but it was %v", expectedMap, actualMap)
	}
}

func TestToStringDirectWarrantSpec(t *testing.T) {
	spec := WarrantSpec{
		ObjectType: "permission",
		ObjectId:   "test",
		Relation:   "member",
		Subject: &SubjectSpec{
			ObjectType: "user",
			ObjectId:   "user-A",
		},
	}
	expectedWarrantStr := "permission:test#member@user:user-A"
	actualWarrantStr := spec.String()
	if actualWarrantStr != expectedWarrantStr {
		t.Fatalf("Expected spec string to be %s, but it was %s", expectedWarrantStr, actualWarrantStr)
	}
}

func TestToStringDirectWarrantSpecWithPolicy(t *testing.T) {
	spec := WarrantSpec{
		ObjectType: "permission",
		ObjectId:   "test",
		Relation:   "member",
		Subject: &SubjectSpec{
			ObjectType: "user",
			ObjectId:   "user-A",
		},
		Policy: Policy("tenant == 101"),
	}
	expectedWarrantStr := "permission:test#member@user:user-A[tenant == 101]"
	actualWarrantStr := spec.String()
	if actualWarrantStr != expectedWarrantStr {
		t.Fatalf("Expected spec string to be %s, but it was %s", expectedWarrantStr, actualWarrantStr)
	}
}

func TestToStringIndirectWarrantSpec(t *testing.T) {
	spec := WarrantSpec{
		ObjectType: "permission",
		ObjectId:   "test",
		Relation:   "member",
		Subject: &SubjectSpec{
			ObjectType: "role",
			ObjectId:   "admin",
			Relation:   "member",
		},
	}
	expectedWarrantStr := "permission:test#member@role:admin#member"
	actualWarrantStr := spec.String()
	if actualWarrantStr != expectedWarrantStr {
		t.Fatalf("Expected spec string to be %s, but it was %s", expectedWarrantStr, actualWarrantStr)
	}
}

func TestToStringIndirectWarrantSpecWithPolicy(t *testing.T) {
	spec := WarrantSpec{
		ObjectType: "permission",
		ObjectId:   "test",
		Relation:   "member",
		Subject: &SubjectSpec{
			ObjectType: "role",
			ObjectId:   "admin",
			Relation:   "member",
		},
		Policy: "tenant == \"101\"",
	}
	expectedWarrantStr := `permission:test#member@role:admin#member[tenant == "101"]`
	actualWarrantStr := spec.String()
	if actualWarrantStr != expectedWarrantStr {
		t.Fatalf("Expected spec string to be %s, but it was %s", expectedWarrantStr, actualWarrantStr)
	}
}

func TestStringToWarrantSpecDirectWarrantSpec(t *testing.T) {
	warrantStr := "permission:test#member@user:user-A"
	expectedWarrantSpec := &WarrantSpec{
		ObjectType: "permission",
		ObjectId:   "test",
		Relation:   "member",
		Subject: &SubjectSpec{
			ObjectType: "user",
			ObjectId:   "user-A",
		},
	}
	actualWarrantSpec, err := StringToWarrantSpec(warrantStr)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(actualWarrantSpec, expectedWarrantSpec) {
		t.Fatalf("Expected warrant spec to be %v, but it was %v", expectedWarrantSpec, actualWarrantSpec)
	}
}

func TestStringToWarrantSpecDirectWarrantSpecWithPolicy(t *testing.T) {
	warrantStr := "permission:test#member@user:user-A[tenant == 101]"
	expectedWarrantSpec := &WarrantSpec{
		ObjectType: "permission",
		ObjectId:   "test",
		Relation:   "member",
		Subject: &SubjectSpec{
			ObjectType: "user",
			ObjectId:   "user-A",
		},
		Policy: Policy("tenant == 101"),
	}
	actualWarrantSpec, err := StringToWarrantSpec(warrantStr)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(actualWarrantSpec, expectedWarrantSpec) {
		t.Fatalf("Expected warrant spec to be %v, but it was %v", expectedWarrantSpec, actualWarrantSpec)
	}
}

func TestStringToWarrantSpecIndirectWarrantSpec(t *testing.T) {
	warrantStr := "permission:test#member@role:admin#member"
	expectedWarrantSpec := &WarrantSpec{
		ObjectType: "permission",
		ObjectId:   "test",
		Relation:   "member",
		Subject: &SubjectSpec{
			ObjectType: "role",
			ObjectId:   "admin",
			Relation:   "member",
		},
	}
	actualWarrantSpec, err := StringToWarrantSpec(warrantStr)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(actualWarrantSpec, expectedWarrantSpec) {
		t.Fatalf("Expected warrant spec to be %v, but it was %v", expectedWarrantSpec, actualWarrantSpec)
	}
}

func TestStringToWarrantSpecIndirectWarrantSpecWithPolicy(t *testing.T) {
	warrantStr := "permission:test#member@role:admin#member[tenant == 101]"
	expectedWarrantSpec := &WarrantSpec{
		ObjectType: "permission",
		ObjectId:   "test",
		Relation:   "member",
		Subject: &SubjectSpec{
			ObjectType: "role",
			ObjectId:   "admin",
			Relation:   "member",
		},
		Policy: Policy("tenant == 101"),
	}
	actualWarrantSpec, err := StringToWarrantSpec(warrantStr)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(actualWarrantSpec, expectedWarrantSpec) {
		t.Fatalf("Expected warrant spec to be %v, but it was %v", expectedWarrantSpec, actualWarrantSpec)
	}
}

func TestStringToWarrantSpecInvalidObject(t *testing.T) {
	warrantStr := "permissiontest#member@user:user:-A"
	expectedErrStr := fmt.Sprintf("invalid object in warrant string %s", warrantStr)
	_, err := StringToWarrantSpec(warrantStr)
	if err == nil || err.Error() != expectedErrStr {
		t.Fatalf("Expected err to be %s, but it was %v", expectedErrStr, err)
	}
}

func TestStringToWarrantSpecInvalidSubject(t *testing.T) {
	warrantStr := "permission:test#member@user:user-A#member#"
	expectedErrStr := fmt.Sprintf("invalid subject in warrant string %s", warrantStr)
	_, err := StringToWarrantSpec(warrantStr)
	if err == nil || !strings.Contains(err.Error(), expectedErrStr) {
		t.Fatalf("Expected err to contain %s, but it was %v", expectedErrStr, err)
	}
}

func TestStringToWarrantSpecInvalidPolicy(t *testing.T) {
	warrantStr := "permission:test#member@user:user-A#member[tenant == 101"
	expectedErrStr := fmt.Sprintf("invalid policy in warrant string %s", warrantStr)
	_, err := StringToWarrantSpec(warrantStr)
	if err == nil || err.Error() != expectedErrStr {
		t.Fatalf("Expected err to be %s, but it was %v", expectedErrStr, err)
	}
}
