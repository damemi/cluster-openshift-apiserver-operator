package resourcesynccontroller

import (
	"encoding/json"
	"net/http"

	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resourcesynccontroller"
	"github.com/openshift/library-go/pkg/operator/v1helpers"

	"github.com/openshift/cluster-openshift-apiserver-operator/pkg/operator/operatorclient"
)

func NewResourceSyncController(
	operatorConfigClient v1helpers.OperatorClient,
	kubeInformersForNamespaces v1helpers.KubeInformersForNamespaces,
	configMapsGetter corev1client.ConfigMapsGetter,
	secretsGetter corev1client.SecretsGetter,
	eventRecorder events.Recorder) (*resourcesynccontroller.ResourceSyncController, error) {

	resourceSyncController := resourcesynccontroller.NewResourceSyncController(
		operatorConfigClient,
		kubeInformersForNamespaces,
		secretsGetter,
		configMapsGetter,
		eventRecorder,
	)
	if err := resourceSyncController.SyncConfigMap(
		resourcesynccontroller.ResourceLocation{Namespace: operatorclient.TargetNamespace, Name: "etcd-serving-ca"},
		resourcesynccontroller.ResourceLocation{Namespace: operatorclient.EtcdNamespaceName, Name: "etcd-serving-ca"},
	); err != nil {
		return nil, err
	}
	if err := resourceSyncController.SyncSecret(
		resourcesynccontroller.ResourceLocation{Namespace: operatorclient.TargetNamespace, Name: "etcd-client"},
		resourcesynccontroller.ResourceLocation{Namespace: operatorclient.EtcdNamespaceName, Name: "etcd-client"},
	); err != nil {
		return nil, err
	}
	if err := resourceSyncController.SyncConfigMap(
		resourcesynccontroller.ResourceLocation{Namespace: operatorclient.TargetNamespace, Name: "client-ca"},
		resourcesynccontroller.ResourceLocation{Namespace: operatorclient.GlobalMachineSpecifiedConfigNamespace, Name: "kube-apiserver-client-ca"},
	); err != nil {
		return nil, err
	}
	if err := resourceSyncController.SyncConfigMap(
		resourcesynccontroller.ResourceLocation{Namespace: operatorclient.TargetNamespace, Name: "aggregator-client-ca"},
		resourcesynccontroller.ResourceLocation{Namespace: operatorclient.GlobalMachineSpecifiedConfigNamespace, Name: "kube-apiserver-aggregator-client-ca"},
	); err != nil {
		return nil, err
	}

	return resourceSyncController, nil
}

func NewDebugHandler(controller *resourcesynccontroller.ResourceSyncController) http.Handler {
	return &debugHTTPHandler{controller: controller}
}

type debugHTTPHandler struct {
	controller *resourcesynccontroller.ResourceSyncController
}

func (h *debugHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/syncRules" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data, err := json.Marshal(h.controller.SecretSyncRules())
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
	w.WriteHeader(http.StatusOK)
}
