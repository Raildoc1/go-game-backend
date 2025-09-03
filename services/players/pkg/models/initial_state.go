package models

// WalletState represents player's wallet currencies.
type WalletState struct {
	Soft int64 `json:"soft"`
	Hard int64 `json:"hard"`
}

// InitialState aggregates various player-related feature states.
type InitialState struct {
	Nickname string      `json:"nickname"`
	Wallet   WalletState `json:"wallet"`
}
