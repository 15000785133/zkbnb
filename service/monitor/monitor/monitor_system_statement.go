/*
 * Copyright © 2021 ZkBNB Protocol
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
 */

package monitor

import (
	"github.com/zeromicro/go-zero/core/logx"
)

const DefaultBlockProcessThreshold = 8

func (m *Monitor) MonitorSystemAndBusiness() error {
	err := m.monitorBlockHeightProcess()
	if err != nil {
		return err
	}
	return nil
}

func (m *Monitor) monitorBlockHeightProcess() (err error) {
	// Try to get the latest block height in all the blocks
	curBlockHeight, err := m.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		return err
	}

	// Try to get the latest block height in verified blocks
	verifiedBlockHeight, err := m.BlockModel.GetLatestVerifiedHeight()
	if err != nil {
		return err
	}

	// Get the block process threshold from the configuration
	blockProcessThreshold := m.Config.ChainConfig.BlockProcessThreshold
	if blockProcessThreshold == 0 {
		blockProcessThreshold = DefaultBlockProcessThreshold
	}

	// Calculate the unverified block count
	unverifiedBlockCount := curBlockHeight - verifiedBlockHeight

	//If the unverified block count is greater than the
	//threshold, output the log to warn this progress
	if unverifiedBlockCount > blockProcessThreshold {
		logx.Error("Unverified Block Count is greater than threshold : unverifiedBlockCount:%d, "+
			"blockProcessThreshold:%d", unverifiedBlockCount, blockProcessThreshold)
	} else {
		logx.Infof("Unverified Block Count is less than or equal to threshold : unverifiedBlockCount:%d, "+
			"blockProcessThreshold:%d", unverifiedBlockCount, blockProcessThreshold)
	}
	return nil
}
