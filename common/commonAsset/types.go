/*
 * Copyright © 2021 Zecrey Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package commonAsset

import (
	"encoding/json"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"gorm.io/gorm"
)

type Liquidity struct {
	PairIndex int64
	AssetAId  int64
	AssetA    string
	AssetBId  int64
	AssetB    string
	LpAmount  string
}

type FormatAccountInfo struct {
	AccountId       uint
	AccountIndex    int64
	AccountName     string
	PublicKey       string
	AccountNameHash string
	L1Address       string
	Nonce           int64
	// map[int64]string
	AssetInfo map[int64]string
	AssetRoot string
	// map[int64]*Liquidity
	LiquidityInfo map[int64]*Liquidity
	LiquidityRoot string
}

func FromFormatAccountInfo(formatAccountInfo *FormatAccountInfo) (accountInfo *account.Account, err error) {
	assetInfoBytes, err := json.Marshal(formatAccountInfo.AssetInfo)
	if err != nil {
		return nil, err
	}
	liquidityInfoBytes, err := json.Marshal(formatAccountInfo.LiquidityInfo)
	if err != nil {
		return nil, err
	}
	accountInfo = &account.Account{
		Model: gorm.Model{
			ID: formatAccountInfo.AccountId,
		},
		AccountIndex:    formatAccountInfo.AccountIndex,
		AccountName:     formatAccountInfo.AccountName,
		PublicKey:       formatAccountInfo.PublicKey,
		AccountNameHash: formatAccountInfo.AccountNameHash,
		L1Address:       formatAccountInfo.L1Address,
		Nonce:           formatAccountInfo.Nonce,
		AssetInfo:       string(assetInfoBytes),
		AssetRoot:       formatAccountInfo.AssetRoot,
		LiquidityInfo:   string(liquidityInfoBytes),
		LiquidityRoot:   formatAccountInfo.LiquidityRoot,
	}
	return accountInfo, nil
}

func ToFormatAccountInfo(accountInfo *account.Account) (formatAccountInfo *FormatAccountInfo, err error) {
	var (
		assetInfo     map[int64]string
		liquidityInfo map[int64]*Liquidity
	)
	err = json.Unmarshal([]byte(accountInfo.AssetInfo), &assetInfo)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(accountInfo.LiquidityInfo), &liquidityInfo)
	if err != nil {
		return nil, err
	}
	formatAccountInfo = &FormatAccountInfo{
		AccountId:       accountInfo.ID,
		AccountIndex:    accountInfo.AccountIndex,
		AccountName:     accountInfo.AccountName,
		PublicKey:       accountInfo.PublicKey,
		AccountNameHash: accountInfo.AccountNameHash,
		L1Address:       accountInfo.L1Address,
		Nonce:           accountInfo.Nonce,
		AssetInfo:       assetInfo,
		AssetRoot:       accountInfo.AssetRoot,
		LiquidityInfo:   liquidityInfo,
		LiquidityRoot:   accountInfo.LiquidityRoot,
	}
	return formatAccountInfo, nil
}

type FormatAccountHistoryInfo struct {
	AccountId    uint
	AccountIndex int64
	Nonce        int64
	// map[int64]string
	AssetInfo map[int64]string
	AssetRoot string
	// map[int64]*Liquidity
	LiquidityInfo map[int64]*Liquidity
	LiquidityRoot string
	L2BlockHeight int64
}

func FromFormatAccountHistoryInfo(formatAccountInfo *FormatAccountHistoryInfo) (accountInfo *account.AccountHistory, err error) {
	assetInfoBytes, err := json.Marshal(formatAccountInfo.AssetInfo)
	if err != nil {
		return nil, err
	}
	liquidityInfoBytes, err := json.Marshal(formatAccountInfo.LiquidityInfo)
	if err != nil {
		return nil, err
	}
	accountInfo = &account.AccountHistory{
		Model: gorm.Model{
			ID: formatAccountInfo.AccountId,
		},
		AccountIndex:  formatAccountInfo.AccountIndex,
		Nonce:         formatAccountInfo.Nonce,
		AssetInfo:     string(assetInfoBytes),
		AssetRoot:     formatAccountInfo.AssetRoot,
		LiquidityInfo: string(liquidityInfoBytes),
		LiquidityRoot: formatAccountInfo.LiquidityRoot,
		L2BlockHeight: formatAccountInfo.L2BlockHeight,
	}
	return accountInfo, nil
}

func ToFormatAccountHistoryInfo(accountInfo *account.AccountHistory) (formatAccountInfo *FormatAccountHistoryInfo, err error) {
	var (
		assetInfo     map[int64]string
		liquidityInfo map[int64]*Liquidity
	)
	err = json.Unmarshal([]byte(accountInfo.AssetInfo), &assetInfo)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(accountInfo.LiquidityInfo), &liquidityInfo)
	if err != nil {
		return nil, err
	}
	formatAccountInfo = &FormatAccountHistoryInfo{
		AccountId:     accountInfo.ID,
		AccountIndex:  accountInfo.AccountIndex,
		Nonce:         accountInfo.Nonce,
		AssetInfo:     assetInfo,
		AssetRoot:     accountInfo.AssetRoot,
		LiquidityInfo: liquidityInfo,
		LiquidityRoot: accountInfo.LiquidityRoot,
		L2BlockHeight: accountInfo.L2BlockHeight,
	}
	return formatAccountInfo, nil
}
