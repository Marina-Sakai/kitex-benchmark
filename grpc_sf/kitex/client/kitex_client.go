/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"sync"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex-benchmark/codec/protobuf/kitex_gen/cache"
	cachesvr "github.com/cloudwego/kitex-benchmark/codec/protobuf/kitex_gen/cache/runtimecache"
	"github.com/cloudwego/kitex-benchmark/runner_sf"
)

func NewKClient(opt *runner.Options) runner.Client {
	klog.SetLevel(klog.LevelWarn)
	return &kClient{
		client: cachesvr.MustNewClient("test.cache.kitex",
			client.WithHostPorts(opt.Address),
			client.WithTransportProtocol(transport.GRPC),
			client.WithGRPCConnPoolSize(6), // the cpu cores of server is 4, and 4*3/2 = 6
		),
		reqPool: &sync.Pool{
			New: func() interface{} {
				return &cache.OpCacheStringRequest{}
			},
		},
	}
}

type kClient struct {
	client  cachesvr.Client
	reqPool *sync.Pool
}

func (cli *kClient) OperateCacheString(action, msg string) error {
	ctx := context.Background()
	req := cli.reqPool.Get().(*cache.OpCacheStringRequest)
	defer cli.reqPool.Put(req)

	req.Command = action
	req.Name = msg

	reply, err := cli.client.OperateCacheString(ctx, req)
	if reply != nil {
		runner.ProcessResponse(reply.Success, reply.Data)
	}
	return err
}
