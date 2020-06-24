/*
   Copyright 2020 Docker, Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package proxy

import (
	"github.com/containerd/containerd/log"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"

	streamsv1 "github.com/docker/api/protos/streams/v1"
	"github.com/docker/api/server/proxy/streams"
)

func (p *proxy) NewStream(stream streamsv1.Streaming_NewStreamServer) error {
	var (
		ctx = stream.Context()
		id  = uuid.New().String()
	)
	md := metadata.New(map[string]string{
		"id": id,
	})

	// return the id of the stream to the client
	if err := stream.SendHeader(md); err != nil {
		return err
	}

	errc := make(chan error)

	p.mu.Lock()
	p.streams[id] = &streams.Stream{
		Streaming_NewStreamServer: stream,
		ErrChan:                   errc,
	}
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		delete(p.streams, id)
		p.mu.Unlock()
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		log.G(ctx).Debug("client context canceled")
		return ctx.Err()
	}
}