package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shelomentsevd/wolt_upload/db/sqlc/postgres"
	"github.com/spf13/cobra"
)

const (
	ReceiptID = iota
	Seller
	Venue
	DateTime
	CustomerName
	PaymentMethod
	DeliveryAddress
	ItemName
	Quantity
	UnitPrice
	LineTotal
	Subtotal
	Tax
	DeliveryFee
	Tip
	GrandTotal
)

var (
	postgresURL string
	inputDir    string
	rootCmd     = &cobra.Command{
		Use:   "receipt-processor",
		Short: "Process receipt CSV files",
		RunE:  run,
	}
)

func setupFlags() error {
	rootCmd.Flags().StringVar(&postgresURL, "postgres-url", "", "PostgreSQL connection string (can be set via PG_URL env)")
	rootCmd.Flags().StringVar(&inputDir, "input-dir", "", "Directory containing CSV files to process")
	if err := rootCmd.MarkFlagRequired("input-dir"); err != nil {
		return fmt.Errorf("failed to mark input-dir flag as required: %w", err)
	}
	return nil
}

func run(cmd *cobra.Command, args []string) error {
	if postgresURL == "" {
		postgresURL = os.Getenv("PG_URL")
		if postgresURL == "" {
			return fmt.Errorf("postgres-url flag or PG_URL environment variable is required")
		}
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	defer pool.Close()

	files := make(chan string)
	go processFiles(ctx, pool, files)

	err = filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".csv") {
			files <- path
		}
		return nil
	})

	close(files)
	return err
}

func processFiles(ctx context.Context, pool *pgxpool.Pool, files chan string) {
	sem := make(chan struct{}, 1)
	var wg sync.WaitGroup
	for file := range files {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore
		go func(filename string) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore
			if err := processFile(ctx, pool, filename); err != nil {
				fmt.Fprintf(os.Stderr, "Error processing file %s: %v\n", filename, err)
			}
		}(file)
	}
	wg.Wait()
}

func processFile(ctx context.Context, pool *pgxpool.Pool, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Skip header
	_, err = reader.Read()
	if err == io.EOF {
		return fmt.Errorf("empty file")
	}
	if err != nil {
		return fmt.Errorf("error reading header: %w", err)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading record: %w", err)
		}

		if err := processRecord(ctx, pool, record); err != nil {
			return fmt.Errorf("error processing record: %w", err)
		}
	}

	return nil
}

func processSeller(ctx context.Context, pool *pgxpool.Pool, record []string) (int, error) {
	q := postgres.New(pool)

	seller, err := q.CreateSeller(ctx, record[Seller])

	return seller, err
}

func processVenue(ctx context.Context, pool *pgxpool.Pool, record []string, seller int) (int, error) {
	q := postgres.New(pool)

	venue, err := q.CreateVenue(ctx, postgres.CreateVenueParams{
		Name: record[Venue],
		SellerID: pgtype.Int4{
			Int32: int32(seller),
			Valid: true,
		},
	})

	return venue, err
}

func parseDateTime(dateStr string) (pgtype.Timestamp, error) {
	t, err := time.Parse("02.01.2006 15:04", dateStr)
	if err != nil {
		return pgtype.Timestamp{}, fmt.Errorf("failed to parse date time: %w", err)
	}
	return pgtype.Timestamp{
		Time:  t,
		Valid: true,
	}, nil
}

func parseDecimal(value string) (pgtype.Numeric, error) {
	var num pgtype.Numeric
	if err := num.Scan(value); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("failed to parse decimal: %w", err)
	}
	return num, nil
}

func processReceipt(ctx context.Context, pool *pgxpool.Pool, record []string, seller, venue int) (string, error) {
	q := postgres.New(pool)

	date, _ := parseDateTime(record[DateTime])

	total, err := parseDecimal(record[GrandTotal])
	if err != nil {
		return "", err
	}

	var params = postgres.CreateReceiptParams{
		ID:       record[ReceiptID],
		Date:     date,
		SellerID: seller,
		VenueID:  venue,
		Total:    total,
	}

	receipt, err := q.CreateReceipt(ctx, params)
	if err != nil {
		return "", err
	}

	return receipt.ID, err
}

func processItem(ctx context.Context, pool *pgxpool.Pool, record []string, seller, venue int) (int, error) {
	q := postgres.New(pool)

	date, _ := parseDateTime(record[DateTime])

	lineTotal, err := parseDecimal(record[LineTotal])
	if err != nil {
		return 0, err
	}

	var params = postgres.CreateItemParams{
		ReceiptID: record[ReceiptID],
		Name:      record[ItemName],
		Date:      date,
		Quantity:  record[Quantity],
		UnitPrice: record[UnitPrice],
		LineTotal: lineTotal,
		SellerID:  seller,
		VenueID:   venue,
	}

	item, err := q.CreateItem(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("failed to create item: %w", err)
	}

	return item.ID, nil
}

func processRecord(ctx context.Context, pool *pgxpool.Pool, record []string) error {
	seller, err := processSeller(ctx, pool, record)
	if err != nil {
		return fmt.Errorf("failed to process seller: %w", err)
	}

	venue, err := processVenue(ctx, pool, record, seller)
	if err != nil {
		return fmt.Errorf("failed to process venue: %w", err)
	}

	_, err = processReceipt(ctx, pool, record, seller, venue)
	if err != nil {
		return fmt.Errorf("failed to process receipt: %w", err)
	}

	_, err = processItem(ctx, pool, record, seller, venue)
	if err != nil {
		return fmt.Errorf("failed to process item: %w", err)
	}

	return nil
}

func main() {
	if err := setupFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up flags: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
