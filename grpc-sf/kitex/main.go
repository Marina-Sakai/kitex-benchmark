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
	"fmt"
	"log"
	"net"

	"github.com/cloudwego/kitex/server"

	"github.com/cloudwego/kitex-benchmark/codec/protobuf/kitex_gen/cache"
	cachesvr "github.com/cloudwego/kitex-benchmark/codec/protobuf/kitex_gen/cache/runtimecache"
	"github.com/cloudwego/kitex-benchmark/perf"
	"github.com/cloudwego/kitex-benchmark/runner-sf"
)

const port = 8006

var (
	_ cache.RuntimeCache = &OperateCacheStringImpl{}

	recorder = perf.NewRecorder("KITEX@Server")
)

type OperateCacheStringImpl struct{}

func (o *OperateCacheStringImpl) OperateCacheString(ctx context.Context, req *cache.OpCacheStringRequest) (*cache.OpCacheResponse, error) {
	resp := runner.ProcessRequest(recorder, req.Command, req.Name)

	return &cache.OpCacheResponse{
		Success: resp.Success,
		Data:    resp.Data,
	}, nil
}

func main() {
	// start pprof server
	go func() {
		perf.ServeMonitor(fmt.Sprintf(":%d", port+10000))
	}()
	svr := cachesvr.NewServer(
		new(OperateCacheStringImpl),
		server.WithServiceAddr(&net.TCPAddr{IP: net.IPv4zero, Port: port}),
	)

	if err := svr.Run(); err != nil {
		log.Println(err.Error())
	}
}
