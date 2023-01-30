package nftModels

type NftMetaData struct {
	Image             string `json:"image"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Attributes        string `json:"attributes"`
	MutableAttributes string `json:"mutable_attributes"`
}
