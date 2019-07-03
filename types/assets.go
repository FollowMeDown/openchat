package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ChatDenom = "chat"			// 1 chat
	MicroChatDenom = "uchat"	// 1 uchat = 1e-6 chat
)

var (
	ChatUnit = sdk.OneDec()
	MicroChatUnit = sdk.NewDecWithPrec(1, 6)
)
