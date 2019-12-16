package main

import "github.com/thecasualcoder/kube-template/cmd"

//go:generate mockgen -source pkg/manager/manager.go -destination mock/manager.go -package mock

func main() {
	cmd.Execute()
}
