package manager

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/thecasualcoder/kube-template/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("should create manager", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := mock.NewMockClient(ctrl)

		mgr := New(client)

		assert.NotNil(t, mgr)
	})
}

func TestManager_Endpoints(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock.NewMockClient(ctrl)
	mgr := New(client)
	expectedEndpoints := v1.Endpoints{
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP:       "10.0.0.100",
						Hostname: "test-host-name",
					},
				},
			},
		},
	}
	namespace := "default"
	resourceName := "nginx"
	client.EXPECT().GetEndpoints(namespace, resourceName).Return(&expectedEndpoints, nil)
	client.EXPECT().WatchEndpoints(namespace, resourceName).Return(safeWatcher(ctrl), nil)

	endpoints, err := mgr.Endpoints(namespace, resourceName)

	assert.NoError(t, err)
	assert.Equal(t, expectedEndpoints, *endpoints)
}

func TestManager_PodsWithLabels(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock.NewMockClient(ctrl)
	mgr := New(client)
	expectedPodList := v1.PodList{
		Items: []v1.Pod{
			{
				Status: v1.PodStatus{PodIP: "10.0.0.1"},
			},
		},
	}
	namespace := "default"
	labelSelector := "app=nginx"
	client.EXPECT().GetPodsWithLabels(namespace, labelSelector).Return(&expectedPodList, nil)
	client.EXPECT().WatchPodsWithLabels(namespace, labelSelector).Return(safeWatcher(ctrl), nil)

	endpoints, err := mgr.PodsWithLabels(namespace, labelSelector)

	assert.NoError(t, err)
	assert.Equal(t, expectedPodList, *endpoints)
}

func safeWatcher(ctrl *gomock.Controller) watch.Interface {
	mockWatch := mock.NewMockInterface(ctrl)
	dummyChannel := make(chan watch.Event)
	mockWatch.EXPECT().ResultChan().Return(dummyChannel).AnyTimes()
	return mockWatch
}
