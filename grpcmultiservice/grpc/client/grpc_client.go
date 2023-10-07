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
	"errors"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"

	grpcg "github.com/cloudwego/kitex-benchmark/codec/protobuf/multi_grpc_gen"
	"github.com/cloudwego/kitex-benchmark/runner"
)

const content = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Suspendisse potenti."

func NewPBGrpcClient(opt *runner.Options) runner.Client {
	cli := &pbGrpcClient{}
	cli.reqPool = &sync.Pool{
		New: func() interface{} {
			return &grpcg.Request{}
		},
	}
	cli.connpool = runner.NewPool(func() interface{} {
		// Set up a connection to the server.
		conn, err := grpc.Dial(opt.Address, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		return grpcg.NewEchoClient(conn)
	}, opt.PoolSize)
	return cli
}

func NewGrpcClient(opt *runner.Options) runner.Client {
	conn, err := grpc.Dial(opt.Address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	client := grpcg.NewSEchoClient(conn)
	return &grpcSClient{
		client: client,
		streampool: &sync.Pool{New: func() interface{} {
			stream, err := client.Echo(context.Background())
			if err != nil {
				log.Printf("client new stream failed: %v", err)
				return nil
			}
			return stream
		}},
		reqPool: &sync.Pool{
			New: func() interface{} {
				return &grpcg.Request{}
			},
		},
	}
}

type pbGrpcClient struct {
	reqPool  *sync.Pool
	connpool *runner.Pool
}

func (cli *pbGrpcClient) Echo(action, msg string) error {
	ctx := context.Background()
	req := cli.reqPool.Get().(*grpcg.Request)
	defer cli.reqPool.Put(req)

	req.Action = action
	req.Msg = msg
	req.Content = content

	pbcli := cli.connpool.Get().(grpcg.EchoClient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	reply, err := pbcli.Echo(ctx, req)

	if reply != nil {
		runner.ProcessResponse(reply.Action, reply.Msg)
	}
	return err
}

type grpcSClient struct {
	client     grpcg.SEchoClient
	streampool *sync.Pool
	reqPool    *sync.Pool
}

func (cli *grpcSClient) Echo(action, msg string) (err error) {
	req := cli.reqPool.Get().(*grpcg.Request)
	defer cli.reqPool.Put(req)

	stream, ok := cli.streampool.Get().(grpcg.SEcho_EchoClient)
	if !ok {
		return errors.New("new stream error")
	}
	defer cli.streampool.Put(stream)
	req.Action = action
	req.Msg = msg
	req.Content = content
	if err := stream.Send(req); err != nil {
		return err
	}
	resp, err := stream.Recv()
	if err != nil {
		return err
	}
	runner.ProcessResponse(resp.Action, resp.Msg)
	return nil
}
