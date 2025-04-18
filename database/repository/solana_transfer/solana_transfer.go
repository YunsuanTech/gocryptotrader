package solana_transfer

import (
	"context"
	"database/sql"
	"time"

	"gocryptotrader/database"
	modelSQLite "gocryptotrader/database/models/sqlite3"

	"github.com/thrasher-corp/sqlboiler/boil"
	"github.com/thrasher-corp/sqlboiler/queries/qm"
)

// GetSolanaTransfer returns list of solana transfers matching query
func GetSolanaTransfer(network string, limit int) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	var mods []qm.QueryMod
	if network != "" {
		mods = append(mods, qm.Where("network = ?", network))
	}

	mods = append(mods, qm.OrderBy("send_time DESC"))
	if limit > 0 {
		mods = append(mods, qm.Limit(limit))
	}

	ctx := context.TODO()
	return modelSQLite.SolanaTransfers(mods...).All(ctx, database.DB.SQL)
}

// GetSolanaTransferBySignature returns a specific solana transfer by signature
func GetSolanaTransferBySignature(signature string) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	ctx := context.TODO()
	return modelSQLite.FindSolanaTransfer(ctx, database.DB.SQL, signature)
}

// GetSolanaTransfersByNetwork returns all solana transfers for a specific network
func GetSolanaTransfersByNetwork(network string) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	ctx := context.TODO()
	return modelSQLite.FindSolanaTransfersByNetwork(ctx, database.DB.SQL, network)
}

// GetSolanaTransfersBySender returns all solana transfers from a specific sender
func GetSolanaTransfersBySender(sender string, limit int) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	var mods []qm.QueryMod
	mods = append(mods, qm.Where("sender = ?", sender))
	mods = append(mods, qm.OrderBy("send_time DESC"))
	if limit > 0 {
		mods = append(mods, qm.Limit(limit))
	}

	ctx := context.TODO()
	return modelSQLite.SolanaTransfers(mods...).All(ctx, database.DB.SQL)
}

// GetSolanaTransfersByReceiver returns all solana transfers to a specific receiver
func GetSolanaTransfersByReceiver(receiver string, limit int) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	var mods []qm.QueryMod
	mods = append(mods, qm.Where("receiver = ?", receiver))
	mods = append(mods, qm.OrderBy("send_time DESC"))
	if limit > 0 {
		mods = append(mods, qm.Limit(limit))
	}

	ctx := context.TODO()
	return modelSQLite.SolanaTransfers(mods...).All(ctx, database.DB.SQL)
}

// RecordSOLTransfer records a native SOL transfer in the database
func RecordSOLTransfer(signature, network string, sendTime time.Time, sender, receiver string, amount float64) error {
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// Create a new SolanaTransfer object
	transfer := &modelSQLite.SolanaTransfer{
		Signature:        signature,
		Network:          network,
		SendTime:         sql.NullTime{Time: sendTime, Valid: true},
		Sender:           sql.NullString{String: sender, Valid: sender != ""},
		Receiver:         sql.NullString{String: receiver, Valid: receiver != ""},
		IsTokenTransfer:  false,
		AmountDisplay:    amount,
		TokenMintAddress: sql.NullString{Valid: false}, // 确保TokenMintAddress为NULL
	}

	// Insert the record into the database
	ctx := context.TODO()
	return transfer.Insert(ctx, database.DB.SQL, boil.Infer())
}

// RecordTokenTransfer records a token transfer in the database
func RecordTokenTransfer(signature, network string, sendTime time.Time, sender, receiver string, amount float64, tokenMintAddress string) error {
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// Create a new SolanaTransfer object
	transfer := &modelSQLite.SolanaTransfer{
		Signature:        signature,
		Network:          network,
		SendTime:         sql.NullTime{Time: sendTime, Valid: true},
		Sender:           sql.NullString{String: sender, Valid: sender != ""},
		Receiver:         sql.NullString{String: receiver, Valid: receiver != ""},
		IsTokenTransfer:  true,
		AmountDisplay:    amount,
		TokenMintAddress: sql.NullString{String: tokenMintAddress, Valid: tokenMintAddress != ""},
	}

	// Insert the record into the database
	ctx := context.TODO()
	return transfer.Insert(ctx, database.DB.SQL, boil.Infer())
}
