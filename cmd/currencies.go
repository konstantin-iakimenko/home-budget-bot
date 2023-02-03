package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/encoding/charmap"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	CurrenciesInsert = "INSERT INTO currencies (id, code, title) VALUES ($1, $2, $3)"
)

type ValCurs struct {
	xml.Name `xml:"ValCurs"`
	Valute   []Valute `xml:"Valute"`
}

type Valute struct {
	ID       string `xml:"ID,attr"`
	NumCode  string `xml:"NumCode"`
	CharCode string `xml:"CharCode"`
	Nominal  string `xml:"Nominal"`
	Name     string `xml:"Name"`
	Value    string `xml:"Value"`
}

type Currency struct {
	NumCode int64
	Code    string
	ExRate  *big.Float
	Symbol  string
}

type Currencies struct {
	Rsd Currency
	Eur Currency
	Usd Currency
	Try Currency
	Gbp Currency
	Rub Currency
}

func (c *Currencies) parseAmount(amount string) (int64, *Currency, error) {
	amount = strings.ToLower(amount)
	amount = strings.TrimSpace(amount)
	if strings.HasSuffix(amount, c.Eur.Symbol) {
		return c.Eur.getAmount(amount)
	} else if strings.HasSuffix(amount, c.Usd.Symbol) {
		return c.Usd.getAmount(amount)
	} else if strings.HasSuffix(amount, c.Try.Symbol) {
		return c.Try.getAmount(amount)
	} else if strings.HasSuffix(amount, c.Gbp.Symbol) {
		return c.Gbp.getAmount(amount)
	} else if strings.HasSuffix(amount, c.Rub.Symbol) {
		return c.Rub.getAmount(amount)
	} else {
		value, err := strconv.ParseInt(amount, 10, 64)
		if err != nil {
			return 0, nil, err
		}
		return value, &c.Rsd, nil
	}
}

func (c *Currency) getAmount(amount string) (int64, *Currency, error) {
	amount = strings.TrimSuffix(amount, c.Symbol)
	value, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		return 0, nil, err
	}
	return value, c, nil
}

func GetCurrencies(date time.Time) (*Currencies, error) {
	log.Info().Msgf("GetCurrencies: %s", date.Format("02/01/2006"))
	valCurs, err := getValCurs(date)
	if err != nil {
		log.Error().Err(err).Msgf("failed to get currencies: %w", err)
		return nil, err
	}

	var rsd, eur, usd, try, gbp, rub Currency
	for _, v := range valCurs.Valute {
		code, err := strconv.ParseInt(v.NumCode, 10, 64)
		if err != nil {
			log.Error().Err(err).Msgf("unable to parse num code: %w", err)
			return nil, err
		}
		exRate, err := v.getExRate()
		if err != nil {
			log.Error().Err(err).Msgf("unable to get ex rate: %w", err)
			return nil, err
		}

		if v.CharCode == "RSD" {
			rsd = Currency{
				NumCode: code,
				Code:    v.CharCode,
				ExRate:  exRate,
				Symbol:  "",
			}
		}
		if v.CharCode == "EUR" {
			eur = Currency{
				NumCode: code,
				Code:    v.CharCode,
				ExRate:  exRate,
				Symbol:  "€",
			}
		}
		if v.CharCode == "USD" {
			usd = Currency{
				NumCode: code,
				Code:    v.CharCode,
				ExRate:  exRate,
				Symbol:  "$",
			}
		}
		if v.CharCode == "TRY" {
			try = Currency{
				NumCode: code,
				Code:    v.CharCode,
				ExRate:  exRate,
				Symbol:  "tl",
			}
		}
		if v.CharCode == "GBP" {
			gbp = Currency{
				NumCode: code,
				Code:    v.CharCode,
				ExRate:  exRate,
				Symbol:  "£",
			}
		}
	}
	rub = Currency{
		NumCode: 643,
		Code:    "RUB",
		ExRate:  big.NewFloat(1),
		Symbol:  "₽",
	}

	return &Currencies{
		Rsd: rsd,
		Eur: eur,
		Usd: usd,
		Try: try,
		Gbp: gbp,
		Rub: rub,
	}, nil
}

func (v *Valute) getExRate() (*big.Float, error) {
	str := strings.Replace(v.Value, ",", ".", 1)
	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		log.Error().Err(err).Msgf("unable to parse value: %w", err)
		return big.NewFloat(0), err
	}
	nominal, err := strconv.ParseInt(v.Nominal, 10, 64)
	if err != nil {
		log.Error().Err(err).Msgf("unable to parse nominal: %w", err)
		return big.NewFloat(0), err
	}

	if nominal == 1 {
		return big.NewFloat(value), nil
	} else {
		return new(big.Float).Quo(big.NewFloat(value), big.NewFloat(float64(nominal))), nil
	}
}

/*
func main() {
	currencies, err := GetCurrencies(time.Now())
	if err != nil {
		panic(err)
	}

	testPrint("100", currencies)
	testPrint("200£", currencies)
	testPrint("300€", currencies)
	testPrint("400$", currencies)
	testPrint("500₽", currencies)
	testPrint("600tl", currencies)
	testPrint("700Tl", currencies)
	testPrint("800TL", currencies)
	// uploadCurrencies()
}
*/

//func testPrint(amount string, currencies *Currencies) {
//	val, cur, err := currencies.parseAmount(amount)
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println("=====================================")
//	str := fmt.Sprintf("amount %s | val %d | cur %s", amount, val, cur)
//	fmt.Println(str)
//
//}

func uploadCurrencies() {
	valCurs, err := getValCurs(time.Now())
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to get currencies: %w", err)
	}
	saveCurriencies(valCurs)
}

func getValCurs(date time.Time) (*ValCurs, error) {
	url := fmt.Sprintf("http://www.cbr.ru/scripts/XML_daily.asp?date_req=%s", date.Format("02/01/2006"))

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	d := xml.NewDecoder(bytes.NewReader(data))
	d.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		switch charset {
		case "windows-1251":
			return charmap.Windows1251.NewDecoder().Reader(input), nil
		default:
			return nil, fmt.Errorf("unknown charset: %s", charset)
		}
	}
	var result ValCurs
	err = d.Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func saveCurriencies(valCurs *ValCurs) {
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(os.Getenv("PG_HOMEBUDGET_DB")) // DatabaseURL
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to parse conn string (%s): %w", os.Getenv("PG_HOMEBUDGET_DB"), err)
	}
	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		log.Fatal().Err(err).Msgf("unable to connect to database: %w", err)
	}
	defer pool.Close()

	for _, valute := range valCurs.Valute {
		log.Info().Msgf("valute: %s", valute)
		code, err := strconv.ParseInt(valute.NumCode, 10, 64)
		if err != nil {
			log.Fatal().Err(err).Msgf("unable to parse num code: %w", err)
		}
		_, err = pool.Exec(ctx, CurrenciesInsert, code, valute.CharCode, valute.Name)
		if err != nil {
			log.Fatal().Err(err).Msgf("unable to insert currency: %w", err)
		}
	}
	// insert into currencies(id, code, title, format) values (643, 'RUB', 'Российский рубль', '%s руб.');
}
