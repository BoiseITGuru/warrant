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

import "time"

type FilterOptions struct {
	ObjectType string
}

type ObjectSpec struct {
	// NOTE: ID is required here for internal use.
	// However, we don't return it to the client.
	ID         int64     `json:"-"`
	ObjectType string    `json:"objectType" validate:"required,valid_object_type"`
	ObjectId   string    `json:"objectId" validate:"required,valid_object_id"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (spec ObjectSpec) ToObject() *Object {
	return &Object{
		ObjectType: spec.ObjectType,
		ObjectId:   spec.ObjectId,
		CreatedAt:  spec.CreatedAt,
	}
}

type CreateObjectSpec struct {
	ObjectType string `json:"objectType" validate:"required"`
	ObjectId   string `json:"objectId" validate:"required"`
}
