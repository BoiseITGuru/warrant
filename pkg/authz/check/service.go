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
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	wookie "github.com/warrant-dev/warrant/pkg/authz/wookie"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/service"
)

type CheckService struct {
	service.BaseService
	WarrantRepository warrant.WarrantRepository
	EventSvc          event.Service
	ObjectTypeSvc     *objecttype.ObjectTypeService
	WookieSvc         *wookie.WookieService
}

func NewService(env service.Env, warrantRepo warrant.WarrantRepository, eventSvc event.Service, objectTypeSvc *objecttype.ObjectTypeService, wookieSvc *wookie.WookieService) *CheckService {
	return &CheckService{
		BaseService:       service.NewBaseService(env),
		WarrantRepository: warrantRepo,
		EventSvc:          eventSvc,
		ObjectTypeSvc:     objectTypeSvc,
		WookieSvc:         wookieSvc,
	}
}

func (svc CheckService) getWithPolicyMatch(ctx context.Context, spec CheckWarrantSpec) (*warrant.WarrantSpec, error) {
	warrants, err := svc.WarrantRepository.GetAllMatchingObjectRelationAndSubject(ctx, spec.ObjectType, spec.ObjectId, spec.Relation, spec.Subject.ObjectType, spec.Subject.ObjectId, spec.Subject.Relation)
	if err != nil || len(warrants) == 0 {
		return nil, err
	}

	// if a warrant without a policy is found, match it
	for _, warrant := range warrants {
		if warrant.GetPolicy() == "" {
			return warrant.ToWarrantSpec(), nil
		}
	}

	for _, warrant := range warrants {
		if warrant.GetPolicy() != "" {
			if policyMatched := evalWarrantPolicy(warrant, spec.Context); policyMatched {
				return warrant.ToWarrantSpec(), nil
			}
		}
	}

	return nil, nil
}

func (svc CheckService) getMatchingSubjects(ctx context.Context, objectType string, objectId string, relation string, checkCtx warrant.PolicyContext) ([]warrant.WarrantSpec, error) {
	log.Ctx(ctx).Debug().Msgf("Getting matching subjects for %s:%s#%s@___%s", objectType, objectId, relation, checkCtx)

	warrantSpecs := make([]warrant.WarrantSpec, 0)
	objectTypeSpec, _, err := svc.ObjectTypeSvc.GetByTypeId(ctx, objectType)
	if err != nil {
		return warrantSpecs, err
	}

	if _, ok := objectTypeSpec.Relations[relation]; !ok {
		return warrantSpecs, nil
	}

	warrants, err := svc.WarrantRepository.GetAllMatchingObjectAndRelation(
		ctx,
		objectType,
		objectId,
		relation,
	)
	if err != nil {
		return warrantSpecs, err
	}

	for _, warrant := range warrants {
		if warrant.GetPolicy() == "" {
			warrantSpecs = append(warrantSpecs, *warrant.ToWarrantSpec())
		} else {
			if policyMatched := evalWarrantPolicy(warrant, checkCtx); policyMatched {
				warrantSpecs = append(warrantSpecs, *warrant.ToWarrantSpec())
			}
		}
	}

	if err != nil {
		return warrantSpecs, err
	}

	return warrantSpecs, nil
}

func (svc CheckService) getMatchingSubjectsBySubjectType(ctx context.Context, objectType string, objectId string, relation string, subjectType string, checkCtx warrant.PolicyContext) ([]warrant.WarrantSpec, error) {
	log.Ctx(ctx).Debug().Msgf("Getting matching subjects for %s:%s#%s@%s:___%s", objectType, objectId, relation, subjectType, checkCtx)

	warrantSpecs := make([]warrant.WarrantSpec, 0)
	objectTypeSpec, _, err := svc.ObjectTypeSvc.GetByTypeId(ctx, objectType)
	if err != nil {
		return warrantSpecs, err
	}

	if _, ok := objectTypeSpec.Relations[relation]; !ok {
		return warrantSpecs, nil
	}

	warrants, err := svc.WarrantRepository.GetAllMatchingObjectAndRelationBySubjectType(
		ctx,
		objectType,
		objectId,
		relation,
		subjectType,
	)
	if err != nil {
		return warrantSpecs, err
	}

	for _, warrant := range warrants {
		if warrant.GetPolicy() == "" {
			warrantSpecs = append(warrantSpecs, *warrant.ToWarrantSpec())
		} else {
			if policyMatched := evalWarrantPolicy(warrant, checkCtx); policyMatched {
				warrantSpecs = append(warrantSpecs, *warrant.ToWarrantSpec())
			}
		}
	}

	if err != nil {
		return warrantSpecs, err
	}

	return warrantSpecs, nil
}

func (svc CheckService) checkRule(ctx context.Context, authInfo *service.AuthInfo, warrantCheck CheckSpec, rule *objecttype.RelationRule) (match bool, decisionPath []warrant.WarrantSpec, err error) {
	warrantSpec := warrantCheck.CheckWarrantSpec
	if rule == nil {
		return false, decisionPath, nil
	}

	switch rule.InheritIf {
	case "":
		// No match found
		return false, decisionPath, nil
	case objecttype.InheritIfAllOf:
		for _, r := range rule.Rules {
			isMatch, matchedPath, err := svc.checkRule(ctx, authInfo, warrantCheck, &r)
			if err != nil {
				return false, decisionPath, err
			}

			decisionPath = append(decisionPath, matchedPath...)
			if !isMatch {
				return false, decisionPath, nil
			}
		}

		return true, decisionPath, nil
	case objecttype.InheritIfAnyOf:
		for _, r := range rule.Rules {
			isMatch, matchedPath, err := svc.checkRule(ctx, authInfo, warrantCheck, &r)
			if err != nil {
				return false, decisionPath, err
			}

			decisionPath = append(decisionPath, matchedPath...)
			if isMatch {
				return true, decisionPath, nil
			}
		}

		return false, decisionPath, nil
	case objecttype.InheritIfNoneOf:
		for _, r := range rule.Rules {
			isMatch, matchedPath, err := svc.checkRule(ctx, authInfo, warrantCheck, &r)
			if err != nil {
				return false, decisionPath, err
			}

			decisionPath = append(decisionPath, matchedPath...)
			if isMatch {
				return false, decisionPath, nil
			}
		}

		return true, decisionPath, nil
	default:
		if rule.OfType == "" && rule.WithRelation == "" {
			match, decisionPath, _, err := svc.Check(ctx, authInfo, CheckSpec{
				CheckWarrantSpec: CheckWarrantSpec{
					ObjectType: warrantSpec.ObjectType,
					ObjectId:   warrantSpec.ObjectId,
					Relation:   rule.InheritIf,
					Subject:    warrantSpec.Subject,
					Context:    warrantSpec.Context,
				},
				Debug: warrantCheck.Debug,
			})
			return match, decisionPath, err
		}

		matchingWarrants, err := svc.getMatchingSubjectsBySubjectType(ctx, warrantSpec.ObjectType, warrantSpec.ObjectId, rule.WithRelation, rule.OfType, warrantSpec.Context)
		if err != nil {
			return false, decisionPath, err
		}

		for _, matchingWarrant := range matchingWarrants {
			match, decisionPath, _, err := svc.Check(ctx, authInfo, CheckSpec{
				CheckWarrantSpec: CheckWarrantSpec{
					ObjectType: matchingWarrant.Subject.ObjectType,
					ObjectId:   matchingWarrant.Subject.ObjectId,
					Relation:   rule.InheritIf,
					Subject:    warrantSpec.Subject,
					Context:    warrantSpec.Context,
				},
				Debug: warrantCheck.Debug,
			})
			if err != nil {
				return false, decisionPath, err
			}

			if match {
				decisionPath = append(decisionPath, matchingWarrant)
				return true, decisionPath, nil
			}
		}

		return false, decisionPath, nil
	}
}

func (svc CheckService) CheckMany(ctx context.Context, authInfo *service.AuthInfo, warrantCheck *CheckManySpec) (*CheckResultSpec, *wookie.Token, error) {
	start := time.Now().UTC()
	if warrantCheck.Op != "" && warrantCheck.Op != objecttype.InheritIfAllOf && warrantCheck.Op != objecttype.InheritIfAnyOf {
		return nil, nil, service.NewInvalidParameterError("op", "must be either anyOf or allOf")
	}

	var checkResult CheckResultSpec
	checkResult.DecisionPath = make(map[string][]warrant.WarrantSpec, 0)

	newWookie, e := svc.WookieSvc.WookieSafeRead(ctx, func(wkCtx context.Context) error {
		if warrantCheck.Op == objecttype.InheritIfAllOf {
			var processingTime int64
			for _, warrantSpec := range warrantCheck.Warrants {
				match, decisionPath, _, err := svc.Check(wkCtx, authInfo, CheckSpec{
					CheckWarrantSpec: warrantSpec,
					Debug:            warrantCheck.Debug,
				})
				if err != nil {
					return err
				}

				if warrantCheck.Debug {
					checkResult.ProcessingTime = processingTime + time.Since(start).Milliseconds()
					if len(decisionPath) > 0 {
						checkResult.DecisionPath[warrantSpec.String()] = decisionPath
					}
				}

				var eventMeta map[string]interface{}
				if warrantSpec.Context != nil {
					eventMeta = make(map[string]interface{})
					eventMeta["context"] = warrantSpec.Context
				}

				if !match {
					err = svc.EventSvc.TrackAccessDeniedEvent(wkCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
					if err != nil {
						return err
					}

					checkResult.Code = http.StatusForbidden
					checkResult.Result = NotAuthorized
					return nil
				}

				err = svc.EventSvc.TrackAccessAllowedEvent(wkCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
				if err != nil {
					return err
				}
			}

			checkResult.Code = http.StatusOK
			checkResult.Result = Authorized
			return nil
		}

		if warrantCheck.Op == objecttype.InheritIfAnyOf {
			var processingTime int64
			for _, warrantSpec := range warrantCheck.Warrants {
				match, decisionPath, _, err := svc.Check(wkCtx, authInfo, CheckSpec{
					CheckWarrantSpec: warrantSpec,
					Debug:            warrantCheck.Debug,
				})
				if err != nil {
					return err
				}

				if warrantCheck.Debug {
					checkResult.ProcessingTime = processingTime + time.Since(start).Milliseconds()
					if len(decisionPath) > 0 {
						checkResult.DecisionPath[warrantSpec.String()] = decisionPath
					}
				}

				var eventMeta map[string]interface{}
				if warrantSpec.Context != nil {
					eventMeta = make(map[string]interface{})
					eventMeta["context"] = warrantSpec.Context
				}

				if match {
					err = svc.EventSvc.TrackAccessAllowedEvent(wkCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
					if err != nil {
						return err
					}

					checkResult.Code = http.StatusOK
					checkResult.Result = Authorized
					return nil
				}

				if !match {
					err := svc.EventSvc.TrackAccessDeniedEvent(wkCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
					if err != nil {
						return err
					}
				}
			}

			checkResult.Code = http.StatusForbidden
			checkResult.Result = NotAuthorized
			return nil
		}

		if len(warrantCheck.Warrants) > 1 {
			return service.NewInvalidParameterError("warrants", "must include operator when including multiple warrants")
		}

		warrantSpec := warrantCheck.Warrants[0]
		match, decisionPath, _, err := svc.Check(wkCtx, authInfo, CheckSpec{
			CheckWarrantSpec: warrantSpec,
			Debug:            warrantCheck.Debug,
		})
		if err != nil {
			return err
		}

		if warrantCheck.Debug {
			checkResult.ProcessingTime = time.Since(start).Milliseconds()
			if len(decisionPath) > 0 {
				checkResult.DecisionPath[warrantSpec.String()] = decisionPath
			}
		}

		var eventMeta map[string]interface{}
		if warrantSpec.Context != nil {
			eventMeta = make(map[string]interface{})
			eventMeta["context"] = warrantSpec.Context
		}

		if match {
			err = svc.EventSvc.TrackAccessAllowedEvent(wkCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
			if err != nil {
				return err
			}

			checkResult.Code = http.StatusOK
			checkResult.Result = Authorized
			return nil
		}

		err = svc.EventSvc.TrackAccessDeniedEvent(wkCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
		if err != nil {
			return err
		}

		checkResult.Code = http.StatusForbidden
		checkResult.Result = NotAuthorized
		return nil
	})
	if e != nil {
		return nil, nil, e
	}
	return &checkResult, newWookie, nil
}

// Check returns true if the subject has a warrant (explicitly or implicitly) for given objectType:objectId#relation and context
func (svc CheckService) Check(ctx context.Context, authInfo *service.AuthInfo, warrantCheck CheckSpec) (bool, []warrant.WarrantSpec, *wookie.Token, error) {
	log.Ctx(ctx).Debug().Msgf("Checking for warrant %s", warrantCheck.String())

	// Used to automatically append tenant context for session token w/ tenantId checks
	if authInfo != nil && authInfo.TenantId != "" {
		svc.appendTenantContext(&warrantCheck, authInfo.TenantId)
	}

	var match bool
	decisionPath := make([]warrant.WarrantSpec, 0)
	var newWookie *wookie.Token
	newWookie, e := svc.WookieSvc.WookieSafeRead(ctx, func(wkCtx context.Context) error {
		// Check for direct warrant match -> doc:readme#viewer@[10]
		matchedWarrant, err := svc.getWithPolicyMatch(ctx, warrantCheck.CheckWarrantSpec)
		if err != nil {
			return err
		}

		if matchedWarrant != nil {
			match = true
			decisionPath = []warrant.WarrantSpec{*matchedWarrant}
			return nil
		}

		// Check against indirectly related warrants
		matchingWarrants, err := svc.getMatchingSubjects(ctx, warrantCheck.ObjectType, warrantCheck.ObjectId, warrantCheck.Relation, warrantCheck.Context)
		if err != nil {
			return err
		}

		for _, matchingWarrant := range matchingWarrants {
			if matchingWarrant.Subject.Relation == "" {
				continue
			}

			match, decisionPath, _, err = svc.Check(ctx, authInfo, CheckSpec{
				CheckWarrantSpec: CheckWarrantSpec{
					ObjectType: matchingWarrant.Subject.ObjectType,
					ObjectId:   matchingWarrant.Subject.ObjectId,
					Relation:   matchingWarrant.Subject.Relation,
					Subject:    warrantCheck.Subject,
					Context:    warrantCheck.Context,
				},
				Debug: warrantCheck.Debug,
			})
			if err != nil {
				return err
			}

			if match {
				decisionPath = append(decisionPath, matchingWarrant)
				return nil
			}
		}

		// Attempt to match against defined rules for target relation
		objectTypeSpec, _, err := svc.ObjectTypeSvc.GetByTypeId(ctx, warrantCheck.ObjectType)
		if err != nil {
			return err
		}

		relationRule := objectTypeSpec.Relations[warrantCheck.Relation]
		match, decisionPath, err = svc.checkRule(ctx, authInfo, warrantCheck, &relationRule)
		if err != nil {
			return err
		}

		if match {
			return nil
		}

		match = false
		return nil
	})
	if e != nil {
		return false, decisionPath, nil, e
	}
	return match, decisionPath, newWookie, nil
}

func (svc CheckService) appendTenantContext(warrantCheck *CheckSpec, tenantId string) {
	if warrantCheck.CheckWarrantSpec.Context == nil {
		warrantCheck.CheckWarrantSpec.Context = warrant.PolicyContext{
			"tenant": tenantId,
		}
	} else {
		warrantCheck.CheckWarrantSpec.Context["tenant"] = tenantId
	}
}

func evalWarrantPolicy(w warrant.Model, policyCtx warrant.PolicyContext) bool {
	policyCtxWithWarrant := make(warrant.PolicyContext)
	for k, v := range policyCtx {
		policyCtxWithWarrant[k] = v
	}
	policyCtxWithWarrant["warrant"] = w

	policyMatched, err := w.GetPolicy().Eval(policyCtxWithWarrant)
	if err != nil {
		log.Err(err).Msgf("Error while evaluating policy %s", w.GetPolicy())
		return false
	}

	return policyMatched
}
