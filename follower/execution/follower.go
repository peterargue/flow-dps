// Copyright 2021 Optakt Labs OÜ
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package execution

import (
	"fmt"
	"math"

	"github.com/dgraph-io/badger/v2"
	"github.com/onflow/flow-go/ledger"
	"github.com/optakt/flow-dps/codec/zbor"
	"github.com/optakt/flow-dps/models/dps"
	"github.com/rs/zerolog"
)

type blockReader interface {
	Next() ([]byte, error)
}

type Follower struct {
	log zerolog.Logger

	codec  zbor.Codec
	db     *badger.DB
	blocks blockReader

	block  BlockData
	height uint64
}

func New(log zerolog.Logger, blocks blockReader, db *badger.DB) *Follower {
	f := Follower{
		log: log,

		blocks: blocks,
		db:     db,
		height: math.MaxUint64,
	}

	return &f
}

func (f *Follower) Height() uint64 {
	return f.height
}

func (f *Follower) Update() (*ledger.TrieUpdate, error) {
	if len(f.block.TrieUpdates) == 0 {
		// No more trie updates are available.
		return nil, dps.ErrTimeout
	}

	// Copy next update to be returned.
	update := f.block.TrieUpdates[0]

	// Move the slice forward by popping the first element.
	f.block.TrieUpdates = f.block.TrieUpdates[1:]

	return update, nil
}

// FIXME: How to know for sure that everything has been read before calling Next?

func (f *Follower) Next() error {
	b, err := f.blocks.Next()
	if err != nil {
		return fmt.Errorf("could not retrieve block: %w", err)
	}

	err = f.codec.Unmarshal(b, &f.block)
	if err != nil {
		return fmt.Errorf("could not unmarshal block: %w", err)
	}

	// FIXME: Index everything that is needed.

	return nil
}
