// Code generated by goctl. DO NOT EDIT!
// Source: globalRPC.proto

package globalrpc

import (
	"context"

	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	AssetResult                            = globalRPCProto.AssetResult
	ReqGetLatestAccountInfoByAccountIndex  = globalRPCProto.ReqGetLatestAccountInfoByAccountIndex
	ReqGetLatestAssetsListByAccountIndex   = globalRPCProto.ReqGetLatestAssetsListByAccountIndex
	ReqGetLatestPairInfo                   = globalRPCProto.ReqGetLatestPairInfo
	ReqGetLpValue                          = globalRPCProto.ReqGetLpValue
	ReqGetMaxOfferId                       = globalRPCProto.ReqGetMaxOfferId
	ReqGetNextNonce                        = globalRPCProto.ReqGetNextNonce
	ReqGetSwapAmount                       = globalRPCProto.ReqGetSwapAmount
	ReqSendCreateCollectionTx              = globalRPCProto.ReqSendCreateCollectionTx
	ReqSendMintNftTx                       = globalRPCProto.ReqSendMintNftTx
	ReqSendTx                              = globalRPCProto.ReqSendTx
	ReqSendTxByRawInfo                     = globalRPCProto.ReqSendTxByRawInfo
	RespGetLatestAccountInfoByAccountIndex = globalRPCProto.RespGetLatestAccountInfoByAccountIndex
	RespGetLatestAssetsListByAccountIndex  = globalRPCProto.RespGetLatestAssetsListByAccountIndex
	RespGetLatestPairInfo                  = globalRPCProto.RespGetLatestPairInfo
	RespGetLpValue                         = globalRPCProto.RespGetLpValue
	RespGetMaxOfferId                      = globalRPCProto.RespGetMaxOfferId
	RespGetNextNonce                       = globalRPCProto.RespGetNextNonce
	RespGetSwapAmount                      = globalRPCProto.RespGetSwapAmount
	RespSendCreateCollectionTx             = globalRPCProto.RespSendCreateCollectionTx
	RespSendMintNftTx                      = globalRPCProto.RespSendMintNftTx
	RespSendTx                             = globalRPCProto.RespSendTx
	TxDetailInfo                           = globalRPCProto.TxDetailInfo
	TxInfo                                 = globalRPCProto.TxInfo

	GlobalRPC interface {
		GetLatestAssetsListByAccountIndex(ctx context.Context, in *ReqGetLatestAssetsListByAccountIndex, opts ...grpc.CallOption) (*RespGetLatestAssetsListByAccountIndex, error)
		GetLatestAccountInfoByAccountIndex(ctx context.Context, in *ReqGetLatestAccountInfoByAccountIndex, opts ...grpc.CallOption) (*RespGetLatestAccountInfoByAccountIndex, error)
		GetLatestPairInfo(ctx context.Context, in *ReqGetLatestPairInfo, opts ...grpc.CallOption) (*RespGetLatestPairInfo, error)
		GetSwapAmount(ctx context.Context, in *ReqGetSwapAmount, opts ...grpc.CallOption) (*RespGetSwapAmount, error)
		GetLpValue(ctx context.Context, in *ReqGetLpValue, opts ...grpc.CallOption) (*RespGetLpValue, error)
		SendTx(ctx context.Context, in *ReqSendTx, opts ...grpc.CallOption) (*RespSendTx, error)
		SendCreateCollectionTx(ctx context.Context, in *ReqSendCreateCollectionTx, opts ...grpc.CallOption) (*RespSendCreateCollectionTx, error)
		SendMintNftTx(ctx context.Context, in *ReqSendMintNftTx, opts ...grpc.CallOption) (*RespSendMintNftTx, error)
		GetNextNonce(ctx context.Context, in *ReqGetNextNonce, opts ...grpc.CallOption) (*RespGetNextNonce, error)
		GetMaxOfferId(ctx context.Context, in *ReqGetMaxOfferId, opts ...grpc.CallOption) (*RespGetMaxOfferId, error)
		SendAddLiquidityTx(ctx context.Context, in *ReqSendTxByRawInfo, opts ...grpc.CallOption) (*RespSendTx, error)
		SendAtomicMatchTx(ctx context.Context, in *ReqSendTxByRawInfo, opts ...grpc.CallOption) (*RespSendTx, error)
		SendCancelOfferTx(ctx context.Context, in *ReqSendTxByRawInfo, opts ...grpc.CallOption) (*RespSendTx, error)
		SendRemoveLiquidityTx(ctx context.Context, in *ReqSendTxByRawInfo, opts ...grpc.CallOption) (*RespSendTx, error)
		SendSwapTx(ctx context.Context, in *ReqSendTxByRawInfo, opts ...grpc.CallOption) (*RespSendTx, error)
		SendTransferNftTx(ctx context.Context, in *ReqSendTxByRawInfo, opts ...grpc.CallOption) (*RespSendTx, error)
		SendTransferTx(ctx context.Context, in *ReqSendTxByRawInfo, opts ...grpc.CallOption) (*RespSendTx, error)
		SendWithdrawNftTx(ctx context.Context, in *ReqSendTxByRawInfo, opts ...grpc.CallOption) (*RespSendTx, error)
		SendWithdrawTx(ctx context.Context, in *ReqSendTxByRawInfo, opts ...grpc.CallOption) (*RespSendTx, error)
	}

	defaultGlobalRPC struct {
		cli zrpc.Client
	}
)

func NewGlobalRPC(cli zrpc.Client) GlobalRPC {
	return &defaultGlobalRPC{
		cli: cli,
	}
}

func (m *defaultGlobalRPC) GetLatestAssetsListByAccountIndex(ctx context.Context, in *ReqGetLatestAssetsListByAccountIndex, opts ...grpc.CallOption) (*RespGetLatestAssetsListByAccountIndex, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.GetLatestAssetsListByAccountIndex(ctx, in, opts...)
}

func (m *defaultGlobalRPC) GetLatestAccountInfoByAccountIndex(ctx context.Context, in *ReqGetLatestAccountInfoByAccountIndex, opts ...grpc.CallOption) (*RespGetLatestAccountInfoByAccountIndex, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.GetLatestAccountInfoByAccountIndex(ctx, in, opts...)
}

func (m *defaultGlobalRPC) GetLatestPairInfo(ctx context.Context, in *ReqGetLatestPairInfo, opts ...grpc.CallOption) (*RespGetLatestPairInfo, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.GetLatestPairInfo(ctx, in, opts...)
}

func (m *defaultGlobalRPC) GetSwapAmount(ctx context.Context, in *ReqGetSwapAmount, opts ...grpc.CallOption) (*RespGetSwapAmount, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.GetSwapAmount(ctx, in, opts...)
}

func (m *defaultGlobalRPC) GetLpValue(ctx context.Context, in *ReqGetLpValue, opts ...grpc.CallOption) (*RespGetLpValue, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.GetLpValue(ctx, in, opts...)
}

func (m *defaultGlobalRPC) SendTx(ctx context.Context, in *ReqSendTx, opts ...grpc.CallOption) (*RespSendTx, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.SendTx(ctx, in, opts...)
}

func (m *defaultGlobalRPC) SendCreateCollectionTx(ctx context.Context, in *ReqSendCreateCollectionTx, opts ...grpc.CallOption) (*RespSendCreateCollectionTx, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.SendCreateCollectionTx(ctx, in, opts...)
}

func (m *defaultGlobalRPC) SendMintNftTx(ctx context.Context, in *ReqSendMintNftTx, opts ...grpc.CallOption) (*RespSendMintNftTx, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.SendMintNftTx(ctx, in, opts...)
}

func (m *defaultGlobalRPC) GetNextNonce(ctx context.Context, in *ReqGetNextNonce, opts ...grpc.CallOption) (*RespGetNextNonce, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.GetNextNonce(ctx, in, opts...)
}

func (m *defaultGlobalRPC) GetMaxOfferId(ctx context.Context, in *ReqGetMaxOfferId, opts ...grpc.CallOption) (*RespGetMaxOfferId, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.GetMaxOfferId(ctx, in, opts...)
}

func (m *defaultGlobalRPC) SendAddLiquidityTx(ctx context.Context, in *ReqSendTxByRawInfo, opts ...grpc.CallOption) (*RespSendTx, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.SendAddLiquidityTx(ctx, in, opts...)
}

func (m *defaultGlobalRPC) SendAtomicMatchTx(ctx context.Context, in *ReqSendTxByRawInfo, opts ...grpc.CallOption) (*RespSendTx, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.SendAtomicMatchTx(ctx, in, opts...)
}

func (m *defaultGlobalRPC) SendCancelOfferTx(ctx context.Context, in *ReqSendTxByRawInfo, opts ...grpc.CallOption) (*RespSendTx, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.SendCancelOfferTx(ctx, in, opts...)
}

func (m *defaultGlobalRPC) SendRemoveLiquidityTx(ctx context.Context, in *ReqSendTxByRawInfo, opts ...grpc.CallOption) (*RespSendTx, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.SendRemoveLiquidityTx(ctx, in, opts...)
}

func (m *defaultGlobalRPC) SendSwapTx(ctx context.Context, in *ReqSendTxByRawInfo, opts ...grpc.CallOption) (*RespSendTx, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.SendSwapTx(ctx, in, opts...)
}

func (m *defaultGlobalRPC) SendTransferNftTx(ctx context.Context, in *ReqSendTxByRawInfo, opts ...grpc.CallOption) (*RespSendTx, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.SendTransferNftTx(ctx, in, opts...)
}

func (m *defaultGlobalRPC) SendTransferTx(ctx context.Context, in *ReqSendTxByRawInfo, opts ...grpc.CallOption) (*RespSendTx, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.SendTransferTx(ctx, in, opts...)
}

func (m *defaultGlobalRPC) SendWithdrawNftTx(ctx context.Context, in *ReqSendTxByRawInfo, opts ...grpc.CallOption) (*RespSendTx, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.SendWithdrawNftTx(ctx, in, opts...)
}

func (m *defaultGlobalRPC) SendWithdrawTx(ctx context.Context, in *ReqSendTxByRawInfo, opts ...grpc.CallOption) (*RespSendTx, error) {
	client := globalRPCProto.NewGlobalRPCClient(m.cli.Conn())
	return client.SendWithdrawTx(ctx, in, opts...)
}
