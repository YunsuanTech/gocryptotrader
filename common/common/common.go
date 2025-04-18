package common

import (
	"errors"
	"fmt"
	"github.com/urfave/cli"
	"math"
	"math/big"
	"os"

	"github.com/shopspring/decimal"
)

func FatalExit(err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	os.Exit(1)
}

var NotFoundErr = errors.New("not found")

var (
	weiPerGO   = big.NewInt(1e18)
	weiPerGwei = big.NewInt(1e9)
)

// Base converts b base units to wei (*1e18).
func Base(b int64) *big.Int {
	i := big.NewInt(b)
	return i.Mul(i, weiPerGO)
}

// Gwei converts g gwei to wei (*1e9).
func Gwei(g int64) *big.Int {
	i := big.NewInt(g)
	return i.Mul(i, weiPerGwei)
}

// WeiAsBase converts w wei in to the base unit, and formats it as a decimal fraction with full precision (up to 18 decimals).
func WeiAsBase(w *big.Int) string {
	return new(big.Rat).SetFrac(w, weiPerGO).FloatString(18)
}

// WeiAsGwei converts w wei in to gwei, and formats it as a decimal fraction with full precision (up to 9 decimals).
func WeiAsGwei(w *big.Int) string {
	return new(big.Rat).SetFrac(w, weiPerGwei).FloatString(9)
}

// IntAsFloat converts a *big.Int (ie: wei), to *big.Float (ie: ETH)
func IntAsFloat(i *big.Int, decimals int) *big.Float {
	f := new(big.Float)
	f.SetPrec(100)
	f.SetInt(i)
	f.Quo(f, big.NewFloat(math.Pow10(decimals)))
	return f
}

// DecToInt converts a decimal to a big int
func DecToInt(d decimal.Decimal, decimals int32) *big.Int {
	// multiply amount by number of decimals
	d1 := decimal.New(1, decimals)
	d = d.Mul(d1)
	return d.BigInt()
}

// IntToDec converts a big int to a decimal
func IntToDec(i *big.Int, decimals int32) decimal.Decimal {
	d := decimal.NewFromBigInt(i, 0)
	d = d.Div(decimal.New(1, decimals))
	return d
}

// FloatAsInt converts a float to a *big.Int based on the decimals passed in
func FloatAsInt(amountF *big.Float, decimals int) *big.Int {
	bigval := new(big.Float)
	bigval.SetPrec(100)
	bigval.SetString(amountF.String()) // have to do this to not lose precision

	coinDecimals := new(big.Float)
	coinDecimals.SetFloat64(math.Pow10(decimals))
	bigval.Mul(bigval, coinDecimals)

	amountI := new(big.Int)
	// todo: could sanity check the accuracy here
	bigval.Int(amountI) // big.NewInt(int64(amountInWeiF)) // amountInGo.Mul(amountInGo, big.NewInt(int64(math.Pow10(18))))
	return amountI
}

func ParseGasPriceAndLimit(c *cli.Context) (*big.Int, uint64) {
	gasLimit := c.Uint64("gas-limit")
	gp := c.String("gas-price")
	var price *big.Int
	var ok bool
	if gp != "" {
		price, ok = new(big.Int).SetString(gp, 10)
		if !ok {
			FatalExit(fmt.Errorf("invalid price %v", gp))
		}
	}
	gp = c.String("gas-price-gwei")
	if gp != "" {
		price, ok = new(big.Int).SetString(gp, 10)
		if !ok {
			FatalExit(fmt.Errorf("invalid price %v", gp))
		}
		price = Gwei(price.Int64())
	}
	return price, gasLimit
}

// ToWei converts decimals to wei this is copied from the
// utils package to prevent import cycles
func ToWei(iamount interface{}, decimals int) *big.Int {
	amount := decimal.NewFromFloat(0)
	switch v := iamount.(type) {
	case string:
		amount, _ = decimal.NewFromString(v)
	case float64:
		amount = decimal.NewFromFloat(v)
	case int64:
		amount = decimal.NewFromFloat(float64(v))
	case decimal.Decimal:
		amount = v
	case *decimal.Decimal:
		amount = *v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	result := amount.Mul(mul)

	wei := new(big.Int)
	wei.SetString(result.String(), 10)

	return wei
}
