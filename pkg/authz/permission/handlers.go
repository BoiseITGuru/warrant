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
	"net/http"

	"github.com/gorilla/mux"
	wookie "github.com/warrant-dev/warrant/pkg/authz/wookie"
	"github.com/warrant-dev/warrant/pkg/service"
)

// GetRoutes registers all route handlers for this module
func (svc PermissionService) Routes() ([]service.Route, error) {
	return []service.Route{
		// create
		service.WarrantRoute{
			Pattern: "/v1/permissions",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, CreateHandler),
		},

		// get
		service.WarrantRoute{
			Pattern: "/v1/permissions",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, ListHandler),
				service.ListMiddleware[PermissionListParamParser],
			),
		},
		service.WarrantRoute{
			Pattern: "/v1/permissions/{permissionId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, GetHandler),
		},

		// update
		service.WarrantRoute{
			Pattern: "/v1/permissions/{permissionId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},
		service.WarrantRoute{
			Pattern: "/v1/permissions/{permissionId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},

		// delete
		service.WarrantRoute{
			Pattern: "/v1/permissions/{permissionId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, DeleteHandler),
		},
	}, nil
}

func CreateHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
	var newPermission PermissionSpec
	err := service.ParseJSONBody(r.Body, &newPermission)
	if err != nil {
		return err
	}

	createdPermission, err := svc.Create(r.Context(), newPermission)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdPermission)
	return nil
}

func GetHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
	permissionId := mux.Vars(r)["permissionId"]
	permission, err := svc.GetByPermissionId(r.Context(), permissionId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, permission)
	return nil
}

func ListHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
	listParams := service.GetListParamsFromContext[PermissionListParamParser](r.Context())
	permissions, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, permissions)
	return nil
}

func UpdateHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
	var updatePermission UpdatePermissionSpec
	err := service.ParseJSONBody(r.Body, &updatePermission)
	if err != nil {
		return err
	}

	permissionId := mux.Vars(r)["permissionId"]
	updatedPermission, err := svc.UpdateByPermissionId(r.Context(), permissionId, updatePermission)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedPermission)
	return nil
}

func DeleteHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
	permissionId := mux.Vars(r)["permissionId"]
	if permissionId == "" {
		return service.NewMissingRequiredParameterError("permissionId")
	}

	newWookie, err := svc.DeleteByPermissionId(r.Context(), permissionId)
	if err != nil {
		return err
	}
	wookie.AddAsResponseHeader(w, newWookie)

	return nil
}
