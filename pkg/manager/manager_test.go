package manager

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/thecasualcoder/kube-template/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"testing"
	"time"
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
	watcher, resultChan := safeWatcher(ctrl)
	resultChan <- watch.Event{
		Object: &expectedEndpoints,
	}
	client.EXPECT().WatchEndpoints(namespace, resourceName).Return(watcher, nil)

	var actualEndpoints v1.Endpoints

	for i := 1; i <= 3; i++ {
		endpoints, err := mgr.Endpoints(namespace, resourceName)
		if err == ErrDataNotReady {
			time.Sleep(time.Duration(i*100) * time.Millisecond)
			continue
		}

		if assert.NoError(t, err) {
			actualEndpoints = *endpoints
		}
	}

	assert.Equal(t, expectedEndpoints, actualEndpoints)
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
	watcher, eventChan := safeWatcher(ctrl)
	eventChan <- watch.Event{}
	client.EXPECT().WatchPodsWithLabels(namespace, labelSelector).Return(watcher, nil)

	var actualPodList v1.PodList
	for i := 1; i <= 3; i++ {
		podList, err := mgr.PodsWithLabels(namespace, labelSelector)
		if err == ErrDataNotReady {
			time.Sleep(time.Duration(i*100) * time.Millisecond)
			continue
		}

		if assert.NoError(t, err) {
			actualPodList = *podList
		}
	}

	assert.Equal(t, expectedPodList, actualPodList)
}

func safeWatcher(ctrl *gomock.Controller) (watch.Interface, chan watch.Event) {
	mockWatch := mock.NewMockInterface(ctrl)
	dummyChannel := make(chan watch.Event, 1)
	mockWatch.EXPECT().ResultChan().Return(dummyChannel).AnyTimes()
	return mockWatch, dummyChannel
}
