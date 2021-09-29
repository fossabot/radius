// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package deployment

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/radius/pkg/azure/azresources"
	"github.com/Azure/radius/pkg/handlers"
	"github.com/Azure/radius/pkg/healthcontract"
	"github.com/Azure/radius/pkg/model"
	"github.com/Azure/radius/pkg/radlogger"
	"github.com/Azure/radius/pkg/radrp/armerrors"
	"github.com/Azure/radius/pkg/radrp/db"
	"github.com/Azure/radius/pkg/radrp/outputresource"
	"github.com/Azure/radius/pkg/radrp/rest"
	"github.com/Azure/radius/pkg/renderers"
)

//go:generate mockgen -destination=./mock_deploymentprocessor.go -package=deployment -self_package github.com/Azure/radius/pkg/radrp/backend/deployment github.com/Azure/radius/pkg/radrp/backend/deployment DeploymentProcessor

type DeploymentProcessor interface {
	// NOTE: the DeploymentProcessor returns errors but they are just for logging, since it's called
	// asynchronously.

	Deploy(ctx context.Context, operationID azresources.ResourceID, resource db.RadiusResource) error
	Delete(ctx context.Context, id azresources.ResourceID, resource db.RadiusResource) error
}

func NewDeploymentProcessor(appmodel model.ApplicationModelV3, db db.RadrpDB, healthChannels *healthcontract.HealthChannels) DeploymentProcessor {
	return &deploymentProcessor{appmodel: appmodel, db: db, healthChannels: healthChannels}
}

var _ DeploymentProcessor = (*deploymentProcessor)(nil)

type deploymentProcessor struct {
	appmodel       model.ApplicationModelV3
	db             db.RadrpDB
	healthChannels *healthcontract.HealthChannels
}

func (dp *deploymentProcessor) Deploy(ctx context.Context, operationID azresources.ResourceID, resource db.RadiusResource) error {
	logger := radlogger.GetLogger(ctx).WithValues(
		radlogger.LogFieldOperationID, operationID.ID,
	)

	resourceID := operationID.Truncate()

	logger.Info(fmt.Sprintf("Rendering resource: %s, application: %s", resource.ResourceName, resource.ApplicationName))
	renderer, err := dp.appmodel.LookupRenderer(resourceID.Types[len(resourceID.Types)-1].Type) // Using the last type segment as key
	if err != nil {
		armerr := &armerrors.ErrorDetails{
			Code:    armerrors.Invalid,
			Message: err.Error(),
			Target:  resourceID.ID,
		}
		dp.updateOperation(ctx, rest.FailedStatus, operationID, armerr)
		return err
	}

	// Build inputs for renderer
	rendererResource := renderers.RendererResource{
		ApplicationName: resource.ApplicationName,
		ResourceName:    resource.ResourceName,
		ResourceType:    resource.Type,
		Definition:      resource.Definition,
	}

	// Get resources that the resource being deployed has connection with.
	dependencyResourceIDs, err := renderer.GetDependencyIDs(ctx, rendererResource)
	if err != nil {
		armerr := &armerrors.ErrorDetails{
			Code:    armerrors.Invalid,
			Message: err.Error(),
			Target:  resourceID.ID,
		}
		dp.updateOperation(ctx, rest.FailedStatus, operationID, armerr)
		return err
	}

	rendererDependencies, err := dp.fetchDepenendencies(ctx, dependencyResourceIDs)
	if err != nil {
		armerr := &armerrors.ErrorDetails{
			Code:    armerrors.Internal,
			Message: err.Error(),
			Target:  resourceID.ID,
		}
		dp.updateOperation(ctx, rest.FailedStatus, operationID, armerr)
		return err
	}

	// Render - output resources to be deployed for the radius resource
	rendererOutput, err := renderer.Render(ctx, rendererResource, rendererDependencies)
	if err != nil {
		armerr := &armerrors.ErrorDetails{
			Code:    armerrors.Invalid,
			Message: err.Error(),
			Target:  resourceID.ID,
		}
		dp.updateOperation(ctx, rest.FailedStatus, operationID, armerr)
		return err
	}
	// Order output resources in deployment dependency order
	orderedOutputResources, err := outputresource.OrderOutputResources(rendererOutput.Resources)
	if err != nil {
		armerr := &armerrors.ErrorDetails{
			Code:    armerrors.Internal,
			Message: err.Error(),
			Target:  resourceID.ID,
		}
		dp.updateOperation(ctx, rest.FailedStatus, operationID, armerr)
		return err
	}

	// Get existing state of the resource from database, if it's an existing resource
	existingDBResource, err := dp.db.GetV3Resource(ctx, resourceID)
	if err == db.ErrNotFound {
		// no-op - a resource will only exist if this is an update
	} else if err != nil {
		armerr := &armerrors.ErrorDetails{
			Code:    armerrors.Invalid,
			Message: err.Error(),
			Target:  resourceID.ID,
		}
		dp.updateOperation(ctx, rest.FailedStatus, operationID, armerr)
		return err
	}
	existingDBOutputResources := existingDBResource.Status.OutputResources

	// Deploy and update the radius resource
	dbOutputResources := []db.OutputResource{}
	// values consumed by other Radius resource types through connections
	computedValues := map[string]interface{}{}
	// Map of localID to properties deployed for each output resource. Consumed by handler of any output resource with dependencies on other output resources
	// Example - CosmosDBAccountName consumed by CosmosDBMongo/SQL handler
	deployedOutputResourceProperties := map[string]map[string]string{}
	for _, outputResource := range orderedOutputResources {
		logger.Info(fmt.Sprintf("Deploying output resource - LocalID: %s, type: %s\n", outputResource.LocalID, outputResource.ResourceKind))

		var existingOutputResourceState db.OutputResource
		for _, dbOutputResource := range existingDBOutputResources {
			if dbOutputResource.LocalID == outputResource.LocalID {
				existingOutputResourceState = dbOutputResource
				break
			}
		}

		resourceHandlers, err := dp.appmodel.LookupHandlers(outputResource.ResourceKind)
		if err != nil {
			armerr := &armerrors.ErrorDetails{
				Code:    armerrors.Invalid,
				Message: err.Error(),
				Target:  resourceID.ID,
			}
			dp.updateOperation(ctx, rest.FailedStatus, operationID, armerr)
			return err
		}

		properties, err := resourceHandlers.ResourceHandler.Put(ctx, &handlers.PutOptions{
			Application:            resource.ApplicationName,
			Component:              resource.ResourceName,
			Resource:               &outputResource,
			ExistingOutputResource: &existingOutputResourceState,
			DependencyProperties:   deployedOutputResourceProperties,
		})
		if err != nil {
			armerr := &armerrors.ErrorDetails{
				Code:    armerrors.Internal,
				Message: err.Error(),
				Target:  resourceID.ID,
			}
			dp.updateOperation(ctx, rest.FailedStatus, operationID, armerr)
			return err
		}
		deployedOutputResourceProperties[outputResource.LocalID] = properties

		if outputResource.Identity.Kind == "" {
			armerr := &armerrors.ErrorDetails{
				Code:    armerrors.Internal,
				Message: err.Error(),
				Target:  resourceID.ID,
			}
			dp.updateOperation(ctx, rest.FailedStatus, operationID, armerr)
			return fmt.Errorf("output resource %q does not have an identity. This is a bug in the handler. ", outputResource.LocalID)
		}

		// Copy deployed output resource property values into corresponding expected computed values
		for k, v := range rendererOutput.ComputedValues {
			if outputResource.LocalID == v.LocalID {
				computedValues[k] = properties[v.PropertyReference]
			}
		}

		// Register health checks for the output resource
		healthResource := healthcontract.HealthResource{
			Identity:         outputResource.Identity,
			ResourceKind:     outputResource.ResourceKind,
			RadiusResourceID: resource.ID,
		}

		dp.registerOutputResourceForHealthChecks(ctx, healthResource, resourceHandlers.HealthHandler.GetHealthOptions(ctx))

		// Build database resource - copy updated properties to Resource field
		dbOutputResource := db.OutputResource{
			LocalID:             outputResource.LocalID,
			ResourceKind:        outputResource.ResourceKind,
			Identity:            outputResource.Identity,
			Managed:             outputResource.Managed,
			PersistedProperties: properties,
			Status: db.OutputResourceStatus{
				ProvisioningState:        db.Provisioned,
				ProvisioningErrorDetails: "",
			},
		}
		dbOutputResources = append(dbOutputResources, dbOutputResource)
	}

	// Update static values for connections
	for k, computedValue := range rendererOutput.ComputedValues {
		if computedValue.Value != nil {
			computedValues[k] = computedValue.Value
		}
	}

	// Persist updated/created resource in the database
	resourceStatus := db.RadiusResourceStatus{
		ProvisioningState: db.Provisioned,
		OutputResources:   dbOutputResources,
	}
	updatedRadiusResource := db.RadiusResource{
		ID:              resource.ID,
		Type:            resource.Type,
		SubscriptionID:  resource.SubscriptionID,
		ResourceGroup:   resource.ResourceGroup,
		ApplicationName: resource.ApplicationName,
		ResourceName:    resource.ResourceName,

		Definition:     resource.Definition,
		ComputedValues: computedValues,

		Status: resourceStatus,

		ProvisioningState: string(rest.SuccededStatus),
	}

	err = dp.db.UpdateV3ResourceStatus(ctx, resourceID, updatedRadiusResource)
	if err != nil {
		armerr := &armerrors.ErrorDetails{
			Code:    armerrors.Internal,
			Message: err.Error(),
			Target:  resourceID.ID,
		}
		dp.updateOperation(ctx, rest.FailedStatus, operationID, armerr)
		return err
	}

	// Update operation
	dp.updateOperation(ctx, rest.SuccededStatus, operationID, nil /* success */)

	return nil
}

func (dp *deploymentProcessor) Delete(ctx context.Context, operationID azresources.ResourceID, resource db.RadiusResource) error {
	logger := radlogger.GetLogger(ctx).WithValues(
		radlogger.LogFieldOperationID, operationID.ID,
	)

	resourceID := operationID.Truncate()

	// Loop over each output resource and delete in reverse dependency order - resource deployed last should be deleted first
	deployedOutputResources := resource.Status.OutputResources
	for i := len(deployedOutputResources) - 1; i >= 0; i-- {
		outputResource := deployedOutputResources[i]
		resourceHandlers, err := dp.appmodel.LookupHandlers(outputResource.ResourceKind)
		if err != nil {
			armerr := &armerrors.ErrorDetails{
				Code:    armerrors.Invalid,
				Message: err.Error(),
				Target:  resourceID.ID,
			}
			dp.updateOperation(ctx, rest.FailedStatus, operationID, armerr)
			return err
		}

		logger.Info(fmt.Sprintf("Deleting output resource - LocalID: %s, type: %s\n", outputResource.LocalID, outputResource.ResourceKind))
		err = resourceHandlers.ResourceHandler.Delete(ctx, handlers.DeleteOptions{
			Application:            resource.ApplicationName,
			Component:              resource.ResourceGroup,
			ExistingOutputResource: &outputResource,
		})
		if err != nil {
			armerr := &armerrors.ErrorDetails{
				Code:    armerrors.Internal,
				Message: err.Error(),
				Target:  resourceID.ID,
			}
			dp.updateOperation(ctx, rest.FailedStatus, operationID, armerr)
			return err
		}

		healthResource := healthcontract.HealthResource{
			ResourceKind:     outputResource.ResourceKind,
			Identity:         outputResource.Identity,
			RadiusResourceID: resource.ID,
		}
		dp.unregisterOutputResourceForHealthChecks(ctx, healthResource)
	}

	// Delete resource from database
	err := dp.db.DeleteV3Resource(ctx, resourceID)
	if err != nil {
		armerr := &armerrors.ErrorDetails{
			Code:    armerrors.Internal,
			Message: err.Error(),
			Target:  resourceID.ID,
		}
		dp.updateOperation(ctx, rest.FailedStatus, operationID, armerr)
		return err
	}

	// Update operation
	dp.updateOperation(ctx, rest.SuccededStatus, operationID, nil /* success */)

	return nil
}

func (dp *deploymentProcessor) registerOutputResourceForHealthChecks(ctx context.Context, resource healthcontract.HealthResource, healthCheckOptions healthcontract.HealthCheckOptions) {
	logger := radlogger.GetLogger(ctx)

	msg := healthcontract.ResourceHealthRegistrationMessage{
		Action:   healthcontract.ActionRegister,
		Resource: resource,
		Options:  healthCheckOptions,
	}
	dp.healthChannels.ResourceRegistrationWithHealthChannel <- msg

	logger.Info("Registered output resource for health checks", resource.Identity.AsLogValues()...)
}

func (dp *deploymentProcessor) unregisterOutputResourceForHealthChecks(ctx context.Context, resource healthcontract.HealthResource) {
	logger := radlogger.GetLogger(ctx)
	logger.Info("Unregistering resource with the health service...", resource.Identity.AsLogValues()...)
	msg := healthcontract.ResourceHealthRegistrationMessage{
		Action:   healthcontract.ActionUnregister,
		Resource: resource,
	}
	dp.healthChannels.ResourceRegistrationWithHealthChannel <- msg
}

// Retrieves and updates existing database entry for the operation
func (dp *deploymentProcessor) updateOperation(ctx context.Context, status rest.OperationStatus, operationResourceID azresources.ResourceID, armerr *armerrors.ErrorDetails) {
	logger := radlogger.GetLogger(ctx)

	operation, err := dp.db.GetOperationByID(ctx, operationResourceID)
	if err == db.ErrNotFound {
		// Operation entry should have been created in the db before we get here
		logger.Error(err, fmt.Sprintf("Update operation failed - operation with id %s was not found in the database.", operationResourceID.ID))
		return
	} else if err != nil {
		logger.Error(err, "Failed to update the operation in database.")
		return
	}

	operation.EndTime = time.Now().UTC().Format(time.RFC3339)
	operation.PercentComplete = 100
	operation.Status = string(status)
	operation.Error = armerr

	_, err = dp.db.PatchOperationByID(ctx, operationResourceID, operation)
	if err != nil {
		logger.Error(err, "Failed to update the operation in database.")
	}
}

func (dp *deploymentProcessor) fetchDepenendencies(ctx context.Context, dependencyResourceIDs []azresources.ResourceID) (map[string]renderers.RendererDependency, error) {
	rendererDependencies := map[string]renderers.RendererDependency{}
	for _, dependencyResourceID := range dependencyResourceIDs {
		// Fetch resource from db
		dbDependencyResource, err := dp.db.GetV3Resource(ctx, dependencyResourceID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch dependency resource %q: %w", dependencyResourceID.ID, err)
		}

		rendererDependency := renderers.RendererDependency{
			ResourceID:     dependencyResourceID,
			Definition:     dbDependencyResource.Definition,
			ComputedValues: dbDependencyResource.ComputedValues,
		}

		rendererDependencies[dependencyResourceID.ID] = rendererDependency
	}

	return rendererDependencies, nil
}