package witness

import (
	"context"
	"encoding/json"
	"fmt"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/log"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/plugin/dbresolver"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb-crypto/circuit"
	bsmt "github.com/bnb-chain/zkbnb-smt"
	smt "github.com/bnb-chain/zkbnb-smt"
	utils "github.com/bnb-chain/zkbnb/common/prove"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/blockwitness"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/dao/proof"
	"github.com/bnb-chain/zkbnb/service/witness/config"
	"github.com/bnb-chain/zkbnb/tree"
	"github.com/bnb-chain/zkbnb/types"
)

const (
	UnprovedBlockWitnessTimeout = 10 * time.Minute

	BlockProcessDelta = 10
)

var (
	l2BlockWitnessGenerateHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2Block_witness_generate_height",
		Help:      "l2Block_memory_height metrics.",
	})
	AccountLatestVersionTreeMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "witness_account_latest_version",
		Help:      "Account latest version metrics.",
	})
	AccountRecentVersionTreeMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "witness_account_recent_version",
		Help:      "Account recent version metrics.",
	})
	NftTreeLatestVersionMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "witness_nft_latest_version",
		Help:      "Nft latest version metrics.",
	})
	NftTreeRecentVersionMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "witness_nft_recent_version",
		Help:      "Nft recent version metrics.",
	})
)

var (
	l2WitnessHeightMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2_witness_height",
		Help:      "l2_witness_height metrics.",
	})
)

type Witness struct {
	// config
	config config.Config
	helper *utils.WitnessHelper

	// Trees
	treeCtx     *tree.Context
	accountTree smt.SparseMerkleTree
	assetTrees  *tree.AssetTreeCache
	nftTree     smt.SparseMerkleTree

	// The data access object
	db                  *gorm.DB
	blockModel          block.BlockModel
	accountModel        account.AccountModel
	accountHistoryModel account.AccountHistoryModel
	nftHistoryModel     nft.L2NftHistoryModel
	nftModel            nft.L2NftModel
	proofModel          proof.ProofModel
	blockWitnessModel   blockwitness.BlockWitnessModel
}

func NewWitness(c config.Config, shouldCheckStateRoot bool) (*Witness, error) {

	if err := prometheus.Register(l2BlockWitnessGenerateHeightMetric); err != nil {
		return nil, fmt.Errorf("prometheus.Register l2BlockWitnessGenerateHeightMetric error: %v", err)
	}
	if err := prometheus.Register(AccountLatestVersionTreeMetric); err != nil {
		return nil, fmt.Errorf("prometheus.Register AccountLatestVersionTreeMetric error: %v", err)
	}
	if err := prometheus.Register(AccountRecentVersionTreeMetric); err != nil {
		return nil, fmt.Errorf("prometheus.Register AccountRecentVersionTreeMetric error: %v", err)
	}
	if err := prometheus.Register(NftTreeLatestVersionMetric); err != nil {
		return nil, fmt.Errorf("prometheus.Register NftTreeLatestVersionMetric error: %v", err)
	}
	if err := prometheus.Register(NftTreeRecentVersionMetric); err != nil {
		return nil, fmt.Errorf("prometheus.Register NftTreeRecentVersionMetric error: %v", err)
	}

	masterDataSource := c.Postgres.MasterDataSource
	slaveDataSource := c.Postgres.SlaveDataSource
	db, err := gorm.Open(postgres.Open(masterDataSource))
	if err := prometheus.Register(l2WitnessHeightMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register l2WitnessHeightMetrics error: %v", err)
	}

	db.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{postgres.Open(masterDataSource)},
		Replicas: []gorm.Dialector{postgres.Open(slaveDataSource)},
	}))

	w := &Witness{
		config:              c,
		db:                  db,
		blockModel:          block.NewBlockModel(db),
		blockWitnessModel:   blockwitness.NewBlockWitnessModel(db),
		accountModel:        account.NewAccountModel(db),
		accountHistoryModel: account.NewAccountHistoryModel(db),
		nftHistoryModel:     nft.NewL2NftHistoryModel(db),
		nftModel:            nft.NewL2NftModel(db),
		proofModel:          proof.NewProofModel(db),
	}
	err = w.initState(shouldCheckStateRoot)
	return w, err
}

func (w *Witness) initState(shouldCheckStateRoot bool) error {
	witnessHeight, err := w.blockWitnessModel.GetLatestBlockWitnessHeight()
	if err != nil {
		if err != types.DbErrNotFound {
			return fmt.Errorf("GetLatestBlockWitness error: %v", err)
		}

		witnessHeight = 0
	}

	// dbinitializer tree database
	treeCtx, err := tree.NewContext("witness", w.config.TreeDB.Driver, false, false, w.config.TreeDB.RoutinePoolSize, &w.config.TreeDB.LevelDBOption, &w.config.TreeDB.RedisDBOption)
	if err != nil {
		return err
	}

	treeCtx.SetOptions(bsmt.BatchSizeLimit(3 * 1024 * 1024))
	err = tree.SetupTreeDB(treeCtx)
	if err != nil {
		return fmt.Errorf("init tree database failed %v", err)
	}
	w.treeCtx = treeCtx
	blockInfo, err := w.blockModel.GetBlockByHeightWithoutTx(witnessHeight + 1)
	if err != nil && err != types.DbErrNotFound {
		logx.Error("get block failed: ", err)
		panic("get block failed: " + err.Error())
	}
	accountIndexes := make([]int64, 0)
	if blockInfo != nil && blockInfo.AccountIndexes != "[]" && blockInfo.AccountIndexes != "" {
		err = json.Unmarshal([]byte(blockInfo.AccountIndexes), &accountIndexes)
		if err != nil {
			logx.Error("json err unmarshal failed")
			panic("json err unmarshal failed: " + err.Error())
		}
	}
	w.accountTree, w.assetTrees, err = tree.InitAccountTree(
		w.accountModel,
		w.accountHistoryModel,
		accountIndexes,
		witnessHeight,
		treeCtx,
		w.config.TreeDB.AssetTreeCacheSize,
		true,
	)
	// the blockHeight depends on the proof start position
	if err != nil {
		return fmt.Errorf("initMerkleTree error: %v", err)
	}

	w.nftTree, err = tree.InitNftTree(w.nftModel, w.nftHistoryModel, witnessHeight,
		treeCtx, true)
	if err != nil {
		return fmt.Errorf("initNftTree error: %v", err)
	}

	if shouldCheckStateRoot {
		err = w.checkStateRoot(witnessHeight)
		if err != nil {
			return err
		}
	}
	w.helper = utils.NewWitnessHelper(w.treeCtx, w.accountTree, w.nftTree, w.assetTrees, w.accountModel, w.accountHistoryModel)
	return nil
}

func (w *Witness) GenerateBlockWitness() (err error) {
	var latestWitnessHeight int64
	latestWitnessHeight, err = w.blockWitnessModel.GetLatestBlockWitnessHeight()
	if err != nil && err != types.DbErrNotFound {
		return err
	}
	// get next batch of blocks
	blocks, err := w.blockModel.GetPendingBlocksBetween(latestWitnessHeight+1, latestWitnessHeight+BlockProcessDelta)
	if err != nil {
		if err != types.DbErrNotFound {
			return err
		}
		return nil
	}
	// get latestVerifiedBlockNr
	latestVerifiedBlockNr, err := w.blockModel.GetLatestVerifiedHeight()
	if err != nil {
		return err
	}

	// scan each block
	for _, block := range blocks {
		ctx := log.NewCtxWithKV(log.BlockHeightContext, block.BlockHeight)
		logx.WithContext(ctx).Infof("construct witness for block %d", block.BlockHeight)
		// Step1: construct witness
		blockWitness, err := w.constructBlockWitness(block, latestVerifiedBlockNr, ctx)
		if err != nil {
			return fmt.Errorf("failed to construct block witness, block:%d, err: %v", block.BlockHeight, err)
		}
		// Step2: commit trees for witness
		err = tree.CommitTrees(uint64(latestVerifiedBlockNr), block.BlockHeight, w.accountTree, w.assetTrees, w.nftTree)
		if err != nil {
			return fmt.Errorf("unable to commit trees after txs is executed, block:%d, error: %v", block.BlockHeight, err)
		}
		// Step3: insert witness into database
		err = w.blockWitnessModel.CreateBlockWitness(blockWitness)
		l2BlockWitnessGenerateHeightMetric.Set(float64(latestVerifiedBlockNr))
		AccountLatestVersionTreeMetric.Set(float64(w.accountTree.LatestVersion()))
		AccountRecentVersionTreeMetric.Set(float64(w.accountTree.RecentVersion()))
		NftTreeLatestVersionMetric.Set(float64(w.nftTree.LatestVersion()))
		NftTreeRecentVersionMetric.Set(float64(w.nftTree.RecentVersion()))
		l2WitnessHeightMetrics.Set(float64(blockWitness.Height))
		if err != nil {
			// rollback trees
			rollBackErr := tree.RollBackTrees(uint64(block.BlockHeight)-1, w.accountTree, w.assetTrees, w.nftTree)
			if rollBackErr != nil {
				logx.WithContext(ctx).Errorf("unable to rollback trees %v", rollBackErr)
			}
			return fmt.Errorf("create unproved crypto block error, block:%d, err: %v", block.BlockHeight, err)
		}
		w.assetTrees.CleanChanges()

	}
	return nil
}

func (w *Witness) RescheduleBlockWitness() {
	nextBlockNumber, err := w.getNextWitnessToCheck()
	ctx := log.NewCtxWithKV(log.BlockHeightContext, nextBlockNumber)
	if err != nil {
		logx.WithContext(ctx).Errorf("failed to get next witness to check, err: %s", err.Error())
	}
	nextBlockWitness, err := w.blockWitnessModel.GetBlockWitnessByHeight(nextBlockNumber)
	if err != nil && err != types.DbErrNotFound {
		logx.WithContext(ctx).Errorf("failed to get latest block witness, err: %s", err.Error())
		return
	}

	if nextBlockWitness == nil {
		return
	}

	// skip if next block is not processed
	if nextBlockWitness.Status == blockwitness.StatusPublished {
		return
	}

	// skip if the next block proof exists
	// if the proof is not submitted and verified in L1, there should be another alerts
	_, err = w.proofModel.GetProofByBlockHeight(nextBlockNumber)
	if err == nil {
		return
	}

	// update block status to Published if it's timeout
	if time.Now().After(nextBlockWitness.UpdatedAt.Add(UnprovedBlockWitnessTimeout)) {
		logx.WithContext(ctx).Infof("reschedule block %d", nextBlockWitness.Height)
		err := w.blockWitnessModel.UpdateBlockWitnessStatus(nextBlockWitness, blockwitness.StatusPublished)
		if err != nil {
			logx.WithContext(ctx).Errorf("update unproved block status error, err: %s", err.Error())
			return
		}
	}
}

func (w *Witness) getNextWitnessToCheck() (int64, error) {
	latestProof, err := w.proofModel.GetLatestProof()
	if err != nil && err != types.DbErrNotFound {
		logx.Errorf("failed to get latest proof, err: %s", err.Error())
		return 0, err
	}

	if err == types.DbErrNotFound {
		return 1, nil
	}

	latestConfirmedProof, err := w.proofModel.GetLatestConfirmedProof()
	if err != nil && err != types.DbErrNotFound {
		logx.Errorf("failed to get latest confirmed proof, err: %s", err.Error())
		return 0, err
	}

	var startToCheck, endToCheck int64 = 1, latestProof.BlockNumber
	if err != types.DbErrNotFound {
		startToCheck = latestConfirmedProof.BlockNumber + 1
	}

	for blockHeight := startToCheck; blockHeight < endToCheck; blockHeight++ {
		_, err = w.proofModel.GetProofByBlockHeight(blockHeight)
		if err != nil {
			return blockHeight, nil
		}
	}
	return endToCheck + 1, nil
}

func (w *Witness) constructBlockWitness(block *block.Block, latestVerifiedBlockNr int64, ctx context.Context) (*blockwitness.BlockWitness, error) {
	var oldStateRoot, newStateRoot []byte
	txsWitness := make([]*utils.TxWitness, 0, block.BlockSize)
	// scan each transaction
	err := w.helper.ResetCache(block.BlockHeight)
	if err != nil {
		return nil, err
	}
	for idx, tx := range block.Txs {
		txWitness, err := w.helper.ConstructTxWitness(tx, uint64(latestVerifiedBlockNr), ctx)
		if err != nil {
			return nil, err
		}
		txsWitness = append(txsWitness, txWitness)
		// if it is the first tx of the block
		if idx == 0 {
			oldStateRoot = txWitness.StateRootBefore
		}
		// if it is the last tx of the block
		if idx == len(block.Txs)-1 {
			newStateRoot = txWitness.StateRootAfter
		}
	}

	emptyTxCount := int(block.BlockSize) - len(block.Txs)
	for i := 0; i < emptyTxCount; i++ {
		txsWitness = append(txsWitness, circuit.EmptyTx(newStateRoot))
	}

	gasWitness, err := w.helper.ConstructGasWitness(block)
	if err != nil {
		return nil, err
	}

	accountTreeRoot := w.accountTree.Root()
	nftTreeRoot := w.nftTree.Root()
	logx.WithContext(ctx).Infof("witness account tree root=%s,nft tree root=%s", common.Bytes2Hex(accountTreeRoot), common.Bytes2Hex(nftTreeRoot))
	newStateRoot = tree.ComputeStateRootHash(accountTreeRoot, nftTreeRoot)
	newStateRootStr := common.Bytes2Hex(newStateRoot)
	if newStateRootStr != block.StateRoot {
		return nil, types.AppErrStateRootNotMatch
	}

	b := &circuit.Block{
		BlockNumber:     block.BlockHeight,
		CreatedAt:       block.CreatedAt.UnixMilli(),
		OldStateRoot:    oldStateRoot,
		NewStateRoot:    newStateRoot,
		BlockCommitment: common.FromHex(block.BlockCommitment),
		Txs:             txsWitness,
		Gas:             gasWitness,
	}
	bz, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	blockWitness := blockwitness.BlockWitness{
		Height:      block.BlockHeight,
		WitnessData: string(bz),
		BlockSize:   block.BlockSize,
		Status:      blockwitness.StatusPublished,
	}
	return &blockWitness, nil
}

func (w *Witness) Shutdown() {
	sqlDB, err := w.db.DB()
	if err == nil && sqlDB != nil {
		err = sqlDB.Close()
	}
	if err != nil {
		logx.Errorf("close db error: %s", err.Error())
	}

	err = w.treeCtx.TreeDB.Close()
	if err != nil {
		logx.Errorf("close treedb error: %s", err.Error())
	}
}

func (w *Witness) Rollback(height int64) (err error) {
	ctxLog := log.NewCtxWithKV(log.BlockHeightContext, height)
	curHeight, err := w.blockModel.GetCurrentBlockHeight()
	if err != nil {
		return fmt.Errorf("get current block height failed: %s", err.Error())
	}
	logx.WithContext(ctxLog).Infof("get current block height: %d", curHeight)

	latestVerifiedBlockNr, err := w.blockModel.GetLatestVerifiedHeight()
	if err != nil {
		return fmt.Errorf("get latest verified height failed: %v", err)
	}
	if height <= latestVerifiedBlockNr {
		return fmt.Errorf("failed to rollback,height must be more than %d", latestVerifiedBlockNr)
	}

	blocks, err := w.blockModel.GetBlockByHeights(height, curHeight)
	if err != nil {
		return fmt.Errorf("failed to GetBlockByHeights,from height=%d,to height=%d,err %v", height, curHeight, err)
	}

	accountIndexMap := make(map[int64]bool, 0)
	for _, blockInfo := range blocks {
		accountIndexes := make([]int64, 0)
		if blockInfo != nil && blockInfo.AccountIndexes != "[]" && blockInfo.AccountIndexes != "" {
			err = json.Unmarshal([]byte(blockInfo.AccountIndexes), &accountIndexes)
			if err != nil {
				return fmt.Errorf("json err unmarshal failed,height=%d", blockInfo.BlockHeight)
			}
		}

		for _, accountIndex := range accountIndexes {
			accountIndexMap[accountIndex] = true
		}
	}

	treeHeight := height - 1
	for accountIndex := range accountIndexMap {
		asset := w.assetTrees.Get(accountIndex)
		assetRoot := common.Bytes2Hex(asset.Root())
		logx.WithContext(ctxLog).Infof("start to rollback asset tree, accountIndex:%d, latestVersion:%d,versions=%s,assetRoot:%s,rollback to height:%d", accountIndex, asset.LatestVersion(), common2.FormatVersion(asset.Versions()), assetRoot, treeHeight)
		if asset.LatestVersion() > bsmt.Version(treeHeight) && !asset.IsEmpty() {
			logx.WithContext(ctxLog).Infof("asset tree accountIndex:%d latestVersion:%d is higher than block, rollback to height:%d", accountIndex, asset.LatestVersion(), treeHeight)
			err := asset.Rollback(bsmt.Version(treeHeight))
			if err != nil {
				return fmt.Errorf("unable to rollback asset accountIndex:%d, latestVersion: %d,tree err: %s", accountIndex, asset.LatestVersion(), err.Error())
			}
			if asset.LatestVersion() > bsmt.Version(treeHeight) {
				return fmt.Errorf("call asset.Rollback successfully,but fail to rollback asset accountIndex:%d latestVersion: %d,versions=%s", accountIndex, asset.LatestVersion(), common2.FormatVersion(asset.Versions()))
			}
		}
	}
	logx.WithContext(ctxLog).Infof("start to rollback account tree, latestVersion:%d,versions=%s,accountRoot:%s,rollback to height:%d", w.accountTree.LatestVersion(), common2.FormatVersion(w.accountTree.Versions()), common.Bytes2Hex(w.accountTree.Root()), treeHeight)
	if w.accountTree.LatestVersion() > bsmt.Version(treeHeight) && !w.accountTree.IsEmpty() {
		logx.WithContext(ctxLog).Infof("account tree latestVersion:%d is higher than block, rollback to %d", w.accountTree.LatestVersion(), treeHeight)
		err := w.accountTree.Rollback(bsmt.Version(treeHeight))
		if err != nil {
			if err != bsmt.ErrVersionTooOld {
				return fmt.Errorf("unable to rollback account latestVersion:%d,err:%s, ", treeHeight, err.Error())
			}

		}
		if w.accountTree.LatestVersion() > bsmt.Version(treeHeight) {
			return fmt.Errorf("call accountTree.Rollback successfully,but fail to rollback accountTree,latestVersion: %d,versions=%s", w.accountTree.LatestVersion(), common2.FormatVersion(w.accountTree.Versions()))
		}
	}

	logx.WithContext(ctxLog).Infof("start to rollback nft tree, latestVersion:%d,versions=%s,nftRoot:%s,rollback to height:%d", w.accountTree.LatestVersion(), common2.FormatVersion(w.accountTree.Versions()), common.Bytes2Hex(w.nftTree.Root()), treeHeight)
	if w.nftTree.LatestVersion() > bsmt.Version(treeHeight) && !w.nftTree.IsEmpty() {
		logx.WithContext(ctxLog).Infof("nft tree latestVersion:%d is higher than block, rollback to %d", w.nftTree.LatestVersion(), treeHeight)
		err := w.nftTree.Rollback(bsmt.Version(treeHeight))
		if err != nil {
			return fmt.Errorf("unable to rollback nft latestVersion:%d,err:%s", treeHeight, err.Error())
		}
		if w.nftTree.LatestVersion() > bsmt.Version(treeHeight) {
			return fmt.Errorf("call nftTree.Rollback successfully,but fail to rollback nftTree,latestVersion: %d,versions=%s", w.nftTree.LatestVersion(), common2.FormatVersion(w.nftTree.Versions()))
		}
	}

	err = tree.CheckAssetRoot(accountIndexMap, treeHeight, w.assetTrees, w.accountHistoryModel)
	if err != nil {
		return err
	}

	err = w.checkStateRoot(height - 1)
	if err != nil {
		return err
	}

	return nil
}

func (w *Witness) checkStateRoot(height int64) (err error) {
	accountTreeRoot := w.accountTree.Root()
	nftTreeRoot := w.nftTree.Root()
	logx.Infof("witness account tree root=%s,nft tree root=%s", common.Bytes2Hex(accountTreeRoot), common.Bytes2Hex(nftTreeRoot))
	newStateRoot := tree.ComputeStateRootHash(accountTreeRoot, nftTreeRoot)
	newStateRootStr := common.Bytes2Hex(newStateRoot)
	blockInfo, err := w.blockModel.GetBlockByHeightWithoutTx(height)
	if err != nil {
		return fmt.Errorf("failed to get block info by height=%d error=%v", height, err)
	}

	if newStateRootStr != blockInfo.StateRoot {
		return fmt.Errorf("state root not matched between witness and committer: %d, witness: %s, committer: %s", height, newStateRootStr, blockInfo.StateRoot)
	}
	return nil
}
