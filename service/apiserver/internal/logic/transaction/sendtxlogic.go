package transaction

import (
	"context"
	"github.com/bnb-chain/zkbnb/core"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/signature"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type SendTxLogic struct {
	logx.Logger
	ctx              context.Context
	svcCtx           *svc.ServiceContext
	l1AddressFetcher *signature.L1AddressFetcher
}

func NewSendTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendTxLogic {

	l1AddressFetcher := signature.NewL1AddressFetcher(ctx, svcCtx)
	return &SendTxLogic{
		Logger:           logx.WithContext(ctx),
		ctx:              ctx,
		svcCtx:           svcCtx,
		l1AddressFetcher: l1AddressFetcher,
	}
}

func (s *SendTxLogic) SendTx(req *types.ReqSendTx) (resp *types.TxHash, err error) {
	txStatuses := []int64{tx.StatusPending}
	pendingTxCount, err := s.svcCtx.TxPoolModel.GetTxsTotalCount(tx.GetTxWithStatuses(txStatuses))
	if err != nil {
		return nil, types2.AppErrInternal
	}

	if s.svcCtx.Config.TxPool.MaxPendingTxCount > 0 && pendingTxCount >= int64(s.svcCtx.Config.TxPool.MaxPendingTxCount) {
		return nil, types2.AppErrTooManyTxs
	}

	err = s.verifySignature(req.TxType, req.TxInfo, req.TxSignature)
	if err != nil {
		return nil, err
	}

	resp = &types.TxHash{}
	bc, err := core.NewBlockChainForDryRun(s.svcCtx.AccountModel, s.svcCtx.NftModel, s.svcCtx.TxPoolModel,
		s.svcCtx.AssetModel, s.svcCtx.SysConfigModel, s.svcCtx.RedisCache)
	if err != nil {
		logx.Error("fail to init blockchain runner:", err)
		return nil, types2.AppErrInternal
	}
	newTx := &tx.Tx{
		TxHash: types2.EmptyTxHash, // Would be computed in prepare method of executors.
		TxType: int64(req.TxType),
		TxInfo: req.TxInfo,

		GasFeeAssetId: types2.NilAssetId,
		GasFee:        types2.NilAssetAmount,
		NftIndex:      types2.NilNftIndex,
		CollectionId:  types2.NilCollectionNonce,
		AssetId:       types2.NilAssetId,
		TxAmount:      types2.NilAssetAmount,
		NativeAddress: types2.EmptyL1Address,

		BlockHeight: types2.NilBlockHeight,
		TxStatus:    tx.StatusPending,
	}

	err = bc.ApplyTransaction(newTx)
	if err != nil {
		return resp, err
	}
	if err := s.svcCtx.TxPoolModel.CreateTxs([]*tx.Tx{newTx}); err != nil {
		logx.Errorf("fail to create pool tx: %v, err: %s", newTx, err.Error())
		return resp, types2.AppErrInternal
	}

	resp.TxHash = newTx.TxHash
	return resp, nil
}

func (s *SendTxLogic) verifySignature(TxType uint32, TxInfo, Signature string) error {
	signBody, err := signature.GenerateSignatureBody(TxType, TxInfo)
	if err != nil {
		return err
	}
	content := accounts.TextHash([]byte(signBody))
	sigPublicKey, err := crypto.SigToPub(content, []byte(Signature))
	if err != nil {
		return err
	}

	publicAddress := crypto.PubkeyToAddress(*sigPublicKey)
	originAddressStr, err := s.l1AddressFetcher.GetL1AddressByTx(TxType, TxInfo)
	if err != nil {
		return err
	}

	originAddress := common.HexToAddress(originAddressStr)
	if publicAddress != originAddress {
		return errors.New("Tx Signature Error")
	}
	return nil
}
