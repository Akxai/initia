package accnum

import (
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmosante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
)

// AccountNumberDecorator is a custom ante handler that increments the account number depending on
// the execution mode (Simulate, CheckTx, Finalize).
//
// This is to avoid account number conflicts when running concurrent Simulate, CheckTx, and Finalize.
type AccountNumberDecorator struct {
	ak cosmosante.AccountKeeper
}

// NewAccountNumberDecorator creates a new instance of AccountNumberDecorator.
func NewAccountNumberDecorator(ak cosmosante.AccountKeeper) AccountNumberDecorator {
	return AccountNumberDecorator{ak}
}

// AnteHandle is the AnteHandler implementation for AccountNumberDecorator.
// AccountNumberDecorator.AnteHandle updated to use atomic increment
func (and AccountNumberDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if !ctx.IsCheckTx() && !ctx.IsReCheckTx() && !simulate {
		return next(ctx, tx, simulate)
	}

	// Safely cast to the concrete AccountKeeper type.
	ak, ok := and.ak.(*authkeeper.AccountKeeper)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "failed to cast AccountKeeper")
	}

	gasFreeCtx := ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
	accountNumAddition := uint64(1_000_000)
	if simulate {
		accountNumAddition += 1_000_000
	}

	// Perform an atomic increment of the account number.
	if err := ak.IncrementAccountNumber(gasFreeCtx, accountNumAddition); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}
