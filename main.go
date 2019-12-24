package main

import "github.com/thecasualcoder/kube-template/cmd"

//go:generate mockgen -source pkg/manager/manager.go -destination mock/manager.go -package mock
//go:generate mockgen -source pkg/kubernetes/client.go -destination mock/client.go -package mock
//go:generate mockgen -package mock -destination mock/watch.go k8s.io/apimachinery/pkg/watch Interface

func main() {
	cmd.Execute()
}
