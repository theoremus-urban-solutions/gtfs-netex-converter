package converter

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

// StreamingCSVReader reads CSV files in batches to reduce memory usage
type StreamingCSVReader struct {
	filepath  string
	batchSize int
}

// NewStreamingCSVReader creates a new streaming CSV reader
func NewStreamingCSVReader(filepath string, batchSize int) *StreamingCSVReader {
	if batchSize <= 0 {
		batchSize = 10000 // Default batch size
	}
	return &StreamingCSVReader{
		filepath:  filepath,
		batchSize: batchSize,
	}
}

// BatchResult contains a batch of records and metadata
type BatchResult struct {
	Headers []string
	Records [][]string
	Offset  int
	IsLast  bool
	Error   error
}

// ReadInBatches reads CSV file in batches and sends them through a channel
func (sr *StreamingCSVReader) ReadInBatches() (<-chan BatchResult, error) {
	file, err := os.Open(sr.filepath)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	// Read headers first
	headers, err := reader.Read()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}

	resultChan := make(chan BatchResult, 2) // Buffer 2 batches

	// Start goroutine to read batches
	go func() {
		defer func() {
			_ = file.Close()
		}()
		defer close(resultChan)

		offset := 0
		for {
			batch := make([][]string, 0, sr.batchSize)

			// Read batch
			for i := 0; i < sr.batchSize; i++ {
				record, err := reader.Read()
				if err == io.EOF {
					// Send last batch if not empty
					if len(batch) > 0 {
						resultChan <- BatchResult{
							Headers: headers,
							Records: batch,
							Offset:  offset,
							IsLast:  true,
						}
					} else if offset == 0 {
						// Empty file (only headers)
						resultChan <- BatchResult{
							Headers: headers,
							Records: [][]string{},
							Offset:  0,
							IsLast:  true,
						}
					}
					return
				}
				if err != nil {
					resultChan <- BatchResult{Error: err}
					return
				}
				batch = append(batch, record)
			}

			// Send batch
			resultChan <- BatchResult{
				Headers: headers,
				Records: batch,
				Offset:  offset,
				IsLast:  false,
			}
			offset += len(batch)
		}
	}()

	return resultChan, nil
}

// ProcessBatches processes CSV batches with a callback function
func (sr *StreamingCSVReader) ProcessBatches(processor func(headers []string, records [][]string, offset int) error) error {
	batches, err := sr.ReadInBatches()
	if err != nil {
		return err
	}

	for batch := range batches {
		if batch.Error != nil {
			return batch.Error
		}
		if err := processor(batch.Headers, batch.Records, batch.Offset); err != nil {
			return fmt.Errorf("batch processing failed at offset %d: %w", batch.Offset, err)
		}
	}

	return nil
}

// CountRecordsInCSV counts total records in CSV file without loading all into memory
func CountRecordsInCSV(filepath string) (int, error) {
	file, err := os.Open(filepath) //nolint:gosec // filepath is controlled by application, not user input
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = file.Close()
	}()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	// Skip header
	if _, err := reader.Read(); err != nil {
		return 0, err
	}

	count := 0
	for {
		_, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
		count++
	}

	return count, nil
}
