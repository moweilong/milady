package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/moweilong/milady/pkg/grpc/benchmark"

	serverNameExampleV1 "github.com/moweilong/milady/api/serverNameExample/v1"
	"github.com/moweilong/milady/api/types"
	"github.com/moweilong/milady/configs"
	"github.com/moweilong/milady/internal/config"
)

// Test each method of userExample via the rpc client
func Test_service_userExample_methods(t *testing.T) {
	conn := getRPCClientConnForTest()
	cli := serverNameExampleV1.NewUserExampleClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
	//ctx = interceptor.SetJwtTokenToCtx(ctx, token)

	tests := []struct {
		name    string
		fn      func() (interface{}, error)
		wantErr bool
	}{
		// todo generate the service struct code here
		// delete the templates code start
		{
			name: "Create",
			fn: func() (interface{}, error) {
				// todo type in the parameters before testing
				req := &serverNameExampleV1.CreateUserExampleRequest{
					Name:     "foo7",
					Email:    "foo7@bar.com",
					Password: "f447b20a7fcbf53a5d5be013ea0b15af",
					Phone:    "16000000000",
					Avatar:   "http://internal.com/7.jpg",
					Age:      11,
					Gender:   2,
				}
				return cli.Create(ctx, req)
			},
			wantErr: false,
		},

		{
			name: "UpdateByID",
			fn: func() (interface{}, error) {
				// todo type in the parameters before testing
				req := &serverNameExampleV1.UpdateUserExampleByIDRequest{
					Id:    7,
					Phone: "16000000001",
					Age:   11,
				}
				return cli.UpdateByID(ctx, req)
			},
			wantErr: false,
		},
		// delete the templates code end
		{
			name: "DeleteByID",
			fn: func() (interface{}, error) {
				// todo type in the parameters before testing
				req := &serverNameExampleV1.DeleteUserExampleByIDRequest{
					Id: 100,
				}
				return cli.DeleteByID(ctx, req)
			},
			wantErr: false,
		},

		{
			name: "GetByID",
			fn: func() (interface{}, error) {
				// todo type in the parameters before testing
				req := &serverNameExampleV1.GetUserExampleByIDRequest{
					Id: 1,
				}
				return cli.GetByID(ctx, req)
			},
			wantErr: false,
		},

		{
			name: "List",
			fn: func() (interface{}, error) {
				// todo type in the parameters before testing
				req := &serverNameExampleV1.ListUserExampleRequest{
					Params: &types.Params{
						Page:  0,
						Limit: 10,
						Sort:  "",
						Columns: []*types.Column{
							{
								Name:  "id",
								Exp:   ">=",
								Value: "1",
								Logic: "",
							},
						},
					},
				}
				return cli.List(ctx, req)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fn()
			if (err != nil) != tt.wantErr {
				t.Logf("test '%s' error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			data, _ := json.MarshalIndent(got, "", "    ")
			fmt.Println(string(data))
		})
	}
}

// Perform a stress test on {{.LowerName}}'s method and
// copy the press test report to your browser when you are finished.
func Test_service_userExample_benchmark(t *testing.T) {
	err := config.Init(configs.Path("serverNameExample.yml"))
	if err != nil {
		panic(err)
	}

	grpcClientCfg := getGRPCClientCfg()
	host := fmt.Sprintf("%s:%d", grpcClientCfg.Host, grpcClientCfg.Port)
	protoFile := configs.Path("../api/serverNameExample/v1/userExample.proto")
	// If third-party dependencies are missing during the press test,
	// copy them to the project's third_party directory.
	dependentProtoFilePath := []string{
		configs.Path("../third_party"), // third_party directory
		configs.Path(".."),             // Previous level of third_party
	}

	tests := []struct {
		name    string
		fn      func() error
		wantErr bool
	}{
		{
			name: "GetByID",
			fn: func() error {
				// todo type in the parameters before testing
				message := &serverNameExampleV1.GetUserExampleByIDRequest{
					Id: 1,
				}
				total := 10 // total number of requests

				b, err := benchmark.New(host, protoFile, "GetByID", message, dependentProtoFilePath, total)
				if err != nil {
					return err
				}
				return b.Run()
			},
			wantErr: false,
		},

		{
			name: "List",
			fn: func() error {
				// todo type in the parameters before testing
				message := &serverNameExampleV1.ListUserExampleRequest{
					Params: &types.Params{
						Page:  0,
						Limit: 10,
						Sort:  "",
						Columns: []*types.Column{
							{
								Name:  "id",
								Exp:   ">=",
								Value: "1",
								Logic: "",
							},
						},
					},
				}
				total := 1000 // total number of requests

				b, err := benchmark.New(host, protoFile, "List", message, dependentProtoFilePath, total)
				if err != nil {
					return err
				}
				return b.Run()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if (err != nil) != tt.wantErr {
				t.Errorf("test '%s' error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
		})
	}
}
