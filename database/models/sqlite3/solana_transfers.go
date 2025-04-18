package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/thrasher-corp/sqlboiler/boil"
	"github.com/thrasher-corp/sqlboiler/queries"
	"github.com/thrasher-corp/sqlboiler/queries/qm"
	"github.com/thrasher-corp/sqlboiler/strmangle"
)

// SolanaTransfer is an object representing the database table.
type SolanaTransfer struct {
	Signature        string         `boil:"signature" json:"signature" toml:"signature" yaml:"signature"`
	Network          string         `boil:"network" json:"network" toml:"network" yaml:"network"`
	SendTime         sql.NullTime   `boil:"send_time" json:"send_time" toml:"send_time" yaml:"send_time"`
	Sender           sql.NullString `boil:"sender" json:"sender" toml:"sender" yaml:"sender"`
	Receiver         sql.NullString `boil:"receiver" json:"receiver" toml:"receiver" yaml:"receiver"`
	IsTokenTransfer  bool           `boil:"is_token_transfer" json:"is_token_transfer" toml:"is_token_transfer" yaml:"is_token_transfer"`
	AmountDisplay    float64        `boil:"amount_display" json:"amount_display" toml:"amount_display" yaml:"amount_display"`
	TokenMintAddress sql.NullString `boil:"token_mint_address" json:"token_mint_address" toml:"token_mint_address" yaml:"token_mint_address"`

	R *solanaTransferR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L solanaTransferL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *SolanaTransfer) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("sqlite3: no solana_transfers provided for insertion")
	}

	// 定义列信息
	solanaTransferAllColumns := []string{"signature", "network", "send_time", "sender", "receiver", "is_token_transfer", "amount_display", "token_mint_address"}
	solanaTransferColumnsWithDefault := []string{}
	solanaTransferColumnsWithoutDefault := []string{"signature", "network", "send_time", "sender", "receiver", "is_token_transfer", "amount_display", "token_mint_address"}

	// 获取要插入的列
	wl, _ := columns.InsertColumnSet(
		solanaTransferAllColumns,
		solanaTransferColumnsWithDefault,
		solanaTransferColumnsWithoutDefault,
		queries.NonZeroDefaultSet(solanaTransferColumnsWithDefault, o),
	)

	// 构建SQL查询
	var query string
	if len(wl) != 0 {
		query = fmt.Sprintf("INSERT INTO \"solana_transfers\" (\"%s\") VALUES (%s)",
			strings.Join(wl, "\",\""),
			strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
	} else {
		query = "INSERT INTO \"solana_transfers\" () VALUES ()"
	}

	// 直接从结构体获取值
	solanaTransferType := reflect.TypeOf(SolanaTransfer{})
	value := reflect.Indirect(reflect.ValueOf(o))

	// 创建值映射
	vals := make([]interface{}, len(wl))
	for i, colName := range wl {
		field, _ := solanaTransferType.FieldByNameFunc(func(fieldName string) bool {
			field, _ := solanaTransferType.FieldByName(fieldName)
			tag := field.Tag.Get("boil")
			return tag == colName
		})
		vals[i] = value.FieldByIndex(field.Index).Interface()
	}

	// 执行插入
	_, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to insert into solana_transfers")
	}

	return nil
}

// solanaTransferR is where relationships are stored.
type solanaTransferR struct {
}

// NewStruct creates a new relationship struct
func (*solanaTransferR) NewStruct() *solanaTransferR {
	return &solanaTransferR{}
}

// solanaTransferL is where Load methods for each relationship are stored.
type solanaTransferL struct{}

// SolanaTransferQuery is used to build up a query for SolanaTransfer records
type solanaTransferQuery struct {
	*queries.Query
}

// SolanaTransferSlice is an alias for a slice of pointers to SolanaTransfer
type SolanaTransferSlice []*SolanaTransfer

// SolanaTransfers retrieves all the records using an executor
func SolanaTransfers(mods ...qm.QueryMod) solanaTransferQuery {
	mods = append(mods, qm.From("\"solana_transfers\""))
	return solanaTransferQuery{NewQuery(mods...)}
}

// FindSolanaTransfer retrieves a single record by Signature with an executor.
// If selectCols is empty Find will return all columns.
func FindSolanaTransfer(ctx context.Context, exec boil.ContextExecutor, signature string, selectCols ...string) (*SolanaTransfer, error) {
	solanaTransferObj := &SolanaTransfer{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"solana_transfers\" where \"signature\"=$1", sel,
	)

	q := queries.Raw(query, signature)

	err := q.Bind(ctx, exec, solanaTransferObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到签名为 %s 的转账记录", signature)
		}
		return nil, fmt.Errorf("查询转账记录失败: %v", err)
	}

	return solanaTransferObj, nil
}

// FindSolanaTransfersByNetwork retrieves records by Network with an executor.
func FindSolanaTransfersByNetwork(ctx context.Context, exec boil.ContextExecutor, network string, selectCols ...string) (SolanaTransferSlice, error) {
	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"solana_transfers\" where \"network\"=$1", sel,
	)

	q := queries.Raw(query, network)

	var result SolanaTransferSlice
	err := q.Bind(ctx, exec, &result)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到网络为 %s 的转账记录", network)
		}
		return nil, fmt.Errorf("查询转账记录失败: %v", err)
	}

	return result, nil
}

// All returns all SolanaTransfer records from the query.
func (q solanaTransferQuery) All(ctx context.Context, exec boil.ContextExecutor) (SolanaTransferSlice, error) {
	var o SolanaTransferSlice

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, fmt.Errorf("查询所有转账记录失败: %v", err)
	}

	return o, nil
}

// One returns a single SolanaTransfer record from the query.
func (q solanaTransferQuery) One(ctx context.Context, exec boil.ContextExecutor) (*SolanaTransfer, error) {
	o := &SolanaTransfer{}

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到符合条件的转账记录")
		}
		return nil, fmt.Errorf("查询单条转账记录失败: %v", err)
	}

	return o, nil
}
