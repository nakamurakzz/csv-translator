package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	translate "cloud.google.com/go/translate/apiv3"
	translatepb "cloud.google.com/go/translate/apiv3/translatepb"
)

// Translator wraps a Translate client with configuration and cache.
type Translator struct {
	client      *translate.TranslationClient
	projectID   string
	location    string
	cache       map[string]string
	excludeCols map[string]struct{}
	ctx         context.Context
}

// NewTranslator initializes a new Translator.
func NewTranslator(ctx context.Context, projectID, location string, exclude []string) (*Translator, error) {
	client, err := translate.NewTranslationClient(ctx)
	if err != nil {
		return nil, err
	}
	excludeMap := make(map[string]struct{}, len(exclude))
	for _, col := range exclude {
		excludeMap[col] = struct{}{}
	}

	return &Translator{
		client:      client,
		projectID:   projectID,
		location:    location,
		cache:       make(map[string]string),
		excludeCols: excludeMap,
		ctx:         ctx,
	}, nil
}

// TranslateCell returns the translated text for a single cell.
func (t *Translator) TranslateCell(text string, header string) string {
	if text == "" {
		return text
	}
	if _, skip := t.excludeCols[header]; skip {
		return text
	}
	if cached, ok := t.cache[text]; ok {
		return cached
	}

	req := &translatepb.TranslateTextRequest{
		Parent:             fmt.Sprintf("projects/%s/locations/%s", t.projectID, t.location),
		Contents:           []string{text},
		MimeType:           "text/plain",
		TargetLanguageCode: "en",
	}
	resp, err := t.client.TranslateText(t.ctx, req)
	if err != nil {
		log.Printf("Translation error for column %s: %v", header, err)
		return text
	}
	if len(resp.GetTranslations()) > 0 {
		translated := resp.GetTranslations()[0].GetTranslatedText()
		t.cache[text] = translated
		return translated
	}
	return text
}

// ProcessCSV reads input CSV, translates cells, and writes to output.
func ProcessCSV(inputPath string, translator *Translator) error {
	inFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer inFile.Close()

	r := csv.NewReader(inFile)
	headers, err := r.Read()
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	outPath := strings.TrimSuffix(inputPath, ".csv") + "_translated.csv"
	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer outFile.Close()

	w := csv.NewWriter(outFile)
	defer w.Flush()

	if err := w.Write(headers); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	rowNum := 1
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read row %d: %w", rowNum, err)
		}

		fmt.Printf("Processing row %d\n", rowNum)
		rowNum++

		for i, cell := range rec {
			rec[i] = translator.TranslateCell(cell, headers[i])
		}

		if err := w.Write(rec); err != nil {
			return fmt.Errorf("write row %d: %w", rowNum-1, err)
		}
	}

	fmt.Printf("Translated CSV written to %s\n", outPath)
	return nil
}

func main() {
	// Parse command-line arguments
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <input.csv> <exclude_cols>\n", os.Args[0])
	}
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	inputFile := flag.Arg(0)      // "input.csv"
	excludeColsArg := flag.Arg(1) // "id,tel,postal"
	// Load project ID from environment
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatal("Environment variable GOOGLE_CLOUD_PROJECT not set")
	}

	// Initialize translator
	excludeCols := strings.Split(excludeColsArg, ",")
	ctx := context.Background()
	translator, err := NewTranslator(ctx, projectID, "global", excludeCols)
	if err != nil {
		log.Fatalf("Failed to create translator: %v", err)
	}
	defer translator.client.Close()

	// Process CSV file
	if err := ProcessCSV(inputFile, translator); err != nil {
		log.Fatalf("Error processing CSV: %v", err)
	}
}
