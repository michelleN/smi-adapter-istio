package trafficsplit

import (
	"context"
	"testing"

	splitv1alpha1 "github.com/deislabs/smi-sdk-go/pkg/apis/split/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	apitypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"

	networkingv1alpha3 "github.com/deislabs/smi-adapter-istio/pkg/apis/networking/v1alpha3"
)

func TestNewReconciler(t *testing.T) {
	mgr := FakeManager{}
	r := newReconciler(mgr)
	var _ reconcile.Reconciler = r // test r is reconcile.Reconciler
}

func TestReconcile_ErrorIsNotFound(t *testing.T) {
	trafficSplitObj := &splitv1alpha1.TrafficSplit{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "traffic-split-name",
			Namespace: "default",
		},
		Spec: splitv1alpha1.TrafficSplitSpec{},
	}
	objs := []runtime.Object{}
	cl := fake.NewFakeClient(objs...)
	s := scheme.Scheme
	s.AddKnownTypes(splitv1alpha1.SchemeGroupVersion, trafficSplitObj)
	reconcileTrafficSplit := &ReconcileTrafficSplit{client: cl, scheme: s}
	req := reconcile.Request{NamespacedName: apitypes.NamespacedName{
		Namespace: "default",
		Name:      "traffic-split-name"},
	}
	_, err := reconcileTrafficSplit.Reconcile(req)
	if err != nil {
		t.Errorf("Expected no err, got %v", err)
	}
}

func TestReconcile(t *testing.T) {
	trafficSplitObj := &splitv1alpha1.TrafficSplit{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "traffic-split-name",
			Namespace: "default",
		},
		Spec: splitv1alpha1.TrafficSplitSpec{},
	}
	virtualServiceObj := &networkingv1alpha3.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "traffic-split-name-vs",
			Namespace: "default",
			Labels:    map[string]string{"traffic-split": "traffic-split-name"},
		},
		Spec: networkingv1alpha3.VirtualServiceSpec{},
	}
	// Register the object in the fake client.
	objs := []runtime.Object{
		trafficSplitObj,
	}
	s := scheme.Scheme
	s.AddKnownTypes(splitv1alpha1.SchemeGroupVersion, trafficSplitObj)
	s.AddKnownTypes(networkingv1alpha3.SchemeGroupVersion, virtualServiceObj)

	cl := fake.NewFakeClient(objs...)
	err := cl.Get(context.TODO(), apitypes.NamespacedName{
		Namespace: "default",
		Name:      "traffic-split-name-vs"},
		virtualServiceObj)
	if !apierrors.IsNotFound(err) {
		t.Fatalf("expected virtual service not to exist, got err: %s", err)
	}

	reconcileTrafficSplit := &ReconcileTrafficSplit{client: cl, scheme: s}
	req := reconcile.Request{NamespacedName: apitypes.NamespacedName{
		Namespace: "default",
		Name:      "traffic-split-name"},
	}

	_, err = reconcileTrafficSplit.Reconcile(req)
	if err != nil {
		t.Errorf("Expected no err, got %s", err)
	}

	err = cl.Get(context.TODO(), apitypes.NamespacedName{
		Namespace: "default",
		Name:      "traffic-split-name-vs"},
		virtualServiceObj)
	if err != nil {
		t.Errorf("Expected virtual service object to be created successfully, but was not: %s", err)
	}
}

func TestNewVSForCR(t *testing) {
	cr := &splitv1alpha1.TrafficSplit{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "simple-ints",
			Namespace: "default",
		},
		Spec: splitv1alpha1.TrafficSplitSpec{},
	}

	vs := newVSForCR(cr)
}

type FakeManager struct{}

func (fm FakeManager) Add(manager.Runnable) error                   { return nil }
func (fm FakeManager) SetFields(interface{}) error                  { return nil }
func (fm FakeManager) Start(<-chan struct{}) error                  { return nil }
func (fm FakeManager) GetConfig() *rest.Config                      { return &rest.Config{} }
func (fm FakeManager) GetScheme() *runtime.Scheme                   { return &runtime.Scheme{} }
func (fm FakeManager) GetAdmissionDecoder() types.Decoder           { return nil }
func (fm FakeManager) GetClient() client.Client                     { return nil }
func (fm FakeManager) GetFieldIndexer() client.FieldIndexer         { return nil }
func (fm FakeManager) GetCache() cache.Cache                        { return nil }
func (fm FakeManager) GetRecorder(name string) record.EventRecorder { return nil }
func (fm FakeManager) GetRESTMapper() meta.RESTMapper               { return nil }
func (fm FakeManager) GetAPIReader() client.Reader                  { return nil }
func (fm FakeManager) GetWebhookServer() *webhook.Server            { return &webhook.Server{} }
