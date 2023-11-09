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

	"github.com/cloudwego/kitex-benchmark/codec/protobuf/kitex_gen/echo"
	echosvr "github.com/cloudwego/kitex-benchmark/codec/protobuf/kitex_gen/echo/echo"
	sechosvr "github.com/cloudwego/kitex-benchmark/codec/protobuf/kitex_gen/echo/secho"
	"github.com/cloudwego/kitex-benchmark/perf"
	"github.com/cloudwego/kitex-benchmark/runner"
)

const port = 8006

var (
	_ echo.Echo  = &EchoImpl{}
	_ echo.SEcho = &SEchoImpl{}

	recorder = perf.NewRecorder("KITEX@Server")
)

// EchoImpl implements the last service interface defined in the IDL.
type EchoImpl struct{}

// Echo implements the EchoImpl interface.
func (s *EchoImpl) Echo(ctx context.Context, req *echo.Request) (*echo.Response, error) {
	resp := runner.ProcessRequest(recorder, req.Action, req.Msg)

	return &echo.Response{
		Action: resp.Action,
		Msg:    resp.Msg,
	}, nil
}

type SEchoImpl struct{}

// Echo implements the EchoImpl interface.
func (s *SEchoImpl) Echo(stream echo.SEcho_EchoServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		resp := runner.ProcessRequest(recorder, req.Action, req.Msg)

		err = stream.Send(&echo.Response{
			Action: resp.Action,
			Msg:    resp.Msg,
		})
		if err != nil {
			return err
		}
	}
}

func main() {
	// start pprof server
	go func() {
		perf.ServeMonitor(fmt.Sprintf(":%d", port+10000))
	}()
	svr := server.NewServer(server.WithServiceAddr(&net.TCPAddr{IP: net.IPv4zero, Port: port}))
	svr.RegisterService(echosvr.NewServiceInfo(), new(EchoImpl))
	svr.RegisterService(sechosvr.NewServiceInfo(), new(SEchoImpl))

	if err := svr.Run(); err != nil {
		log.Println(err.Error())
	}
}
