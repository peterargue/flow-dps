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

type consensusFollower interface {
	Height() uint64
}

type Follower struct {
	log zerolog.Logger

	consensus consensusFollower
	codec     zbor.Codec
	db        *badger.DB
	blocks    blockReader

	block BlockData
}

func New(log zerolog.Logger, blocks blockReader, db *badger.DB) *Follower {
	f := Follower{
		log: log,

		blocks: blocks,
		db:     db,
	}

	return &f
}

func (f *Follower) Height() uint64 {
	if f.block.Block != nil {
		return f.block.Block.Header.Height
	}
	return math.MaxUint64
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
// FIXME: Only index once the block is finalized (height is available in the consensus follower)

func (f *Follower) Next() error {
	// Do not advance unless the consensus moved forward.
	if f.consensus.Height() == f.Height() {
		return nil
	}

	// FIXME: Change f.blocks to download a specific block by BlockID, the one
	// that the consensus got.

	// FIXME: Make the consensus follower return the latest blockID.

	b, err := f.blocks.Next()
	if err != nil {
		return fmt.Errorf("could not retrieve block: %w", err)
	}

	err = f.codec.Unmarshal(b, &f.block)
	if err != nil {
		return fmt.Errorf("could not unmarshal block: %w", err)
	}

	// if f.block.Block.ID() == blockID {
	// 	break
	// }
	//
	// err = f.IndexAll()
	// if err != nil {
	// 	return fmt.Errorf("could not index execution state data: %w", err)
	// }

	return nil
}

func (f *Follower) IndexAll() {
}
