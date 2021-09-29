// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package handlers

import (
	"context"
	"fmt"

	"github.com/Azure/radius/pkg/healthcontract"
	"github.com/Azure/radius/pkg/radlogger"
	"github.com/Azure/radius/pkg/resourcemodel"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

func NewKubernetesServiceHandler(k8s kubernetes.Interface) HealthHandler {
	return &kubernetesServiceHandler{k8s: k8s}
}

type kubernetesServiceHandler struct {
	k8s kubernetes.Interface
}

func (handler *kubernetesServiceHandler) GetHealthState(ctx context.Context, registration HealthRegistration, options Options) HealthState {
	logger := radlogger.GetLogger(ctx)
	kID := registration.Identity.Data.(resourcemodel.KubernetesIdentity)

	var healthState string
	var healthStateErrorDetails string

	// Only checking service existence to mark status as healthy
	_, err := handler.k8s.CoreV1().Services(kID.Namespace).Get(ctx, kID.Name, metav1.GetOptions{})
	if err != nil {
		healthState = healthcontract.HealthStateUnhealthy
		healthStateErrorDetails = err.Error()
	} else {
		healthState = healthcontract.HealthStateHealthy
		healthStateErrorDetails = ""
	}

	// Notify initial health state transition. This needs to be done explicitly since
	// the service might already exist when the health is first probed and the watcher
	// will not detect the initial transition
	msg := HealthState{
		Registration:            registration,
		HealthState:             healthState,
		HealthStateErrorDetails: healthStateErrorDetails,
	}
	options.WatchHealthChangesChannel <- msg
	logger.Info(fmt.Sprintf("Detected health change event for Resource: %+v. Notifying watcher.", registration.Identity))

	// Now watch for changes to the service
	watcher, err := handler.k8s.CoreV1().Services(kID.Namespace).Watch(ctx, metav1.ListOptions{
		Watch:         true,
		LabelSelector: fmt.Sprintf("%s=%s", KubernetesLabelName, kID.Name),
	})
	if err != nil {
		msg := HealthState{
			Registration:            registration,
			HealthState:             healthcontract.HealthStateUnhealthy,
			HealthStateErrorDetails: err.Error(),
		}
		options.WatchHealthChangesChannel <- msg
		return msg
	}
	defer watcher.Stop()

	svcChans := watcher.ResultChan()

	for {
		state := ""
		detail := ""
		select {
		case svcEvent := <-svcChans:
			switch svcEvent.Type {
			case watch.Deleted:
				state = healthcontract.HealthStateUnhealthy
				detail = "Service deleted"
			case watch.Added:
			case watch.Modified:
				state = healthcontract.HealthStateHealthy
				detail = ""
			}
			// Notify the watcher. Let the watcher determine if an action is needed
			msg := HealthState{
				Registration:            registration,
				HealthState:             state,
				HealthStateErrorDetails: detail,
			}
			options.WatchHealthChangesChannel <- msg
			logger.Info(fmt.Sprintf("Detected health change event for Resource: %+v. Notifying watcher.", registration.Identity))
		case <-options.StopChannel:
			logger.Info(fmt.Sprintf("Stopped health monitoring for namespace: %v", kID.Namespace))
			return HealthState{}
		}
	}
}
