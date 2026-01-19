package converter

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"

	"github.com/theoremus-urban-solutions/gtfs-netex-converter/netex"
)

// StreamingXMLWriter writes NeTEx XML output in streaming mode to reduce memory usage
type StreamingXMLWriter struct {
	output  io.Writer
	encoder *xml.Encoder
}

// NewStreamingXMLWriter creates a new streaming XML writer
func NewStreamingXMLWriter(output io.Writer) *StreamingXMLWriter {
	encoder := xml.NewEncoder(output)
	encoder.Indent("", "  ")

	return &StreamingXMLWriter{
		output:  output,
		encoder: encoder,
	}
}

// WritePublicationDelivery writes the NeTEx PublicationDelivery in streaming mode
func (w *StreamingXMLWriter) WritePublicationDelivery(pd *netex.PublicationDelivery) error {
	// Write XML header
	if _, err := w.output.Write([]byte(xml.Header)); err != nil {
		return fmt.Errorf("failed to write XML header: %w", err)
	}

	// Encode the entire structure
	// Note: For true streaming, we'd need to manually write elements in chunks
	// For now, this uses the standard encoder but writes directly to file
	// which avoids keeping the entire marshaled string in memory
	if err := w.encoder.Encode(pd); err != nil {
		return fmt.Errorf("failed to encode XML: %w", err)
	}

	// Flush any buffered data
	if err := w.encoder.Flush(); err != nil {
		return fmt.Errorf("failed to flush XML encoder: %w", err)
	}

	return nil
}

// WriteToFile writes the NeTEx data directly to a file in streaming mode
func WriteToFile(filepath string, pd *netex.PublicationDelivery) error {
	file, err := os.Create(filepath) //nolint:gosec // user input controlled
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	writer := NewStreamingXMLWriter(file)
	if err := writer.WritePublicationDelivery(pd); err != nil {
		return err
	}

	return nil
}
