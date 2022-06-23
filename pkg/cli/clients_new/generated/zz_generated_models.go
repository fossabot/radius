//go:build go1.16
// +build go1.16

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

package generated

import (
	"encoding/json"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"reflect"
	"time"
)

// ErrorAdditionalInfo - The resource management error additional info.
type ErrorAdditionalInfo struct {
	// READ-ONLY; The additional info.
	Info map[string]interface{} `json:"info,omitempty" azure:"ro"`

	// READ-ONLY; The additional info type.
	Type *string `json:"type,omitempty" azure:"ro"`
}

// ErrorDetail - The error detail.
type ErrorDetail struct {
	// READ-ONLY; The error additional info.
	AdditionalInfo []*ErrorAdditionalInfo `json:"additionalInfo,omitempty" azure:"ro"`

	// READ-ONLY; The error code.
	Code *string `json:"code,omitempty" azure:"ro"`

	// READ-ONLY; The error details.
	Details []*ErrorDetail `json:"details,omitempty" azure:"ro"`

	// READ-ONLY; The error message.
	Message *string `json:"message,omitempty" azure:"ro"`

	// READ-ONLY; The error target.
	Target *string `json:"target,omitempty" azure:"ro"`
}

// MarshalJSON implements the json.Marshaller interface for type ErrorDetail.
func (e ErrorDetail) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "additionalInfo", e.AdditionalInfo)
	populate(objectMap, "code", e.Code)
	populate(objectMap, "details", e.Details)
	populate(objectMap, "message", e.Message)
	populate(objectMap, "target", e.Target)
	return json.Marshal(objectMap)
}

// ErrorResponse - Common error response for all Azure Resource Manager APIs to return error details for failed operations. (This also follows the OData
// error response format.).
// Implements the error and azcore.HTTPResponse interfaces.
type ErrorResponse struct {
	raw string
	// The error object.
	InnerError *ErrorDetail `json:"error,omitempty"`
}

// Error implements the error interface for type ErrorResponse.
// The contents of the error text are not contractual and subject to change.
func (e ErrorResponse) Error() string {
	return e.raw
}

// GenericResource - Generic resource
type GenericResource struct {
	TrackedResource
	// Any object
	ResourceProperties map[string]interface{} `json:"resourceProperties,omitempty"`

	// READ-ONLY; Metadata pertaining to creation and last modification of the resource.
	SystemData *SystemData `json:"systemData,omitempty" azure:"ro"`
}

// MarshalJSON implements the json.Marshaller interface for type GenericResource.
func (g GenericResource) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	g.TrackedResource.marshalInternal(objectMap)
	populate(objectMap, "resourceProperties", g.ResourceProperties)
	populate(objectMap, "systemData", g.SystemData)
	return json.Marshal(objectMap)
}

type GenericResourceProperties struct {
	PropertyKey *string `json:"propertyKey,omitempty"`

	// Any object
	PropertyValue map[string]interface{} `json:"propertyValue,omitempty"`
}

// GenericResourcesCreateOrUpdateOptions contains the optional parameters for the GenericResources.CreateOrUpdate method.
type GenericResourcesCreateOrUpdateOptions struct {
	// placeholder for future optional parameters
}

// GenericResourcesDeleteOptions contains the optional parameters for the GenericResources.Delete method.
type GenericResourcesDeleteOptions struct {
	// placeholder for future optional parameters
}

// GenericResourcesGetOptions contains the optional parameters for the GenericResources.Get method.
type GenericResourcesGetOptions struct {
	// placeholder for future optional parameters
}

// GenericResourcesList - Object that includes an array of GenericResources and a possible link for next set
type GenericResourcesList struct {
	// The link used to fetch the next page of resource list.
	NextLink *string `json:"nextLink,omitempty"`

	// List of GenericResources resources
	Value []*GenericResource `json:"value,omitempty"`
}

// MarshalJSON implements the json.Marshaller interface for type GenericResourcesList.
func (g GenericResourcesList) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "nextLink", g.NextLink)
	populate(objectMap, "value", g.Value)
	return json.Marshal(objectMap)
}

// GenericResourcesListByRootScopeOptions contains the optional parameters for the GenericResources.ListByRootScope method.
type GenericResourcesListByRootScopeOptions struct {
	// placeholder for future optional parameters
}

// Resource - Common fields that are returned in the response for all Azure Resource Manager resources
type Resource struct {
	// READ-ONLY; Fully qualified resource ID for the resource. Ex - /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/{resourceProviderNamespace}/{resourceType}/{resourceName}
	ID *string `json:"id,omitempty" azure:"ro"`

	// READ-ONLY; The name of the resource
	Name *string `json:"name,omitempty" azure:"ro"`

	// READ-ONLY; The type of the resource. E.g. "Microsoft.Compute/virtualMachines" or "Microsoft.Storage/storageAccounts"
	Type *string `json:"type,omitempty" azure:"ro"`
}

// MarshalJSON implements the json.Marshaller interface for type Resource.
func (r Resource) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	r.marshalInternal(objectMap)
	return json.Marshal(objectMap)
}

func (r Resource) marshalInternal(objectMap map[string]interface{}) {
	populate(objectMap, "id", r.ID)
	populate(objectMap, "name", r.Name)
	populate(objectMap, "type", r.Type)
}

// SystemData - Metadata pertaining to creation and last modification of the resource.
type SystemData struct {
	// The timestamp of resource creation (UTC).
	CreatedAt *time.Time `json:"createdAt,omitempty"`

	// The identity that created the resource.
	CreatedBy *string `json:"createdBy,omitempty"`

	// The type of identity that created the resource.
	CreatedByType *CreatedByType `json:"createdByType,omitempty"`

	// The timestamp of resource last modification (UTC)
	LastModifiedAt *time.Time `json:"lastModifiedAt,omitempty"`

	// The identity that last modified the resource.
	LastModifiedBy *string `json:"lastModifiedBy,omitempty"`

	// The type of identity that last modified the resource.
	LastModifiedByType *CreatedByType `json:"lastModifiedByType,omitempty"`
}

// MarshalJSON implements the json.Marshaller interface for type SystemData.
func (s SystemData) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "createdAt", (*timeRFC3339)(s.CreatedAt))
	populate(objectMap, "createdBy", s.CreatedBy)
	populate(objectMap, "createdByType", s.CreatedByType)
	populate(objectMap, "lastModifiedAt", (*timeRFC3339)(s.LastModifiedAt))
	populate(objectMap, "lastModifiedBy", s.LastModifiedBy)
	populate(objectMap, "lastModifiedByType", s.LastModifiedByType)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type SystemData.
func (s *SystemData) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return err
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "createdAt":
				var aux timeRFC3339
				err = unpopulate(val, &aux)
				s.CreatedAt = (*time.Time)(&aux)
				delete(rawMsg, key)
		case "createdBy":
				err = unpopulate(val, &s.CreatedBy)
				delete(rawMsg, key)
		case "createdByType":
				err = unpopulate(val, &s.CreatedByType)
				delete(rawMsg, key)
		case "lastModifiedAt":
				var aux timeRFC3339
				err = unpopulate(val, &aux)
				s.LastModifiedAt = (*time.Time)(&aux)
				delete(rawMsg, key)
		case "lastModifiedBy":
				err = unpopulate(val, &s.LastModifiedBy)
				delete(rawMsg, key)
		case "lastModifiedByType":
				err = unpopulate(val, &s.LastModifiedByType)
				delete(rawMsg, key)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// TrackedResource - The resource model definition for an Azure Resource Manager tracked top level resource which has 'tags' and a 'location'
type TrackedResource struct {
	Resource
	// REQUIRED; The geo-location where the resource lives
	Location *string `json:"location,omitempty"`

	// Resource tags.
	Tags map[string]*string `json:"tags,omitempty"`
}

// MarshalJSON implements the json.Marshaller interface for type TrackedResource.
func (t TrackedResource) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	t.marshalInternal(objectMap)
	return json.Marshal(objectMap)
}

func (t TrackedResource) marshalInternal(objectMap map[string]interface{}) {
	t.Resource.marshalInternal(objectMap)
	populate(objectMap, "location", t.Location)
	populate(objectMap, "tags", t.Tags)
}

func populate(m map[string]interface{}, k string, v interface{}) {
	if v == nil {
		return
	} else if azcore.IsNullValue(v) {
		m[k] = nil
	} else if !reflect.ValueOf(v).IsNil() {
		m[k] = v
	}
}

func unpopulate(data json.RawMessage, v interface{}) error {
	if data == nil {
		return nil
	}
	return json.Unmarshal(data, v)
}
