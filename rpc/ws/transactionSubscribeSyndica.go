// Copyright 2022 github.com/gagliardetto
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ws

import (
	"fmt"
	"time"

	"github.com/dlocmelis/solana-go"
	"github.com/dlocmelis/solana-go/rpc"
)

type TransactionSyndicaResult struct {
	Value   *rpc.TransactionWithMeta `json:"value"`
	Context Context                  `json:"context"`
	Err     interface{}              `json:"err,omitempty"`
}

type Context struct {
	Index      uint64           `json:"index"`
	IsVote     bool             `json:"isVote"`
	NodeTime   time.Time        `json:"nodeTime"`
	Signature  solana.Signature `json:"signature"`
	SlotStatus string           `json:"slotStatus"`
}

type TransactionSyndicaParams struct {
	Network  string                            `json:"network"`
	Verified bool                              `json:"verified"`
	Filter   TransactionSyndicaSubscribeFilter `json:"filter"`
}

type AccountKeys struct {
	All     []string `json:"all"`
	OneOf   []string `json:"oneOf"`
	Exclude []string `json:"exclude"`
}

type TransactionSyndicaSubscribeFilter struct {
	ExcludeVotes bool        `json:"excludeVotes"`
	AccountKeys  AccountKeys `json:"accountKeys"`
}

func (cl *Client) TransactionSyndicaSubscribe(params *TransactionSyndicaParams) (*TransactionSyndicaSubscription, error) {
	if params == nil {
		return nil, fmt.Errorf("params are required")
	}
	genSub, err := cl.subscribeSyndica(
		params,
		"chainstream.transactionsSubscribe",
		"chainstream.transactionsUnsubscribe",
		func(msg []byte) (interface{}, error) {
			var res TransactionResult
			err := decodeResponseFromMessage(msg, &res)
			return &res, err
		},
	)
	if err != nil {
		return nil, err
	}
	return &TransactionSyndicaSubscription{
		sub: genSub,
	}, nil
}

type TransactionSyndicaSubscription struct {
	sub *Subscription
}

func (sw *TransactionSyndicaSubscription) Recv() (*TransactionSyndicaResult, error) {
	select {
	case d := <-sw.sub.stream:
		return d.(*TransactionSyndicaResult), nil
	case err := <-sw.sub.err:
		return nil, err
	}
}

func (sw *TransactionSyndicaSubscription) Err() <-chan error {
	return sw.sub.err
}

func (sw *TransactionSyndicaSubscription) Response() <-chan *TransactionSyndicaResult {
	typedChan := make(chan *TransactionSyndicaResult, 1)
	go func(ch chan *TransactionSyndicaResult) {
		// TODO: will this subscription yield more than one result?
		d, ok := <-sw.sub.stream
		if !ok {
			return
		}
		ch <- d.(*TransactionSyndicaResult)
	}(typedChan)
	return typedChan
}

func (sw *TransactionSyndicaSubscription) Unsubscribe() {
	sw.sub.Unsubscribe()
}
