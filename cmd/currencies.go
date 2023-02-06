package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
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

func (c *Currency) getAmount(amount string) (int64, *Currency, error) {
	amount = strings.TrimSuffix(amount, c.Symbol)
	value, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		return 0, nil, err
	}
	return value, c, nil
}

func InitCurCash() *CurCash {
	curMap := map[string](map[string]Currency){}
	return &CurCash{m: curMap}
}

func (c *CurCash) Get(date time.Time, code string) (*Currency, error) {
	if code == "RUB" {
		return &Currency{
			NumCode: 643,
			Code:    "RUB",
			ExRate:  big.NewFloat(1),
			Symbol:  "₽",
		}, nil
	}

	dateName := date.Format("2006-01-02")
	var result *Currency

	if _, ok := c.m[dateName]; ok {
		currency := c.m[dateName][code]
		result = &currency
	} else {
		if _, err := os.Stat(fmt.Sprintf("cmd/filecache/%s.xml", dateName)); err != nil && os.IsNotExist(err) {
			err = createCurFile(date)
			if err != nil {
				return nil, err
			}
		}
		valCurs, err := readFile(fmt.Sprintf("cmd/filecache/%s.xml", dateName))
		if err != nil {
			return nil, err
		}
		valueMap, err := parseValCurs(valCurs)
		if err != nil {
			return nil, err
		}
		c.m[dateName] = valueMap
		res := valueMap[code]
		result = &res
	}
	return result, nil
}

func parseValCurs(valCurs *ValCurs) (map[string]Currency, error) {
	valueMap := make(map[string]Currency)
	for _, valute := range valCurs.Valute {
		exRate, err := valute.getExRate()
		if err != nil {
			return nil, err
		}
		numCode, err := strconv.ParseInt(valute.NumCode, 10, 64)
		if err != nil {
			return nil, err
		}
		currency := Currency{
			NumCode: numCode,
			Code:    valute.CharCode,
			ExRate:  exRate,
			Symbol:  getSymbol(valute.CharCode),
		}
		valueMap[valute.CharCode] = currency
	}
	return valueMap, nil
}

func getSymbol(code string) string {
	switch code {
	case "EUR":
		return "€"
	case "USD":
		return "$"
	case "TRY":
		return "tl"
	case "RUB":
		return "₽"
	case "GBP":
		return "£"
	case "AMD":
		return "Dram"
	default:
		return ""
	}
}

func parseAmount(amount string, cash *CurCash, date time.Time) (int64, *Currency, error) {
	amount = strings.ToLower(amount)
	amount = strings.TrimSpace(amount)
	if strings.HasSuffix(amount, getSymbol("EUR")) {
		return getAmount("EUR", amount, cash, date)
	} else if strings.HasSuffix(amount, getSymbol("USD")) {
		return getAmount("USD", amount, cash, date)
	} else if strings.HasSuffix(amount, getSymbol("TRY")) {
		return getAmount("TRY", amount, cash, date)
	} else if strings.HasSuffix(amount, getSymbol("GBP")) {
		return getAmount("GBP", amount, cash, date)
	} else if strings.HasSuffix(amount, getSymbol("RUB")) {
		return getAmount("RUB", amount, cash, date)
	} else {
		value, err := strconv.ParseInt(amount, 10, 64)
		if err != nil {
			return 0, nil, err
		}
		currency, err := cash.Get(date, "RSD")
		if err != nil {
			return 0, nil, err
		}
		return value, currency, nil
	}
}

func getAmount(code string, amount string, cash *CurCash, date time.Time) (int64, *Currency, error) {
	currency, err := cash.Get(date, code)
	if err != nil {
		return 0, nil, err
	}
	return currency.getAmount(amount)
}

func createCurFile(date time.Time) error {
	valCurs, err := getAllValCurs(date)
	if err != nil {
		return err
	}
	data, err := xml.Marshal(valCurs)
	if err != nil {
		return err
	}
	err = saveFile(string(data), date)
	return err
}

func readFile(fileName string) (*ValCurs, error) {
	f, err := os.OpenFile(fileName, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var valCurs ValCurs
	err = xml.Unmarshal(data, &valCurs)
	if err != nil {
		return nil, err
	}
	return &valCurs, nil
}

func saveFile(message string, date time.Time) error {
	fileName := fmt.Sprintf("cmd/filecache/%v.xml", date.Format("2006-01-02"))

	_, err := os.Create(fileName)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Error().Stack().Err(err).Msg("error saving message on disc")
		return err
	}
	defer func() { _ = f.Close() }()

	if _, err = f.WriteString(message + "\n"); err != nil {
		log.Error().Stack().Err(err).Msg("error saving message on disc")
		return err
	}
	return nil
}

func getAllValCurs(date time.Time) (*ValCurs, error) {
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
