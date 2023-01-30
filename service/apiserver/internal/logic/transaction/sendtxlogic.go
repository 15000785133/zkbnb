package transaction

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	nftModels "github.com/bnb-chain/zkbnb/core/model"
	"github.com/bnb-chain/zkbnb/dao/dbcache"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/signature"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/core"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
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
	pendingTxCount, err := s.svcCtx.MemCache.GetTxPendingCountKeyPrefix(func() (interface{}, error) {
		txStatuses := []int64{tx.StatusPending}
		return s.svcCtx.TxPoolModel.GetTxsTotalCount(tx.GetTxWithStatuses(txStatuses))
	})
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
		s.svcCtx.AssetModel, s.svcCtx.SysConfigModel, s.svcCtx.RedisCache, s.svcCtx.MemCache.GetCache())
	if err != nil {
		logx.Error("fail to init blockchain runner:", err)
		return nil, types2.AppErrInternal
	}
	newPoolTx := tx.BaseTx{
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
	newTx := &tx.Tx{BaseTx: newPoolTx}
	err = bc.ApplyTransaction(newTx)
	if err != nil {
		return resp, err
	}
	newTx.BaseTx.TxType = int64(req.TxType)
	newTx.BaseTx.TxInfo = req.TxInfo
	newTx.BaseTx.BlockHeight = types2.NilBlockHeight
	newTx.BaseTx.TxStatus = tx.StatusPending
	if newTx.BaseTx.TxType == types2.TxTypeMintNft {
		newTx.BaseTx.NftIndex = types2.NilNftIndex
	}
	if newTx.BaseTx.TxType == types2.TxTypeCreateCollection {
		newTx.BaseTx.CollectionId = types2.NilCollectionNonce
	}
	if newTx.BaseTx.TxType == types2.TxTypeMintNft {
		txInfo, _ := types2.ParseMintNftTxInfo(req.TxInfo)
		content, err := sendToIpfs(txInfo, newTx.BaseTx.TxHash)
		if err != nil {
			return resp, err
		}
		var mutable string
		if txInfo.MetaData == nil {
			mutable = ""
		} else {
			mutable = txInfo.MetaData.MutableAttribute
		}
		history := &nft.L2NftMetadataHistory{
			TxHash:   newTx.BaseTx.TxHash,
			NftIndex: newTx.BaseTx.NftIndex,
			IpnsName: txInfo.IpnsName,
			IpnsId:   txInfo.IpnsId,
			Mutable:  mutable,
			Metadata: content,
			Status:   nft.NotConfirmed,
		}
		b, err := json.Marshal(txInfo)
		if err != nil {
			return resp, err
		}
		newTx.BaseTx.TxInfo = string(b)
		err = s.svcCtx.DB.Transaction(func(db *gorm.DB) error {
			err = s.svcCtx.NftMetadataHistoryModel.CreateL2NftMetadataHistoryInTransact(db, history)
			if err != nil {
				return err
			}
			err = s.svcCtx.TxPoolModel.CreateTxsInTransact(db, []*tx.PoolTx{{BaseTx: newTx.BaseTx}})
			return err
		})
		if err != nil {
			logx.Errorf("fail to create pool tx: %v, err: %s", newTx, err.Error())
			return resp, types2.AppErrInternal
		}
	} else {
		if err := s.svcCtx.TxPoolModel.CreateTxs([]*tx.PoolTx{{BaseTx: newTx.BaseTx}}); err != nil {
			logx.Errorf("fail to create pool tx: %v, err: %s", newTx, err.Error())
			return resp, types2.AppErrInternal
		}
	}
	s.svcCtx.RedisCache.Set(context.Background(), dbcache.AccountNonceKeyByIndex(newTx.AccountIndex), newTx.Nonce)
	resp.TxHash = newTx.TxHash
	return resp, nil
}

func (s *SendTxLogic) verifySignature(TxType uint32, TxInfo, Signature string) error {

	// For compatibility consideration, if signature string is empty, directly ignore the validation
	if len(Signature) == 0 {
		return nil
	}

	//Generate the signature body data from the transaction type and transaction info
	signatureBody, err := signature.GenerateSignatureBody(TxType, TxInfo)
	if err != nil {
		return err
	}
	message := accounts.TextHash([]byte(signatureBody))

	//Decode from signature string to get the signature byte array
	signatureContent, err := hexutil.Decode(Signature)
	if err != nil {
		return err
	}
	signatureContent[64] -= 27 // Transform yellow paper V from 27/28 to 0/1

	//Calculate the public key from the signature and source string
	signaturePublicKey, err := crypto.SigToPub(message, signatureContent)
	if err != nil {
		return err
	}

	//Calculate the address from the public key
	publicAddress := crypto.PubkeyToAddress(*signaturePublicKey)

	//Query the origin address from the database
	originAddressStr, err := s.l1AddressFetcher.GetL1AddressByTx(TxType, TxInfo)
	if err != nil {
		return err
	}
	originAddress := common.HexToAddress(originAddressStr)

	//Compare the original address and the public address to verify the identifier
	if publicAddress != originAddress {
		return errors.New("Tx Signature Error")
	}
	return nil
}

func sendToIpfs(txInfo *txtypes.MintNftTxInfo, txHash string) (string, error) {
	ipnsName := txHash
	ipnsId, err := common2.Ipfs.GenerateIPNS(ipnsName)
	if err != nil {
		return "", err
	}
	var cid string
	var content string
	if txInfo.MetaData == nil {
		cid, content, err = uploadIpfs(&nftModels.NftMetaData{
			Image:             "",
			Name:              "",
			Description:       "",
			Attributes:        "",
			MutableAttributes: fmt.Sprintf("%s%s", "https://ipfs.io/ipns/", ipnsId.Id),
		})
		if err != nil {
			return "", err
		}
	} else {
		cid, content, err = uploadIpfs(&nftModels.NftMetaData{
			Image:             txInfo.MetaData.Image,
			Name:              txInfo.MetaData.Name,
			Description:       txInfo.MetaData.Description,
			Attributes:        txInfo.MetaData.Attributes,
			MutableAttributes: fmt.Sprintf("%s%s", "https://ipfs.io/ipns/", ipnsId.Id),
		})
		if err != nil {
			return "", err
		}
	}
	hash, err := common2.Ipfs.GenerateHash(cid)
	if err != nil {
		return "", err
	}
	txInfo.NftContentHash = hash
	txInfo.IpnsName = ipnsName
	txInfo.IpnsId = ipnsId.Id
	return content, nil
}

func uploadIpfs(data *nftModels.NftMetaData) (string, string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", "", err
	}
	content := string(b)
	cid, err := common2.Ipfs.Upload(content)

	if err != nil {
		return "", "", err
	}
	return cid, content, nil
}
