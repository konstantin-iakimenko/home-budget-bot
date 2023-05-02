package export

//package main
//
//import (
//	"context"
//	"fmt"
//	"github.com/jackc/pgx/v4/pgxpool"
//	"github.com/xuri/excelize/v2"
//	"os"
//	"strconv"
//	"strings"
//	"time"
//)
//
//type Bill struct {
//	Id          int64
//	UserId      int64
//	BoughtAt    time.Time
//	Description string
//	Category    string
//	Amount      int64
//	Currency    int64
//	AmountRub   int64
//	AmountUsd   int64
//	CreatedAt   time.Time
//}
//
//type Data struct {
//	monthsRub map[string](map[string]uint64)
//	monthsUsd map[string](map[string]uint64)
//}
//
//type MonthName struct {
//	cell string
//	name string
//}
//
//func main() {
//	ctx := context.Background()
//
//	config, err := pgxpool.ParseConfig(os.Getenv("PG_HOMEBUDGET_DB")) // DatabaseURL
//	if err != nil {
//		errHandler(err)
//	}
//
//	pool, err := pgxpool.ConnectConfig(ctx, config)
//	if err != nil {
//		errHandler(err)
//	}
//
//	cells := []string{"B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "O", "P", "Q", "R", "S", "T", "U", "V", "X", "Y", "Z"}
//	monthNames := []MonthName{}
//	allCategories := map[string]int{}
//
//	bills, err := loadBills(ctx, pool)
//	if err != nil {
//		errHandler(err)
//	}
//
//	data := Data{
//		monthsRub: make(map[string](map[string]uint64)),
//		monthsUsd: make(map[string](map[string]uint64)),
//	}
//	var j = 0
//	var currentMonth = ""
//	for _, bill := range bills {
//		month := bill.BoughtAt.Month().String()[0:3] + " " + fmt.Sprintf("%d", bill.BoughtAt.Year())
//		if currentMonth != month {
//			currentMonth = month
//			monthNames = append(monthNames, MonthName{
//				cell: cells[j],
//				name: month,
//			})
//			j++
//		}
//
//		category := bill.Category
//		allCategories[category] = 1
//		amountRub := bill.AmountRub
//		amountUsd := bill.AmountUsd
//
//		if _, ok := data.monthsRub[month]; !ok {
//			data.monthsRub[month] = make(map[string]uint64)
//		}
//		data.monthsRub[month][category] += uint64(amountRub)
//
//		if _, ok := data.monthsUsd[month]; !ok {
//			data.monthsUsd[month] = make(map[string]uint64)
//		}
//		data.monthsUsd[month][category] += uint64(amountUsd)
//	}
//
//	f := excelize.NewFile()
//
//	idxRub := saveToExcel(f, "RUB", data.monthsRub, monthNames, allCategories)
//	saveToExcel(f, "USD", data.monthsUsd, monthNames, allCategories)
//
//	f.SetActiveSheet(idxRub)
//	if err := f.SaveAs("/Users/jakimenko/Downloads/tmp2/report.xlsx"); err != nil {
//		errHandler(err)
//	}
//
//	pool.Close()
//}
//
//func saveToExcel(f *excelize.File, sheet string, months map[string](map[string]uint64), monthNames []MonthName, allCategories map[string]int) int {
//	sheetIdx, err := f.NewSheet(sheet)
//	if err != nil {
//		errHandler(err)
//	}
//
//	for _, monthName := range monthNames {
//		err = f.SetCellValue(sheet, fmt.Sprintf("%s%d", monthName.cell, 1), monthName.name)
//		if err != nil {
//			errHandler(err)
//		}
//	}
//
//	categoryCells := map[string]int{}
//	i := 2
//	for category, _ := range allCategories {
//		err = f.SetCellValue(sheet, fmt.Sprintf("%s%d", "A", i), category)
//		categoryCells[category] = i
//		i++
//		if err != nil {
//			errHandler(err)
//		}
//	}
//	categoryCells["Итого"] = i
//
//	for _, monthName := range monthNames {
//		sum := 0.0
//		for category, _ := range allCategories {
//			valueStr := strconv.FormatUint(months[monthName.name][category], 10)
//			value := "0.00"
//			valueFloat := 0.0
//			if strings.Trim(valueStr, "") != "" && len(valueStr) > 2 {
//				value = valueStr[0:len(valueStr)-2] + "." + valueStr[len(valueStr)-2:]
//				valueFloat, err = strconv.ParseFloat(value, 64)
//				if err != nil {
//					errHandler(err)
//				}
//			}
//
//			err = f.SetCellValue(sheet, fmt.Sprintf("%s%d", monthName.cell, categoryCells[category]), valueFloat)
//			sum += valueFloat
//			if err != nil {
//				errHandler(err)
//			}
//		}
//		err = f.SetCellValue(sheet, fmt.Sprintf("%s%d", monthName.cell, categoryCells["Итого"]), sum)
//		if err != nil {
//			errHandler(err)
//		}
//	}
//	return sheetIdx
//}
//
//func loadBills(ctx context.Context, pool *pgxpool.Pool) ([]Bill, error) {
//	bills := make([]Bill, 0)
//	b, err := pool.Query(ctx, "select * from bills order by bought_at")
//	defer b.Close()
//	if err != nil {
//		return nil, err
//	}
//	for b.Next() {
//		var bill Bill
//		err = b.Scan(&bill.Id, &bill.UserId, &bill.BoughtAt, &bill.Description, &bill.Category, &bill.Amount, &bill.Currency, &bill.AmountRub, &bill.AmountUsd, &bill.CreatedAt)
//		if err != nil {
//			return nil, err
//		}
//		bills = append(bills, bill)
//	}
//	return bills, nil
//}
//
//func errHandler(err error) {
//	if err != nil {
//		fmt.Println(err)
//		os.Exit(1)
//	}
//}
