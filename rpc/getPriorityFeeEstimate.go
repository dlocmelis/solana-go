// Copyright 2021 github.com/gagliardetto
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

package rpc

import (
	"context"
)

// GetTokenAccountBalance returns the token balance of an SPL Token account.
func (cl *Client) GetPriorityFeeEstimate(
	ctx context.Context,
	accounts []string,
	transactionData string, // Base-64 encoded Transaction or other encoded data specified by the transactionEncoding parameter.
	opts *GetPriorityFeeEstimateOpts,
) (out *GetPriorityFeeEstimateResult, err error) {
	params := []interface{}{}
	obj := M{}
	if len(transactionData) > 0 {
		obj["transaction"] = transactionData
	} else {
		obj["accountKeys"] = accounts
	}
	optionsObj := M{}
	if opts.PriorityLevel != "" {
		optionsObj["priorityLevel"] = opts.PriorityLevel
	}
	if opts.Recommended != nil {
		optionsObj["recommended"] = opts.Recommended
	}
	if opts.IncludeAllPriorityFeeLevels != nil {
		optionsObj["includeAllPriorityFeeLevels"] = opts.IncludeAllPriorityFeeLevels
	}
	if opts.TransactionEncoding != "" {
		optionsObj["transactionEncoding"] = opts.TransactionEncoding
	}
	if len(optionsObj) > 0 {
		obj["options"] = optionsObj
	}
	params = append(params, obj)
	err = cl.rpcClient.CallForInto(ctx, &out, "getPriorityFeeEstimate", params)
	return
}

type GetPriorityFeeEstimateOpts struct {
	// The priority level of the fee
	PriorityLevel PriorityLevel `json:"priorityLevel,omitempty"`

	// If true, return a recommended fee level.
	Recommended *bool `json:"recommended,omitempty"`

	// If true, return all possible fee levels.
	IncludeAllPriorityFeeLevels *bool `json:"includeAllPriorityFeeLevels,omitempty"`

	//Used to specify the encoding for the transaction. If not provided, the default encoding is base64.
	TransactionEncoding string `json:"transactionEncoding,omitempty"`
}

type GetPriorityFeeEstimateResult struct {
	PriorityFeeLevels   MicroLamportPriorityFeeLevels `json:"priorityFeeLevels"`
	PriorityFeeEstimate float64                       `json:"priorityFeeEstimate"`
}

// PriorityLevel is the priority level of the fee.
type MicroLamportPriorityFeeLevels struct {
	Min       float64 `json:"min"`
	Low       float64 `json:"low"`
	Medium    float64 `json:"medium"`
	High      float64 `json:"high"`
	VeryHigh  float64 `json:"veryHigh"`
	UnsafeMax float64 `json:"unsafeMax"`
}
