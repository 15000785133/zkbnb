package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/plugin/dbresolver"
	"math/big"
	"strconv"
	"time"

	"gorm.io/gorm/logger"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	bsmt "github.com/bnb-chain/zkbnb-smt"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/common/zkbnbprometheus"
	"github.com/bnb-chain/zkbnb/core/statedb"
	sdb "github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/asset"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/compressedblock"
	"github.com/bnb-chain/zkbnb/dao/dbcache"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/dao/sysconfig"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/tree"
	"github.com/bnb-chain/zkbnb/types"
)

// metrics
var (
	updateAccountTreeMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "update_asset_smt",
		Help:      "update asset smt tree operation time",
	})

	commitAccountTreeMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "commit_account_smt",
		Help:      "commit account smt tree operation time",
	})

	executeTxPrepareMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_prepare_time",
		Help:      "execute txs prepare operation time",
	})

	executeTxVerifyInputsMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_verify_inputs_time",
		Help:      "execute txs verify inputs operation time",
	})

	executeGenerateTxDetailsMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_generate_tx_details_time",
		Help:      "execute txs generate tx details operation time",
	})

	executeTxApplyTransactionMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_apply_transaction_time",
		Help:      "execute txs apply transaction operation time",
	})

	executeTxGeneratePubDataMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_generate_pub_data_time",
		Help:      "execute txs generate pub data operation time",
	})
	executeTxGetExecutedTxMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_get_executed_tx_time",
		Help:      "execute txs get executed tx operation time",
	})
)

type ChainConfig struct {
	Postgres struct {
		MasterDataSource string
		SlaveDataSource  string
		LogLevel         logger.LogLevel `json:",optional"`
	}
	CacheRedis cache.CacheConf
	//nolint:staticcheck
	CacheConfig statedb.CacheConfig `json:",optional"`
	TreeDB      struct {
		Driver tree.Driver
		//nolint:staticcheck
		LevelDBOption tree.LevelDBOption `json:",optional"`
		//nolint:staticcheck
		RedisDBOption tree.RedisDBOption `json:",optional"`
		//nolint:staticcheck
		RoutinePoolSize    int `json:",optional"`
		AssetTreeCacheSize int
	}
}

type BlockChain struct {
	*sdb.ChainDB
	Statedb *sdb.StateDB // Cache for current block changes.

	chainConfig *ChainConfig
	dryRun      bool //dryRun mode is used for verifying user inputs, is not for execution

	currentBlock *block.Block
	processor    Processor
}

func NewBlockChain(config *ChainConfig, moduleName string) (*BlockChain, error) {
	masterDataSource := config.Postgres.MasterDataSource
	slaveDataSource := config.Postgres.SlaveDataSource
	db, err := gorm.Open(postgres.Open(config.Postgres.MasterDataSource), &gorm.Config{
		Logger: logger.Default.LogMode(config.Postgres.LogLevel),
	})

	db.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{postgres.Open(masterDataSource)},
		Replicas: []gorm.Dialector{postgres.Open(slaveDataSource)},
	}))

	if err != nil {
		logx.Error("gorm connect db failed: ", err)
		return nil, err
	}
	bc := &BlockChain{
		ChainDB:     sdb.NewChainDB(db),
		chainConfig: config,
	}

	rollback(bc)

	curHeight, err := bc.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		logx.Error("get current block height failed: ", err)
		panic("get current block height failed: " + err.Error())
	}
	logx.Infof("get current block height: %d", curHeight)
	bc.currentBlock, err = bc.BlockModel.GetBlockByHeight(curHeight)
	if err != nil {
		return nil, err
	}
	if bc.currentBlock.BlockStatus == block.StatusProposing || bc.currentBlock.BlockStatus == block.StatusPacked {
		logx.Errorf("current block status is StatusProposing or StatusPacked,invalid block, height=%d", bc.currentBlock.BlockHeight)
		panic("current block status is StatusProposing or StatusPacked,invalid block, height=" + strconv.FormatInt(bc.currentBlock.BlockHeight, 10))
	}

	redisCache := dbcache.NewRedisCache(config.CacheRedis[0].Host, config.CacheRedis[0].Pass, 15*time.Minute)
	treeCtx, err := tree.NewContext(moduleName, config.TreeDB.Driver, false, config.TreeDB.RoutinePoolSize, &config.TreeDB.LevelDBOption, &config.TreeDB.RedisDBOption)
	if err != nil {
		return nil, err
	}

	treeCtx.SetOptions(bsmt.BatchSizeLimit(3 * 1024 * 1024))
	bc.Statedb, err = sdb.NewStateDB(treeCtx, bc.ChainDB, redisCache, &config.CacheConfig, config.TreeDB.AssetTreeCacheSize, bc.currentBlock.StateRoot, curHeight)
	if err != nil {
		return nil, err
	}
	bc.Statedb.PreviousStateRoot = bc.currentBlock.StateRoot
	bc.Statedb.UpdatePrunedBlockHeight(curHeight)

	accountFromDbGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "account_from_db_time",
		Help:      "account from db time",
	})

	accountGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "account_time",
		Help:      "account time",
	})

	verifyGasGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "verifyGasGauge_time",
		Help:      "verifyGas time",
	})
	verifySignature := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "verifySignature_time",
		Help:      "verifySignature time",
	})

	accountTreeMultiSetGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "accountTreeMultiSetGauge_time",
		Help:      "accountTreeMultiSetGauge time",
	})

	if err := prometheus.Register(verifyGasGauge); err != nil {
		return nil, fmt.Errorf("prometheus.Register verifyGasGauge error: %v", err)
	}

	if err := prometheus.Register(verifySignature); err != nil {
		return nil, fmt.Errorf("prometheus.Register verifySignature error: %v", err)
	}

	if err := prometheus.Register(accountTreeMultiSetGauge); err != nil {
		return nil, fmt.Errorf("prometheus.Register accountTreeMultiSetGauge error: %v", err)
	}

	if err := prometheus.Register(accountFromDbGauge); err != nil {
		return nil, fmt.Errorf("prometheus.Register accountFromDbMetrics error: %v", err)
	}
	getAccountCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "get_account_counter",
		Help:      "get account counter",
	})
	if err := prometheus.Register(getAccountCounter); err != nil {
		return nil, fmt.Errorf("prometheus.Register getAccountCounter error: %v", err)
	}

	getAccountFromDbCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "get_account_from_db_counter",
		Help:      "get account from db counter",
	})
	if err := prometheus.Register(getAccountFromDbCounter); err != nil {
		return nil, fmt.Errorf("prometheus.Register getAccountFromDbCounter error: %v", err)
	}

	accountTreeTimeGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "zkbnb",
			Name:      "get_account_tree_time",
			Help:      "get_account_tree_time.",
		},
		[]string{"type"})
	if err := prometheus.Register(accountTreeTimeGauge); err != nil {
		return nil, fmt.Errorf("prometheus.Register accountTreeTimeGauge error: %v", err)
	}

	nftTreeTimeGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "zkbnb",
			Name:      "get_nft_tree_time",
			Help:      "get_nft_tree_time.",
		},
		[]string{"type"})
	if err := prometheus.Register(nftTreeTimeGauge); err != nil {
		return nil, fmt.Errorf("prometheus.Register nftTreeTimeGauge error: %v", err)
	}

	stateDBMetrics := &zkbnbprometheus.StateDBMetrics{
		GetAccountFromDbGauge:    accountFromDbGauge,
		GetAccountGauge:          accountGauge,
		GetAccountCounter:        getAccountCounter,
		GetAccountFromDbCounter:  getAccountFromDbCounter,
		VerifyGasGauge:           verifyGasGauge,
		VerifySignature:          verifySignature,
		AccountTreeGauge:         accountTreeTimeGauge,
		NftTreeGauge:             nftTreeTimeGauge,
		AccountTreeMultiSetGauge: accountTreeMultiSetGauge,
	}
	bc.Statedb.Metrics = stateDBMetrics

	if err := prometheus.Register(executeTxPrepareMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register executeTxPrepareMetrics error: %v", err)
	}

	if err := prometheus.Register(executeTxVerifyInputsMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register executeTxVerifyInputsMetrics error: %v", err)
	}

	if err := prometheus.Register(executeGenerateTxDetailsMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register executeGenerateTxDetailsMetrics error: %v", err)
	}

	if err := prometheus.Register(executeTxApplyTransactionMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register executeTxApplyTransactionMetrics error: %v", err)
	}

	if err := prometheus.Register(executeTxGeneratePubDataMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register executeTxGeneratePubDataMetrics error: %v", err)
	}

	if err := prometheus.Register(executeTxGetExecutedTxMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register executeTxGetExecutedTxMetrics error: %v", err)
	}
	prometheusMetrics := &zkbnbprometheus.Metrics{
		TxPrepareMetrics:           executeTxPrepareMetrics,
		TxVerifyInputsMetrics:      executeTxVerifyInputsMetrics,
		TxGenerateTxDetailsMetrics: executeGenerateTxDetailsMetrics,
		TxApplyTransactionMetrics:  executeTxApplyTransactionMetrics,
		TxGeneratePubDataMetrics:   executeTxGeneratePubDataMetrics,
		TxGetExecutedTxMetrics:     executeTxGetExecutedTxMetrics,
	}
	bc.processor = NewCommitProcessor(bc, prometheusMetrics)

	// register metrics
	if err := prometheus.Register(updateAccountTreeMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register updateAccountTreeMetrics error: %v", err)
	}
	if err := prometheus.Register(commitAccountTreeMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register commitAccountTreeMetrics error: %v", err)
	}

	return bc, nil
}

// NewBlockChainForDryRun - for dry run mode, we can reuse existing models for quick creation
// , e.g., for sending tx, we can create blockchain for each request quickly
func NewBlockChainForDryRun(accountModel account.AccountModel,
	nftModel nft.L2NftModel, txPoolModel tx.TxPoolModel, assetModel asset.AssetModel,
	sysConfigModel sysconfig.SysConfigModel, redisCache dbcache.Cache) (*BlockChain, error) {
	chainDb := &sdb.ChainDB{
		AccountModel:     accountModel,
		L2NftModel:       nftModel,
		TxPoolModel:      txPoolModel,
		L2AssetInfoModel: assetModel,
		SysConfigModel:   sysConfigModel,
	}
	statedb, err := sdb.NewStateDBForDryRun(redisCache, &statedb.DefaultCacheConfig, chainDb)
	if err != nil {
		return nil, err
	}
	bc := &BlockChain{
		ChainDB: chainDb,
		dryRun:  true,
		Statedb: statedb,
	}
	bc.processor = NewAPIProcessor(bc)
	return bc, nil
}

func rollback(bc *BlockChain) {
	curHeight, err := bc.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		logx.Error("get current block height failed: ", err)
		panic("get current block height failed: " + err.Error())
	}
	logx.Infof("get current block height: %d", curHeight)
	blocks, err := bc.BlockModel.GetBlockByStatus([]int{block.StatusProposing, block.StatusPacked})
	if err != nil {
		logx.Error("get blocks by status (StatusProposing,StatusPacked) failed: ", err)
		panic("get blocks by status (StatusProposing,StatusPacked) failed: " + err.Error())
	}
	accountIndexMap := make(map[int64]bool, 0)
	nftIndexMap := make(map[int64]bool, 0)
	heights := make([]int64, 0)
	packedHeights := make([]int64, 0)
	if blocks != nil {
		for _, blockInfo := range blocks {
			if blockInfo.AccountIndexes != "[]" && blockInfo.AccountIndexes != "" {
				var accountIndexes []int64
				err = json.Unmarshal([]byte(blockInfo.AccountIndexes), &accountIndexes)
				if err != nil {
					logx.Error("json err unmarshal failed")
					panic("json err unmarshal failed: " + err.Error())
				}
				for _, accountIndex := range accountIndexes {
					accountIndexMap[accountIndex] = true
				}
			}
			if blockInfo.NftIndexes != "[]" && blockInfo.NftIndexes != "" {
				var nftIndexes []int64
				err = json.Unmarshal([]byte(blockInfo.NftIndexes), &nftIndexes)
				if err != nil {
					logx.Error("json err unmarshal failed")
					panic("json err unmarshal failed: " + err.Error())
				}
				for _, nftIndex := range nftIndexes {
					nftIndexMap[nftIndex] = true
				}
			}
			heights = append(heights, blockInfo.BlockHeight)
			if blockInfo.BlockStatus == block.StatusPacked {
				packedHeights = append(packedHeights, blockInfo.BlockHeight)
			}
		}
	}
	if len(heights) > 0 {
		accountIndexList := make([]int64, 0)
		for k := range accountIndexMap {
			accountIndexList = append(accountIndexList, k)
		}

		nftIndexList := make([]int64, 0)
		for k := range nftIndexMap {
			nftIndexList = append(nftIndexList, k)
		}

		height, err := bc.BlockModel.GetLatestPendingHeight()
		if err != nil {
			logx.Error("get latest pending height failed: ", err)
			panic("get latest pending height failed: " + err.Error())
		}
		accountIndexSlice := make([]int64, 0)
		accountHistories := make([]*account.AccountHistory, 0)
		accountIndexLen := len(accountIndexList)
		for _, accountIndex := range accountIndexList {
			accountIndexLen--
			accountIndexSlice = append(accountIndexSlice, accountIndex)
			if len(accountIndexSlice) == 100 || accountIndexLen == 0 {
				_, accountHistoryList, err := bc.AccountHistoryModel.GetLatestAccountHistories(accountIndexSlice, height)
				if err != nil {
					logx.Error("get latest account histories failed: ", err)
					panic("get latest account histories failed: " + err.Error())
				}
				if accountHistoryList != nil {
					accountHistories = append(accountHistories, accountHistoryList...)
				}
				accountIndexSlice = make([]int64, 0)
			}
		}
		deleteAccountIndexes := make([]int64, 0)
		for _, accountIndex := range accountIndexList {
			findAccountIndex := false
			for _, accountHistory := range accountHistories {
				if accountIndex == accountHistory.AccountIndex {
					findAccountIndex = true
					break
				}
			}
			if findAccountIndex == false {
				deleteAccountIndexes = append(deleteAccountIndexes, accountIndex)
			}
		}

		nftIndexSlice := make([]int64, 0)
		nftHistories := make([]*nft.L2NftHistory, 0)
		nftIndexLen := len(nftIndexList)
		for _, nftIndex := range nftIndexList {
			nftIndexLen--
			nftIndexSlice = append(nftIndexSlice, nftIndex)
			if len(nftIndexSlice) == 100 || nftIndexLen == 0 {
				_, nftHistoryList, err := bc.L2NftHistoryModel.GetLatestNftHistories(nftIndexSlice, height)
				if err != nil {
					logx.Error("get latest nft histories failed: ", err)
					panic("get latest nft histories failed: " + err.Error())
				}
				if nftHistoryList != nil {
					nftHistories = append(nftHistories, nftHistoryList...)
				}
				nftIndexSlice = make([]int64, 0)
			}
		}
		deleteNftIndexes := make([]int64, 0)
		for _, nftIndex := range nftIndexList {
			findNftIndex := false
			for _, nftHistory := range nftHistories {
				if nftIndex == nftHistory.NftIndex {
					findNftIndex = true
					break
				}
			}
			if findNftIndex == false {
				deleteNftIndexes = append(deleteNftIndexes, nftIndex)
			}
		}
		bc.DB().DB.Transaction(func(dbTx *gorm.DB) error {
			logx.Info("roll back account start")
			for _, accountHistory := range accountHistories {
				accountInfo := &account.Account{
					AccountIndex:    accountHistory.AccountIndex,
					Nonce:           accountHistory.Nonce,
					CollectionNonce: accountHistory.CollectionNonce,
					AssetInfo:       accountHistory.AssetInfo,
					AssetRoot:       accountHistory.AssetRoot,
					L2BlockHeight:   accountHistory.L2BlockHeight,
				}
				err := bc.AccountModel.UpdateByIndexInTransact(dbTx, accountInfo)
				if err != nil {
					logx.Error("roll back account failed: ", err)
					panic("roll back account failed: " + err.Error())
				}
			}
			logx.Info("roll back account,delete account start")
			for _, accountIndex := range deleteAccountIndexes {
				err := bc.AccountModel.DeleteByIndexInTransact(dbTx, accountIndex)
				if err != nil {
					logx.Error("roll back account,delete account failed: ", err)
					panic("roll back account,delete account failed: " + err.Error())
				}
			}

			logx.Info("roll back account history,delete account start")
			err := bc.AccountHistoryModel.DeleteByHeightInTransact(dbTx, heights)
			if err != nil {
				logx.Error("roll back account history,delete account history failed: ", err)
				panic("roll back account history,delete account history failed: " + err.Error())
			}

			logx.Info("roll back nft start")
			for _, nftHistory := range nftHistories {
				nftInfo := &nft.L2Nft{
					OwnerAccountIndex:   nftHistory.OwnerAccountIndex,
					NftContentHash:      nftHistory.NftContentHash,
					CollectionId:        nftHistory.CollectionId,
					CreatorTreasuryRate: nftHistory.CreatorTreasuryRate,
					CreatorAccountIndex: nftHistory.CreatorAccountIndex,
					L2BlockHeight:       nftHistory.L2BlockHeight,
					IpnsName:            nftHistory.IpnsName,
					IpnsId:              nftHistory.IpnsId,
					Metadata:            nftHistory.Metadata,
				}
				err := bc.L2NftModel.UpdateByIndexInTransact(dbTx, nftInfo)
				if err != nil {
					logx.Error("roll back nft failed: ", err)
					panic("roll back nft failed: " + err.Error())
				}
			}
			logx.Info("roll back nft,delete nft start")
			for _, nftIndex := range deleteNftIndexes {
				err := bc.L2NftModel.DeleteByIndexInTransact(dbTx, nftIndex)
				if err != nil {
					logx.Error("roll back nft,delete nft failed: ", err)
					panic("roll back nft,delete nft failed: " + err.Error())
				}
			}

			logx.Info("roll back l2nft history,delete l2nft history start")
			err = bc.L2NftHistoryModel.DeleteByHeightInTransact(dbTx, heights)
			if err != nil {
				logx.Error("roll back l2nft history,delete l2nft history failed: ", err)
				panic("roll back account l2nft,delete l2nft history failed: " + err.Error())
			}

			logx.Info("roll back tx detail start")
			err = bc.TxDetailModel.DeleteByHeightInTransact(dbTx, heights)
			if err != nil {
				logx.Error("roll back tx detail failed: ", err)
				panic("roll back tx detail failed: " + err.Error())
			}
			logx.Info("roll back tx start")
			err = bc.TxModel.DeleteByHeightInTransact(dbTx, heights)
			if err != nil {
				logx.Error("roll back tx failed: ", err)
				panic("roll back tx failed: " + err.Error())
			}
			logx.Info("roll back block start")
			var statuses = []int{block.StatusProposing, block.StatusPacked}
			err = bc.BlockModel.DeleteBlockInTransact(dbTx, statuses)
			if err != nil {
				logx.Error("roll back block failed: ", err)
				panic("roll back block failed: " + err.Error())
			}

			logx.Info("roll back compressed block start")
			err = bc.CompressedBlockModel.DeleteByHeightInTransact(dbTx, statuses)
			if err != nil {
				logx.Error("roll back compressed block failed: ", err)
				panic("roll back compressed block failed: " + err.Error())
			}

			logx.Info("roll back pool tx step 1 start")
			err = bc.TxPoolModel.UpdateTxsToPending(dbTx)
			if err != nil {
				logx.Error("roll back pool tx step 1 failed: ", err)
				panic("roll back pool tx step 1 failed: " + err.Error())
			}
			logx.Info("roll back pool tx step 2 start")
			err = bc.TxPoolModel.UpdateTxsToPendingByHeight(dbTx, packedHeights)
			if err != nil {
				logx.Error("roll back pool tx step 2 failed: ", err)
				panic("roll back pool tx step 2 failed: " + err.Error())
			}

			curHeight, err := bc.BlockModel.GetCurrentBlockHeightInTransact(dbTx)
			if err != nil {
				logx.Error("get current block height in transact failed: ", err)
				panic("get current block height in transact failed: " + err.Error())
			}
			poolTxId := uint(0)
			if curHeight != 0 {
				poolTxId, err = bc.TxModel.GetMaxPoolTxIdByHeightInTransact(dbTx, curHeight)
				if err != nil {
					logx.Error("get max pool tx id by height failed: ", err)
					panic("get max pool tx id by height failed: " + err.Error())
				}
			}
			logx.Info("roll back pool tx step 3 start")
			err = bc.TxPoolModel.UpdateTxsToPendingByMaxId(dbTx, poolTxId)
			if err != nil {
				logx.Error("roll back pool tx step 3 failed: ", err)
				panic("roll back pool tx step 3 failed: " + err.Error())
			}
			return nil
		})

		blocks, err = bc.BlockModel.GetBlockByStatus([]int{block.StatusProposing, block.StatusPacked})
		if err != nil {
			logx.Error("get proposing block height failed: ", err)
			panic("delete block failed: " + err.Error())
		}
		if blocks != nil {
			logx.Infof("get proposing block heights: %v", blocks)
			panic("delete block failed: " + err.Error())

		}
	}
}

func (bc *BlockChain) ApplyTransaction(tx *tx.Tx) error {
	return bc.processor.Process(tx)
}

func (bc *BlockChain) InitNewBlock() (*block.Block, error) {
	newBlock := &block.Block{
		Model: gorm.Model{
			// The block timestamp will be set when the first transaction executed.
			CreatedAt: time.Time{},
		},
		BlockHeight: bc.currentBlock.BlockHeight + 1,
		StateRoot:   bc.currentBlock.StateRoot,
		BlockStatus: block.StatusProposing,
	}

	bc.currentBlock = newBlock
	bc.Statedb.PurgeCache(bc.currentBlock.StateRoot)
	err := bc.Statedb.MarkGasAccountAsPending()
	return newBlock, err
}

func (bc *BlockChain) CurrentBlock() *block.Block {
	return bc.currentBlock
}

func (bc *BlockChain) UpdateAccountAssetTree(stateDataCopy *statedb.StateDataCopy) error {
	start := time.Now()
	// Intermediate state root.
	err := bc.Statedb.IntermediateRoot(false, stateDataCopy)
	if err != nil {
		return err
	}
	updateAccountTreeMetrics.Set(float64(time.Since(start).Milliseconds()))
	return nil
}

func (bc *BlockChain) UpdateAccountTreeAndNftTree(blockSize int, stateDataCopy *statedb.StateDataCopy) (*block.BlockStates, error) {
	newBlock := stateDataCopy.CurrentBlock
	err := bc.Statedb.AccountTreeAndNftTreeMultiSet(stateDataCopy)
	if err != nil {
		return nil, err
	}
	// Align pub data.
	bc.Statedb.AlignPubData(blockSize, stateDataCopy)

	commitment := chain.CreateBlockCommitment(newBlock.BlockHeight, newBlock.CreatedAt.UnixMilli(),
		common.FromHex(bc.Statedb.PreviousStateRoot), common.FromHex(stateDataCopy.StateCache.StateRoot),
		stateDataCopy.StateCache.PubData, int64(len(stateDataCopy.StateCache.PubDataOffset)))

	newBlock.BlockSize = uint16(blockSize)
	newBlock.BlockCommitment = commitment
	newBlock.StateRoot = stateDataCopy.StateCache.StateRoot
	newBlock.PriorityOperations = stateDataCopy.StateCache.PriorityOperations
	newBlock.PendingOnChainOperationsHash = common.Bytes2Hex(stateDataCopy.StateCache.PendingOnChainOperationsHash)
	newBlock.Txs = stateDataCopy.StateCache.Txs
	for _, executedTx := range newBlock.Txs {
		executedTx.TxStatus = tx.StatusPacked
	}
	if len(stateDataCopy.StateCache.PendingOnChainOperationsPubData) > 0 {
		onChainOperationsPubDataBytes, err := json.Marshal(stateDataCopy.StateCache.PendingOnChainOperationsPubData)
		if err != nil {
			return nil, fmt.Errorf("marshal pending onChain operation pubData failed: %v", err)
		}
		newBlock.PendingOnChainOperationsPubData = string(onChainOperationsPubDataBytes)
	}

	offsetBytes, err := json.Marshal(stateDataCopy.StateCache.PubDataOffset)
	if err != nil {
		return nil, fmt.Errorf("marshal pubData offset failed: %v", err)
	}
	newCompressedBlock := &compressedblock.CompressedBlock{
		BlockSize:         uint16(blockSize),
		BlockHeight:       newBlock.BlockHeight,
		StateRoot:         newBlock.StateRoot,
		PublicData:        common.Bytes2Hex(stateDataCopy.StateCache.PubData),
		Timestamp:         newBlock.CreatedAt.UnixMilli(),
		PublicDataOffsets: string(offsetBytes),
	}
	bc.Statedb.PreviousStateRoot = stateDataCopy.StateCache.StateRoot
	currentHeight := stateDataCopy.CurrentBlock.BlockHeight

	start := time.Now()
	logx.Infof("CommitAccountTreeAndNftTree,latestVersion=%d,prunedBlockHeight=%d", uint64(bc.Statedb.AccountTree.LatestVersion()), uint64(bc.StateDB().GetPrunedBlockHeight()))
	prunedVersion := bc.StateDB().GetPrunedBlockHeight()
	err = tree.CommitAccountTreeAndNftTree(uint64(prunedVersion), bc.Statedb.AccountTree, bc.Statedb.NftTree)
	if err != nil {
		return nil, err
	}
	commitAccountTreeMetrics.Set(float64(time.Since(start).Milliseconds()))

	pendingAccount, pendingAccountHistory, err := bc.Statedb.GetPendingAccount(currentHeight, stateDataCopy)
	if err != nil {
		return nil, err
	}

	pendingNft, pendingNftHistory, err := bc.Statedb.GetPendingNft(currentHeight, stateDataCopy)
	if err != nil {
		return nil, err
	}
	return &block.BlockStates{
		Block:                 newBlock,
		CompressedBlock:       newCompressedBlock,
		PendingAccount:        pendingAccount,
		PendingAccountHistory: pendingAccountHistory,
		PendingNft:            pendingNft,
		PendingNftHistory:     pendingNftHistory,
	}, nil
}

func (bc *BlockChain) VerifyExpiredAt(expiredAt int64) error {
	if !bc.dryRun {
		if expiredAt < bc.currentBlock.CreatedAt.UnixMilli() {
			return types.AppErrInvalidExpireTime
		}
	} else {
		if expiredAt < time.Now().UnixMilli() {
			return types.AppErrInvalidExpireTime
		}
	}
	return nil
}

func (bc *BlockChain) VerifyNonce(accountIndex int64, nonce int64) error {
	if !bc.dryRun {
		expectNonce, err := bc.Statedb.GetCommittedNonce(accountIndex)
		if err != nil {
			return err
		}
		logx.Infof("committer verify nonce start,accountIndex=%d,nonce=%d,expectNonce=%d", accountIndex, nonce, expectNonce)
		if nonce != expectNonce {
			logx.Infof("committer verify nonce failed,accountIndex=%d,nonce=%d,expectNonce=%d", accountIndex, nonce, expectNonce)
			bc.Statedb.SetPendingNonceToRedisCache(accountIndex, expectNonce-1)
			return types.AppErrInvalidNonce
		} else {
			logx.Infof("committer verify nonce success,accountIndex=%d,nonce=%d,expectNonce=%d", accountIndex, nonce, expectNonce)
		}
	} else {
		pendingNonce, err := bc.Statedb.GetPendingNonceFromCache(accountIndex)
		if err != nil {
			return err
		}
		if pendingNonce != nonce {
			logx.Infof("clear pending nonce from redis cache,accountIndex=%d,pendingNonce=%d,nonce=%d", accountIndex, pendingNonce, nonce)
			bc.Statedb.ClearPendingNonceFromRedisCache(accountIndex)
			return types.AppErrInvalidNonce
		}
	}
	return nil
}

func (bc *BlockChain) VerifyGas(gasAccountIndex, gasFeeAssetId int64, txType int, gasFeeAmount *big.Int, skipGasAmtChk bool) error {
	cfgGasAccountIndex, err := bc.Statedb.GetGasAccountIndex()
	if err != nil {
		return err
	}
	if gasAccountIndex != cfgGasAccountIndex {
		return types.AppErrInvalidGasFeeAccount
	}

	cfgGasFee, err := bc.Statedb.GetGasConfig()
	if err != nil {
		return err
	}

	gasAsset, ok := cfgGasFee[uint32(gasFeeAssetId)]
	if !ok {
		logx.Errorf("cannot find gas config for asset id: %d", gasFeeAssetId)
		return types.AppErrInvalidGasFeeAsset
	}

	if !skipGasAmtChk {
		gasFee, ok := gasAsset[txType]
		if !ok {
			return errors.New("invalid tx type")
		}
		if gasFeeAmount.Cmp(big.NewInt(gasFee)) < 0 {
			return types.AppErrInvalidGasFeeAmount
		}
	}
	return nil
}

func (bc *BlockChain) StateDB() *sdb.StateDB {
	return bc.Statedb
}

func (bc *BlockChain) DB() *sdb.ChainDB {
	return bc.ChainDB
}

func (bc *BlockChain) setCurrentBlockTimeStamp() {
	if bc.currentBlock.CreatedAt.IsZero() && len(bc.Statedb.Txs) == 0 {
		creatAt := time.Now().UnixMilli()
		bc.currentBlock.CreatedAt = time.UnixMilli(creatAt)
	}
}

func (bc *BlockChain) resetCurrentBlockTimeStamp() {
	if len(bc.Statedb.Txs) > 0 {
		return
	}

	bc.currentBlock.CreatedAt = time.Time{}
}
