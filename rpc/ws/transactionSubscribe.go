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

	"github.com/dlocmelis/solana-go"
	"github.com/dlocmelis/solana-go/rpc"
)

type TransactionResult struct {
	Transaction *rpc.TransactionWithMeta `json:"transaction"`
	Slot        uint64                   `json:"slot"`
	Signature   solana.Signature         `json:"signature"`
	Err         interface{}              `json:"err,omitempty"`
}

type TransactionSubscribeFilter struct {
	Vote            bool     `json:"vote"`
	Failed          bool     `json:"failed"`
	Signature       string   `json:"signature"`
	AccountInclude  []string `json:"accountInclude"`
	AccountExclude  []string `json:"accountExclude"`
	AccountRequired []string `json:"accountRequired"`
}

type TransactionSubscribeOpts struct {
	Commitment rpc.CommitmentType
	Encoding   solana.EncodingType `json:"encoding,omitempty"`

	// Level of transaction detail to return.
	TransactionDetails rpc.TransactionDetailsType

	// Whether to populate the rewards array. If parameter not provided, the default includes rewards.
	Rewards *bool

	// Max transaction version to return in responses.
	// If the requested block contains a transaction with a higher version, an error will be returned.
	MaxSupportedTransactionVersion *uint64
}

func (cl *Client) TransactionSubscribe(
	filter TransactionSubscribeFilter,
	opts *TransactionSubscribeOpts,
) (*TransactionSubscribtion, error) {
	var params []interface{}
	obj := make(rpc.M)
	if filter.Vote {
		obj["vote"] = filter.Vote
	}
	if filter.Failed {
		obj["failed"] = filter.Failed
	}
	if filter.Signature != "" {
		obj["signature"] = filter.Signature
	}
	if len(filter.AccountInclude) > 0 {
		obj["accountInclude"] = filter.AccountInclude
	}
	if len(filter.AccountExclude) > 0 {
		obj["accountExclude"] = filter.AccountExclude
	}
	if len(filter.AccountRequired) > 0 {
		obj["accountRequired"] = filter.AccountRequired
	}
	if len(obj) > 0 {
		params = append(params, obj)
	}
	if opts != nil {
		obj := make(rpc.M)
		if opts.Commitment != "" {
			obj["commitment"] = opts.Commitment
		}
		if opts.Encoding != "" {
			if !solana.IsAnyOfEncodingType(
				opts.Encoding,
				// Valid encodings:
				// solana.EncodingJSON, // TODO
				// solana.EncodingJSONParsed, // TODO
				solana.EncodingBase58,
				solana.EncodingBase64,
				solana.EncodingBase64Zstd,
			) {
				return nil, fmt.Errorf("provided encoding is not supported: %s", opts.Encoding)
			}
			obj["encoding"] = opts.Encoding
		}
		if opts.TransactionDetails != "" {
			obj["transaction_details"] = opts.TransactionDetails
		}
		if opts.Rewards != nil {
			obj["showRewards"] = opts.Rewards
		}
		if opts.MaxSupportedTransactionVersion != nil {
			obj["maxSupportedTransactionVersion"] = *opts.MaxSupportedTransactionVersion
		}
		if len(obj) > 0 {
			params = append(params, obj)
		}
	}
	genSub, err := cl.subscribe(
		params,
		nil,
		"transactionSubscribe",
		"transactionUnsubscribe",
		func(msg []byte) (interface{}, error) {
			var res TransactionResult
			err := decodeResponseFromMessage(msg, &res)
			return &res, err
		},
	)
	if err != nil {
		return nil, err
	}
	return &TransactionSubscribtion{
		sub: genSub,
	}, nil
}

type TransactionSubscribtion struct {
	sub *Subscription
}

func (sw *TransactionSubscribtion) Recv() (*TransactionResult, error) {
	select {
	case d := <-sw.sub.stream:
		return d.(*TransactionResult), nil
	case err := <-sw.sub.err:
		return nil, err
	}
}

func (sw *TransactionSubscribtion) Err() <-chan error {
	return sw.sub.err
}

func (sw *TransactionSubscribtion) Response() <-chan *TransactionResult {
	typedChan := make(chan *TransactionResult, 1)
	go func(ch chan *TransactionResult) {
		// TODO: will this subscription yield more than one result?
		d, ok := <-sw.sub.stream
		if !ok {
			return
		}
		ch <- d.(*TransactionResult)
	}(typedChan)
	return typedChan
}

func (sw *TransactionSubscribtion) Unsubscribe() {
	sw.sub.Unsubscribe()
}
