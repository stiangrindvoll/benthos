// Copyright (c) 2018 Ashley Jeffs
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package reader

import (
	"net/http"
	"sync"
	"time"

	"github.com/Jeffail/benthos/lib/metrics"
	"github.com/Jeffail/benthos/lib/types"
	"github.com/Jeffail/benthos/lib/util/http/auth"
	"github.com/Jeffail/benthos/lib/util/service/log"
	"github.com/gorilla/websocket"
)

//------------------------------------------------------------------------------

// WebsocketConfig is configuration for the Websocket input type.
type WebsocketConfig struct {
	URL         string `json:"url" yaml:"url"`
	auth.Config `json:",inline" yaml:",inline"`
}

// NewWebsocketConfig creates a new WebsocketConfig with default values.
func NewWebsocketConfig() WebsocketConfig {
	return WebsocketConfig{
		URL:    "ws://localhost:4195/get/ws",
		Config: auth.NewConfig(),
	}
}

//------------------------------------------------------------------------------

// Websocket is an input type that reads Websocket messages.
type Websocket struct {
	log   log.Modular
	stats metrics.Type

	lock *sync.Mutex

	conf   WebsocketConfig
	client *websocket.Conn
}

// NewWebsocket creates a new Websocket input type.
func NewWebsocket(
	conf WebsocketConfig,
	log log.Modular,
	stats metrics.Type,
) (*Websocket, error) {
	ws := &Websocket{
		log:   log.NewModule(".input.websocket"),
		stats: stats,
		lock:  &sync.Mutex{},
		conf:  conf,
	}
	return ws, nil
}

//------------------------------------------------------------------------------

func (w *Websocket) getWS() *websocket.Conn {
	w.lock.Lock()
	ws := w.client
	w.lock.Unlock()
	return ws
}

//------------------------------------------------------------------------------

// Connect establishes a connection to an Websocket server.
func (w *Websocket) Connect() error {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.client != nil {
		return nil
	}

	headers := http.Header{}

	if err := w.conf.Sign(&http.Request{
		Header: headers,
	}); err != nil {
		return err
	}

	client, _, err := websocket.DefaultDialer.Dial(w.conf.URL, headers)
	if err != nil {
		return err
	}

	w.client = client
	return nil
}

//------------------------------------------------------------------------------

// Read attempts to read a new message from the websocket.
func (w *Websocket) Read() (types.Message, error) {
	client := w.getWS()
	if client == nil {
		return nil, types.ErrNotConnected
	}

	_, data, err := client.ReadMessage()
	if err != nil {
		w.lock.Lock()
		w.client = nil
		w.lock.Unlock()
		err = types.ErrNotConnected
		return nil, err
	}

	return types.NewMessage([][]byte{data}), nil
}

// Acknowledge instructs whether the pending messages were propagated
// successfully.
func (w *Websocket) Acknowledge(err error) error {
	return nil
}

// CloseAsync shuts down the Websocket input and stops reading messages.
func (w *Websocket) CloseAsync() {
	w.lock.Lock()
	if w.client != nil {
		w.client.Close()
		w.client = nil
	}
	w.lock.Unlock()
}

// WaitForClose blocks until the Websocket input has closed down.
func (w *Websocket) WaitForClose(timeout time.Duration) error {
	return nil
}

//------------------------------------------------------------------------------
