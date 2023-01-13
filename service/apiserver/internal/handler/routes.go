// Code generated by goctl. DO NOT EDIT.
package handler

import (
	"net/http"

	account "github.com/bnb-chain/zkbnb/service/apiserver/internal/handler/account"
	asset "github.com/bnb-chain/zkbnb/service/apiserver/internal/handler/asset"
	block "github.com/bnb-chain/zkbnb/service/apiserver/internal/handler/block"
	info "github.com/bnb-chain/zkbnb/service/apiserver/internal/handler/info"
	nft "github.com/bnb-chain/zkbnb/service/apiserver/internal/handler/nft"
	root "github.com/bnb-chain/zkbnb/service/apiserver/internal/handler/root"
	transaction "github.com/bnb-chain/zkbnb/service/apiserver/internal/handler/transaction"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/",
				Handler: root.GetStatusHandler(serverCtx),
			},
		},
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/accounts",
				Handler: account.GetAccountsHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/account",
				Handler: account.GetAccountHandler(serverCtx),
			},
		},
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/assets",
				Handler: asset.GetAssetsHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/asset",
				Handler: asset.GetAssetHandler(serverCtx),
			},
		},
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/blocks",
				Handler: block.GetBlocksHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/block",
				Handler: block.GetBlockHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/currentHeight",
				Handler: block.GetCurrentHeightHandler(serverCtx),
			},
		},
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/layer2BasicInfo",
				Handler: info.GetLayer2BasicInfoHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/gasFee",
				Handler: info.GetGasFeeHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/gasFeeAssets",
				Handler: info.GetGasFeeAssetsHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/gasAccount",
				Handler: info.GetGasAccountHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/search",
				Handler: info.SearchHandler(serverCtx),
			},
		},
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/txs",
				Handler: transaction.GetTxsHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/blockTxs",
				Handler: transaction.GetBlockTxsHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/accountTxs",
				Handler: transaction.GetAccountTxsHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/mergedAccountTxs",
				Handler: transaction.GetMergedAccountTxsHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/tx",
				Handler: transaction.GetTxHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/pendingTxs",
				Handler: transaction.GetPendingTxsHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/executedTxs",
				Handler: transaction.GetExecutedTxsHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/accountPendingTxs",
				Handler: transaction.GetAccountPendingTxsHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/nextNonce",
				Handler: transaction.GetNextNonceHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/sendTx",
				Handler: transaction.SendTxHandler(serverCtx),
			},
		},
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/maxOfferId",
				Handler: nft.GetMaxOfferIdHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/accountNfts",
				Handler: nft.GetAccountNftsHandler(serverCtx),
			},
		},
	)
}
