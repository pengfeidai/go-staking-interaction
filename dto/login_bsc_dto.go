package dto

type LoginBSCeResponse struct {
	Token string `json:"token"`
}
type LoginBSCRequest struct {
	Signature     string `json:"signature"`
	Nonce         string `json:"nonce"`
	Timestamp     int64  `json:"timestamp"`
	WalletAddress string `json:"address"`
}

type LoginSolanaRequest struct {
	Signature     string `json:"signature"`
	Nonce         string `json:"nonce"`
	Timestamp     int64  `json:"timestamp"`
	WalletAddress string `json:"address"`
}
