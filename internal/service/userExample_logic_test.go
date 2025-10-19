package service

import (
	"context"
	"testing"
	"time"

	"github.com/moweilong/milady/pkg/utils"

	"github.com/moweilong/milady/configs"
	"github.com/moweilong/milady/internal/config"
	"github.com/moweilong/milady/internal/rpcclient"
)

func TestNewUserExampleServiceClient(t *testing.T) {
	err := config.Init(configs.Path("serverNameExample.yml"))
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		defer func() { recover() }()
		rpcclient.NewServerNameExampleRPCConn()
	}()

	time.Sleep(time.Millisecond * 200)
	cli := NewUserExampleClient()
	ctx := context.Background()

	utils.SafeRunWithTimeout(time.Second, func(cancel context.CancelFunc) {
		reply, err := cli.Create(ctx, nil)
		t.Log(reply, err)
		cancel()
	})
	utils.SafeRunWithTimeout(time.Second, func(cancel context.CancelFunc) {
		reply, err := cli.DeleteByID(ctx, nil)
		t.Log(reply, err)
		cancel()
	})
	utils.SafeRunWithTimeout(time.Second, func(cancel context.CancelFunc) {
		reply, err := cli.UpdateByID(ctx, nil)
		t.Log(reply, err)
		cancel()
	})
	utils.SafeRunWithTimeout(time.Second, func(cancel context.CancelFunc) {
		reply, err := cli.GetByID(ctx, nil)
		t.Log(reply, err)
		cancel()
	})
	utils.SafeRunWithTimeout(time.Second, func(cancel context.CancelFunc) {
		reply, err := cli.List(ctx, nil)
		t.Log(reply, err)
		cancel()
	})

}
