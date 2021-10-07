package main

import (
	"io/ioutil"

	"github.com/fxamacker/cbor/v2"
	"github.com/onflow/flow-go/engine/execution/computation/computer/uploader"
)

func main() {

	data, err := ioutil.ReadFile("/home/m4ks/13591cbbcc21268c7bb7b8297df917eb495e41e60079dfadefc931db777feef8_new.cbor")

	if err != nil {
		panic(err)
	}

	var record uploader.BlockData
	err = cbor.Unmarshal(data, &record)
	if err != nil {
		panic(err)
	}
}
