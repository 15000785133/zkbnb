// Code generated by goctl. DO NOT EDIT.
package types

type ReqGetStatus struct {
}

type RespGetStatus struct {
	Status    uint32 `json:"status"`
	NetworkId uint32 `json:"network_id"`
}

type ReqGetLayer2BasicInfo struct {
}

type RespGetLayer2BasicInfo struct {
	BlockCommitted         int64    `json:"block_committed"`
	BlockExecuted          int64    `json:"block_executed"`
	TotalTransactionsCount int64    `json:"total_transactions_count"`
	ContractAddresses      []string `json:"contract_addresses"`
}

type ReqGetAssetsList struct {
}

type Asset struct {
	AssetId       int64  `json:"asset_id"`
	AssetAddr     string `json:"asset_addr"`
	AssetDecimals int64  `json:"asset_decimals"`
	AssetSymbol   string `json:"asset_symbol"`
}

type RespGetAssetsList struct {
	Assets []*Asset `json:"assets"`
}

type ReqSearch struct {
	Info string `form:"info"`
}

type RespSearch struct {
	DataType int32 `json:"data_type"`
}

type ReqGetAccounts struct {
	Offset uint16 `form:"offset" validator:"min=0"`
	Limit  uint16 `form:"limit" validator:"min=0,max=50"`
}

type Accounts struct {
	AccountIndex uint32 `json:"account_index"`
	AccountName  string `json:"account_name"`
	PublicKey    string `json:"public_key"`
}

type RespGetAccounts struct {
	Total    uint32      `json:"total"`
	Accounts []*Accounts `json:"accounts"`
}

type TxHash struct {
	TxHash    string `json:"tx_hash"`
	CreatedAt int64  `json:"created_at"`
}

type Block struct {
	BlockHeight     int32     `json:"block_height"`
	BlockStatus     int32     `json:"block_status"`
	NewAccountRoot  string    `json:"new_account_root"`
	CommittedAt     int64     `json:"committed_at"`
	VerifiedAt      int64     `json:"verified_at"`
	ExecutedAt      int64     `json:"executed_at"`
	CommittedTxHash []*TxHash `json:"committed_tx_hash"`
	VerifiedTxHash  []*TxHash `json:"verified_tx_hash"`
	ExecutedTxHash  []*TxHash `json:"executed_tx_hash"`
	BlockCommitment string    `json:"block_commitment"`
	TxCount         int64     `json:"tx_count"`
	Txs             []string  `json:"txs"`
}

type ReqGetBlocks struct {
	Offset uint16 `form:"offset"`
	Limit  uint16 `form:"limit"`
}

type RespGetBlocks struct {
	Total  uint32   `json:"total"`
	Blocks []*Block `json:"blocks"`
}

type ReqGetBlockByCommitment struct {
	BlockCommitment string `form:"block_commitment"`
}

type RespGetBlockByCommitment struct {
	Blocks []*Block `json:"blocks"`
}

type ReqGetBlockByBlockHeight struct {
	BlockHeight uint64 `form:"block_height"`
}

type RespGetBlockByBlockHeight struct {
	Block Block `json:"block"`
}

type Tx struct {
	TxHash         string      `json:"tx_hash"`
	TxType         int32       `json:"tx_type"`
	GasFee         int32       `json:"gas_fee"`
	GasFeeAssetId  int32       `json:"gas_fee_asset_id"`
	TxStatus       int32       `json:"tx_status"`
	BlockHeight    int64       `json:"block_height"`
	BlockStatus    int32       `json:"block_status"`
	BlockId        int32       `json:"block_id"`
	AssetAId       int32       `json:"asseta_id"`
	AssetBId       int32       `json:"assetb_id"`
	TxAmount       int64       `json:"tx_amount"`
	TxParticipants []string    `json:"tx_participants"`
	NativeAddress  string      `json:"native_address"`
	CreatedAt      int64       `json:"created_at"`
	TxAssetAId     int32       `json:"tx_asseta_id"`
	TxAssetBId     int32       `json:"tx_assetb_id"`
	TxDetails      []*TxDetail `json:"tx_detail"`
	CommittedAt    int64       `json:"committed_at"`
	VerifiedAt     int64       `json:"verified_at"`
	ExecutedAt     int64       `json:"executed_at"`
	Memo           string      `json:"memo"`
}

type TxDetail struct {
	AssetId        int    `json:"asset_id"`
	AssetType      int    `json:"asset_type"`
	AccountIndex   int32  `json:"account_index"`
	AccountName    string `json:"account_name"`
	AccountBalance string `json:"account_balance"`
	AccountDelta   string `json:"account_delta"`
}

type ReqGetTxsListByBlockHeight struct {
	BlockHeight uint64 `form:"block_height"`
	Limit       uint16 `form:"limit"`
	Offset      uint16 `form:"offset"`
}

type RespGetTxsListByBlockHeight struct {
	Total uint32 `json:"total"`
	Txs   []*Tx  `json:"txs"`
}

type ReqGetTxByHash struct {
	TxHash string `form:"tx_hash"`
}

type RespGetTxByHash struct {
	Txs Tx `json:"result"`
}

type ReqGetTxsListByAccountIndex struct {
	AccountIndex uint32 `form:"account_index"`
	Offset       uint16 `form:"offset"`
	Limit        uint16 `form:"limit"`
}

type RespGetTxsListByAccountIndex struct {
	Total uint32 `json:"total"`
	Txs   []*Tx  `json:"txs"`
}

type ReqGetMempoolTxsList struct {
	Offset uint16 `form:"offset"`
	Limit  uint16 `form:"limit"`
}

type RespGetMempoolTxsList struct {
	Total uint32 `json:"total"`
	Txs   []*Tx  `json:"txs"`
}

type ReqGetMempoolTxsListByPublicKey struct {
	AccountPk string `form:"account_pk"`
	Offset    uint16 `form:"offset"`
	Limit     uint16 `form:"limit"`
}

type RespGetMempoolTxsListByPublicKey struct {
	Total uint32 `json:"total"`
	Txs   []*Tx  `json:"data"`
}

type AssetInfo struct {
	AssetId uint32 `json:"assetId"`
	Balance string `json:"balance"`
}

type AccountInfo struct {
	AccountIndex uint32       `json:"account_index"`
	AccountName  string       `json:"account_name"`
	AccountPk    string       `json:"account_pk"`
	Assets       []*AssetInfo `json:"assets"`
}

type ReqGetAccountInfoByAccountName struct {
	AccountName string `form:"account_name"`
}

type RespGetAccountInfoByAccountName struct {
	Account AccountInfo `json:"account"`
}
