package invoker

import (
	"fmt"
	"github.com/dapperlabs/flow-dps/models/dps"
	"github.com/onflow/flow-go/model/flow"
)

type Blocks struct {
	reader dps.Reader
}

func (b *Blocks) ByHeightFrom(height uint64, header *flow.Header) (*flow.Header, error) {

	if header != nil {
		_, err := b.reader.HeightForBlock(header.ID())
		if err != nil {
			return nil, fmt.Errorf("cannot get height for block: %w", err)
		}
	}

	return b.reader.Header(height)
}
