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

package processor

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/Jeffail/benthos/lib/metrics"
	"github.com/Jeffail/benthos/lib/types"
	"github.com/Jeffail/benthos/lib/util/service/log"
)

func TestArchiveBadAlgo(t *testing.T) {
	conf := NewConfig()
	conf.Archive.Format = "does not exist"

	testLog := log.NewLogger(os.Stdout, log.LoggerConfig{LogLevel: "NONE"})

	_, err := NewArchive(conf, nil, testLog, metrics.DudType{})
	if err == nil {
		t.Error("Expected error from bad algo")
	}
}

func TestArchiveTar(t *testing.T) {
	conf := NewConfig()
	conf.Archive.Format = "tar"

	testLog := log.NewLogger(os.Stdout, log.LoggerConfig{LogLevel: "NONE"})

	exp := [][]byte{
		[]byte("hello world first part"),
		[]byte("hello world second part"),
		[]byte("third part"),
		[]byte("fourth"),
		[]byte("5"),
	}

	proc, err := NewArchive(conf, nil, testLog, metrics.DudType{})
	if err != nil {
		t.Fatal(err)
	}

	msgs, res := proc.ProcessMessage(types.NewMessage(exp))
	if len(msgs) != 1 {
		t.Error("Archive failed")
	} else if res != nil {
		t.Errorf("Expected nil response: %v", res)
	}
	if msgs[0].Len() != 1 {
		t.Fatal("More parts than expected")
	}

	act := [][]byte{}

	buf := bytes.NewBuffer(msgs[0].Get(0))
	tr := tar.NewReader(buf)
	for {
		_, err = tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			t.Fatal(err)
		}

		newPartBuf := bytes.Buffer{}
		if _, err = newPartBuf.ReadFrom(tr); err != nil {
			t.Fatal(err)
		}

		act = append(act, newPartBuf.Bytes())
	}

	if !reflect.DeepEqual(exp, act) {
		t.Errorf("Unexpected output: %s != %s", act, exp)
	}
}

func TestArchiveBinary(t *testing.T) {
	conf := NewConfig()
	conf.Archive.Format = "binary"

	testLog := log.NewLogger(os.Stdout, log.LoggerConfig{LogLevel: "NONE"})
	proc, err := NewArchive(conf, nil, testLog, metrics.DudType{})
	if err != nil {
		t.Error(err)
		return
	}

	testMsg := types.NewMessage([][]byte{[]byte("hello"), []byte("world")})
	testMsgBlob := testMsg.Bytes()

	if msgs, _ := proc.ProcessMessage(testMsg); len(msgs) == 1 {
		if lParts := msgs[0].Len(); lParts != 1 {
			t.Errorf("Wrong number of parts returned: %v != %v", lParts, 1)
		}
		if !reflect.DeepEqual(testMsgBlob, msgs[0].Get(0)) {
			t.Errorf("Returned message did not match: %s != %s", msgs[0].Get(0), testMsgBlob)
		}
	} else {
		t.Error("Failed on good message")
	}
}

func TestArchiveEmpty(t *testing.T) {
	conf := NewConfig()

	testLog := log.NewLogger(os.Stdout, log.LoggerConfig{LogLevel: "NONE"})
	proc, err := NewArchive(conf, nil, testLog, metrics.DudType{})
	if err != nil {
		t.Error(err)
		return
	}

	msgs, _ := proc.ProcessMessage(types.NewMessage([][]byte{}))
	if len(msgs) != 0 {
		t.Error("Expected failure with zero part message")
	}
}
