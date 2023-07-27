package cloudproviderconfig

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	arov1alpha1 "github.com/Azure/ARO-RP/pkg/operator/apis/aro.openshift.io/v1alpha1"
	"github.com/Azure/ARO-RP/pkg/operator/controllers/base"
	"github.com/Azure/ARO-RP/pkg/util/jsonutils"
)

const (
	ControllerName    = "AROCloudProviderConfig"
	controllerEnabled = "aro.cloudproviderconfig.enabled"
)

var cloudProviderConfigName = types.NamespacedName{Name: "cloud-provider-config", Namespace: "openshift-config"}

// CloudProviderConfigReconciler reconciles the openshift-config/cloud-provider-config ConfigMap
type CloudProviderConfigReconciler struct {
	base.AROController
}

func NewReconciler(Log *logrus.Entry, client client.Client) *CloudProviderConfigReconciler {
	return &CloudProviderConfigReconciler{
		AROController: base.AROController{
			Log:    Log,
			Client: client,
			Name:   ControllerName,
		},
	}
}

// Reconcile makes sure that the cloud-provider-config is healthy
func (r *CloudProviderConfigReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	instance := &arov1alpha1.Cluster{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: arov1alpha1.SingletonClusterName}, instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	if !instance.Spec.OperatorFlags.GetSimpleBoolean(controllerEnabled) {
		r.Log.Debug("controller is disabled")
		return reconcile.Result{}, nil
	}

	r.Log.Debug("running")
	return reconcile.Result{}, r.updateCloudProviderConfig(ctx)
}

// SetupWithManager setup our manager
func (r *CloudProviderConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	cloudProviderConfigPredicate := predicate.NewPredicateFuncs(func(o client.Object) bool {
		return o.GetName() == cloudProviderConfigName.Name && o.GetNamespace() == cloudProviderConfigName.Namespace
	})

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}, builder.WithPredicates(cloudProviderConfigPredicate)).
		Named(ControllerName).
		Complete(r)
}

func (r *CloudProviderConfigReconciler) updateCloudProviderConfig(ctx context.Context) error {
	cm := &corev1.ConfigMap{}
	err := r.Client.Get(ctx, cloudProviderConfigName, cm)
	if err != nil {
		if kerrors.IsNotFound(err) {
			r.Log.Debug("the ConfigMap cloud-provider-config was not found in the openshift-config namespace")
		}
		return err
	}

	jsonConfig, ok := cm.Data["config"]
	if !ok {
		return fmt.Errorf("field config in ConfigMap openshift-config/cloud-provider-config is missing")
	}

	updateMap := map[string]string{"disableOutboundSNAT": "true"}
	changed := false
	cm.Data["config"], changed, err = jsonutils.UpdateJsonString(jsonConfig, updateMap)
	if err != nil {
		return err
	}

	if changed {
		r.Log.Info("openshift-config/cloud-provider-config was updated")
		return r.Client.Update(ctx, cm)
	}

	return nil
}
