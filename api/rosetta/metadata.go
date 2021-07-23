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

package rosetta

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/onflow/flow-go-sdk"

	"github.com/optakt/flow-dps/rosetta/identifier"
	"github.com/optakt/flow-dps/rosetta/object"
)

type MetadataRequest struct {
	NetworkID identifier.Network `json:"network_identifier"`
	Options   object.Options     `json:"options"`
}

type MetadataResponse struct {
	Metadata object.Metadata `json:"metadata"`
}

func (c *Construction) Metadata(ctx echo.Context) error {

	var req MetadataRequest
	err := ctx.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidEncoding(invalidJSON, err))
	}

	if req.NetworkID.Blockchain == "" {
		return echo.NewHTTPError(http.StatusBadRequest, invalidFormat(blockchainEmpty))
	}
	if req.NetworkID.Network == "" {
		return echo.NewHTTPError(http.StatusBadRequest, invalidFormat(networkEmpty))
	}

	if req.Options.AccountID.Address == "" {
		return echo.NewHTTPError(http.StatusBadRequest, invalidFormat(addressEmpty))
	}

	current, _, err := c.retriever.Current()
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("could not retrieve last block info: %w", err))
	}

	proposer := flow.HexToAddress(req.Options.AccountID.Address)
	sequenceNr, err := getAccountSequenceNumber(proposer)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("could not retrieve account sequence number: %w", err))
	}

	res := MetadataResponse{
		Metadata: object.Metadata{
			ReferenceBlockID: current,
			SequenceNumber:   sequenceNr,
		},
	}

	return ctx.JSON(http.StatusOK, res)
}

// TODO: implement getAccountSequenceNr()
func getAccountSequenceNumber(address flow.Address) (uint64, error) {
	return 0, fmt.Errorf("TBD: not implemented")
}
