package nft

import (
	"context"
	"errors"
	"github.com/bnb-chain/zkbnb/dao/nft"
	types2 "github.com/bnb-chain/zkbnb/types"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateNftByIndexLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateNftByIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateNftByIndexLogic {
	return &UpdateNftByIndexLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateNftByIndexLogic) UpdateNftByIndex(req *types.ReqUpdateNft) (resp *types.History, err error) {
	l2Nft, err := l.svcCtx.NftModel.GetNft(req.NftIndex)
	if err != nil {
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrNftNotFound
		}
		return nil, types2.AppErrInternal
	}
	if l2Nft.IpfsStatus == nft.NotConfirmed {
		return nil, errors.New("please wait for data synchronization to complete")
	}
	history := &nft.L2NftMetadataHistory{
		NftIndex: req.NftIndex,
		IpnsName: l2Nft.IpnsName,
		IpnsId:   l2Nft.IpnsId,
		Mutable:  req.MutableAttributes,
		Status:   nft.NotConfirmed,
	}
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		err = l.svcCtx.NftMetadataHistoryModel.DeleteL2NftMetadataHistoryInTransact(tx, req.NftIndex)
		if err != nil {
			return err
		}
		err = l.svcCtx.NftMetadataHistoryModel.CreateL2NftMetadataHistoryInTransact(tx, history)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &types.History{
		IpnsId: l2Nft.IpnsId,
	}, nil
}
