package logic

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"rag/internal/svc"

	pb "github.com/qdrant/go-client/qdrant"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// qdrantBatchSize is the number of points to upsert per batch.
	qdrantBatchSize = 200
	// maxTextBytes limits the extracted text to prevent OOM from huge documents.
	maxTextBytes = 50 * 1024 * 1024 // 50MB
	// maxChunks limits the number of chunks generated.
	maxChunks = 10000
	// processingTimeout is the deadline for the entire async pipeline.
	processingTimeout = 10 * time.Minute
)

// ProcessAndIndexDocument handles the async pipeline: parse → chunk → embed → index.
// filePath is a temporary file that will be deleted after processing.
// It is shared between UploadDocumentLogic and external sync operations (e.g. Feishu).
func ProcessAndIndexDocument(ctx context.Context, svcCtx *svc.ServiceContext, docID, kbID, fileName, contentType, filePath string) {
	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, processingTimeout)
	defer cancel()

	logger := logx.WithContext(ctx)
	logger.Infof("Processing document %s (%s) for KB %s, path=%s", docID, fileName, kbID, filePath)

	// Clean up temp file when done
	defer os.Remove(filePath)

	// Read file for parsing
	data, err := os.ReadFile(filePath)
	if err != nil {
		logger.Errorf("read file %s: %v", filePath, err)
		svcCtx.Storage.UpdateDocStatus(ctx, docID, "FAILED", err.Error(), 0)
		return
	}

	// Step 1: Parse document to text
	text, err := parseDocument(data, contentType)
	// Release raw data immediately after parsing
	data = nil
	if err != nil {
		logger.Errorf("parse document %s: %v", docID, err)
		svcCtx.Storage.UpdateDocStatus(ctx, docID, "FAILED", err.Error(), 0)
		return
	}

	// Guard against excessive text
	if len(text) > maxTextBytes {
		logger.Errorf("extracted text too large: %d bytes (max %d)", len(text), maxTextBytes)
		svcCtx.Storage.UpdateDocStatus(ctx, docID, "FAILED",
			fmt.Sprintf("extracted text too large: %d bytes", len(text)), 0)
		return
	}

	// Step 2: Chunk text (byte-level, no rune conversion)
	chunks := chunkText(text, 1000, 200)
	text = "" // release text memory
	if len(chunks) == 0 {
		svcCtx.Storage.UpdateDocStatus(ctx, docID, "FAILED", "no content extracted", 0)
		return
	}

	// Limit chunk count
	if len(chunks) > maxChunks {
		logger.Infof("Document %s: truncating %d chunks to %d", docID, len(chunks), maxChunks)
		chunks = chunks[:maxChunks]
	}

	logger.Infof("Document %s: %d chunks generated", docID, len(chunks))

	// Step 3: Index in Qdrant (batched, streaming)
	totalIndexed := 0
	if svcCtx.Qdrant != nil {
		if err := svcCtx.Qdrant.EnsureCollection(ctx, kbID); err != nil {
			logger.Errorf("ensure qdrant collection: %v", err)
		}

		batch := make([]*pb.PointStruct, 0, qdrantBatchSize)
		for i, chunk := range chunks {
			// Verify context not cancelled
			select {
			case <-ctx.Done():
				svcCtx.Storage.UpdateDocStatus(ctx, docID, "FAILED", "processing timeout", i)
				return
			default:
			}

			chunkID := generateChunkUUID(docID, i)
			vector := textToVector(chunk, svcCtx.Config.Qdrant.VectorDim)
			entities := extractKeywords(chunk)
			entitiesJSON, _ := json.Marshal(entities)

			// Save chunk metadata to PostgreSQL
			if err := svcCtx.Storage.SaveChunk(ctx, chunkID, docID, kbID, chunk,
				string(entitiesJSON), int32(i), int64(len(chunk))); err != nil {
				logger.Errorf("save chunk %s: %v", chunkID, err)
				continue
			}

			point := &pb.PointStruct{
				Id: &pb.PointId{PointIdOptions: &pb.PointId_Uuid{Uuid: chunkID}},
				Vectors: &pb.Vectors{
					VectorsOptions: &pb.Vectors_Vector{
						Vector: &pb.Vector{Data: vector},
					},
				},
				Payload: map[string]*pb.Value{
					"doc_id":      {Kind: &pb.Value_StringValue{StringValue: docID}},
					"kb_id":       {Kind: &pb.Value_StringValue{StringValue: kbID}},
					"doc_name":    {Kind: &pb.Value_StringValue{StringValue: fileName}},
					"content":     {Kind: &pb.Value_StringValue{StringValue: chunk}},
					"chunk_index": {Kind: &pb.Value_IntegerValue{IntegerValue: int64(i)}},
					"entities":    {Kind: &pb.Value_StringValue{StringValue: string(entitiesJSON)}},
				},
			}
			batch = append(batch, point)

			// Flush batch when full
			if len(batch) >= qdrantBatchSize {
				if err := svcCtx.Qdrant.UpsertPoints(ctx, kbID, batch); err != nil {
					logger.Errorf("upsert qdrant batch (chunk %d): %v", i, err)
					svcCtx.Storage.UpdateDocStatus(ctx, docID, "FAILED", err.Error(), i)
					return
				}
				totalIndexed += len(batch)
				batch = batch[:0] // reuse slice, release points for GC
			}
		}

		// Flush remaining batch
		if len(batch) > 0 {
			if err := svcCtx.Qdrant.UpsertPoints(ctx, kbID, batch); err != nil {
				logger.Errorf("upsert qdrant final batch: %v", err)
				svcCtx.Storage.UpdateDocStatus(ctx, docID, "FAILED", err.Error(), totalIndexed)
				return
			}
			totalIndexed += len(batch)
		}

		svcCtx.Storage.IncrementKBChunkCount(ctx, kbID, totalIndexed)
	}

	// Step 4: Build graph in HugeGraph (skip if too many chunks to avoid OOM)
	if svcCtx.HugeGraph != nil && len(chunks) <= 5000 {
		if err := svcCtx.HugeGraph.AddDocumentVertex(ctx, docID, fileName, kbID); err != nil {
			logger.Errorf("add doc vertex: %v", err)
		}

		for i, chunk := range chunks {
			// Verify context not cancelled
			select {
			case <-ctx.Done():
				break
			default:
			}

			chunkID := generateChunkUUID(docID, i)
			if err := svcCtx.HugeGraph.AddChunkVertex(ctx, chunkID, docID, kbID, i); err != nil {
				logger.Errorf("add chunk vertex: %v", err)
				continue
			}

			entities := extractKeywords(chunk)
			for _, entity := range entities {
				entityID := fmt.Sprintf("%s_%s", kbID, sanitizeEntityID(entity))
				_ = svcCtx.HugeGraph.AddEntity(ctx, entityID, entity, "CONCEPT", kbID, "")
				_ = svcCtx.HugeGraph.LinkChunkToEntity(ctx, chunkID, entityID, kbID)
			}

			// Add relations between co-occurring entities
			for a := 0; a < len(entities); a++ {
				for b := a + 1; b < len(entities); b++ {
					entityA := fmt.Sprintf("%s_%s", kbID, sanitizeEntityID(entities[a]))
					entityB := fmt.Sprintf("%s_%s", kbID, sanitizeEntityID(entities[b]))
					_ = svcCtx.HugeGraph.AddRelation(ctx, entityA, entityB, "CO_OCCURS", chunkID, kbID)
				}
			}
		}
	}

	// Done
	svcCtx.Storage.UpdateDocStatus(ctx, docID, "READY", "", len(chunks))
	logger.Infof("Document %s processed successfully: %d chunks indexed", docID, totalIndexed)
}

// parseDocument extracts text from various document formats.
func parseDocument(data []byte, contentType string) (string, error) {
	ct := strings.ToLower(contentType)

	switch {
	case strings.Contains(ct, "text/plain"):
		return string(data), nil

	case strings.Contains(ct, "text/html") || strings.Contains(ct, "html"):
		return stripHTML(data), nil

	case strings.Contains(ct, "text/markdown") || strings.Contains(ct, "md"):
		return string(data), nil

	case strings.Contains(ct, "pdf"):
		return extractPDFText(data), nil

	case strings.Contains(ct, "docx") || strings.Contains(ct, "vnd.openxmlformats-officedocument.wordprocessingml"):
		return parseDocx(data)

	case strings.Contains(ct, "xlsx") || strings.Contains(ct, "vnd.openxmlformats-officedocument.spreadsheetml"):
		return parseXlsx(data)

	case strings.Contains(ct, "pptx") || strings.Contains(ct, "vnd.openxmlformats-officedocument.presentationml"):
		return parsePptx(data)

	case strings.Contains(ct, "csv"):
		return string(data), nil

	case strings.Contains(ct, "json") || strings.Contains(ct, "application/json"):
		return extractJSONText(data), nil

	default:
		if isText(data) {
			return string(data), nil
		}
		return "", fmt.Errorf("unsupported content type: %s", contentType)
	}
}

func stripHTML(data []byte) string {
	s := string(data)
	var result strings.Builder
	inTag := false
	for _, c := range s {
		if c == '<' {
			inTag = true
			continue
		}
		if c == '>' {
			inTag = false
			result.WriteByte(' ')
			continue
		}
		if !inTag {
			result.WriteRune(c)
		}
	}
	return strings.TrimSpace(result.String())
}

func isText(data []byte) bool {
	nonPrintable := 0
	for _, b := range data {
		if b == 0 {
			return false
		}
		if b < 0x20 && b != '\n' && b != '\r' && b != '\t' {
			nonPrintable++
		}
	}
	return float64(nonPrintable)/float64(len(data)) < 0.1
}

// chunkText splits text into overlapping chunks using byte-level indexing.
// This avoids the 4x memory blowup of []rune(text) for ASCII-heavy text.
func chunkText(text string, chunkSize, overlap int) []string {
	if len(text) <= chunkSize {
		return []string{text}
	}

	// Use byte-level operations — most Chinese text still works because
	// we only look for ASCII boundary markers (\n, . etc.)
	b := []byte(text)
	var chunks []string

	start := 0
	for start < len(b) {
		end := start + chunkSize
		if end > len(b) {
			end = len(b)
		}

		// Try to break at natural boundary
		if end < len(b) {
			searchStart := end - 100
			if searchStart < start {
				searchStart = start
			}
			searchSlice := string(b[searchStart:end])
			if idx := strings.LastIndex(searchSlice, "\n\n"); idx >= 0 {
				end = searchStart + idx
			} else if idx := strings.LastIndex(searchSlice, "\n"); idx >= 0 {
				end = searchStart + idx
			} else if idx := strings.LastIndex(searchSlice, ". "); idx >= 0 {
				end = searchStart + idx + 1
			}
		}

		chunk := strings.TrimSpace(string(b[start:end]))
		if len(chunk) > 0 {
			chunks = append(chunks, chunk)
			if len(chunks) >= maxChunks {
				break
			}
		}

		start = end - overlap
		if start <= 0 || start >= len(b) {
			break
		}
	}

	return chunks
}

// extractKeywords extracts important words from text (simplified TF-based).
func extractKeywords(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	wordFreq := make(map[string]int)
	for _, w := range words {
		w = strings.Trim(w, ".,;:!?()[]{}\"'`")
		if len(w) > 2 {
			wordFreq[w]++
		}
	}

	type wf struct {
		word string
		freq int
	}
	var sorted []wf
	for w, f := range wordFreq {
		if f >= 2 {
			sorted = append(sorted, wf{w, f})
		}
	}

	// Sort by frequency descending
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].freq > sorted[i].freq {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	maxN := 10
	if len(sorted) < maxN {
		maxN = len(sorted)
	}
	result := make([]string, maxN)
	for i := 0; i < maxN; i++ {
		result[i] = sorted[i].word
	}
	return result
}

// textToVector converts text to a simple average vector (placeholder).
func textToVector(text string, dim int) []float32 {
	vec := make([]float32, dim)
	b := []byte(strings.ToLower(text))

	for i, byt := range b {
		idx := (int(byt) * (i + 1)) % dim
		vec[idx] += 0.01
	}

	var sum float32
	for _, v := range vec {
		sum += v * v
	}
	if sum > 0 {
		norm := float32(1.0 / float64(sqrtF(sum)))
		for i := range vec {
			vec[i] *= norm
		}
	}
	return vec
}

func sqrtF(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}

// generateChunkUUID creates a deterministic UUID for a chunk from (docID, index).
// Uses SHA-1 hash formatted as UUID v5 so Qdrant accepts it as PointId_Uuid.
func generateChunkUUID(docID string, index int) string {
	input := fmt.Sprintf("%s:%d", docID, index)
	h := sha1.New()
	h.Write([]byte(input))
	b := h.Sum(nil)
	b[6] = (b[6] & 0x0f) | 0x50 // UUID version 5
	b[8] = (b[8] & 0x3f) | 0x80 // UUID variant
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func sanitizeEntityID(entity string) string {
	s := strings.ToLower(entity)
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, s)
	return strings.Trim(s, "_")
}

// ---- Document Format Parsers ----

const docxNS = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"

// parseDocx extracts text from a DOCX file (OOXML ZIP package).
// Uses streaming XML parsing to avoid loading the entire document.xml.
func parseDocx(data []byte) (string, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to open docx as zip: %w", err)
	}

	for _, f := range zipReader.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open document.xml: %w", err)
			}
			defer rc.Close()
			return extractDocxXMLTextStream(rc)
		}
	}
	return "", fmt.Errorf("word/document.xml not found in docx archive")
}

// extractDocxXMLTextStream extracts text using streaming XML to control memory.
func extractDocxXMLTextStream(r io.Reader) (string, error) {
	decoder := xml.NewDecoder(r)
	var buf strings.Builder
	var inText bool
	var textBytes int64

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch el := tok.(type) {
		case xml.StartElement:
			if el.Name.Local == "t" || (el.Name.Space == docxNS && el.Name.Local == "t") {
				inText = true
			}
			if el.Name.Local == "p" || (el.Name.Space == docxNS && el.Name.Local == "p") {
				if buf.Len() > 0 {
					buf.WriteString("\n")
				}
			}
		case xml.EndElement:
			if el.Name.Local == "t" || (el.Name.Space == docxNS && el.Name.Local == "t") {
				inText = false
			}
		case xml.CharData:
			if inText {
				n, _ := buf.Write(el)
				textBytes += int64(n)
				if textBytes > maxTextBytes {
					return "", fmt.Errorf("docx text exceeds max size %d bytes", maxTextBytes)
				}
			}
		}
	}

	result := strings.TrimSpace(buf.String())
	if result == "" {
		return "", fmt.Errorf("no text content found in docx")
	}
	return result, nil
}

// parseXlsx extracts text from an XLSX file (OOXML ZIP package).
func parseXlsx(data []byte) (string, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to open xlsx as zip: %w", err)
	}

	var sharedStrings []string
	for _, f := range zipReader.File {
		if f.Name == "xl/sharedStrings.xml" {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			xmlData, _ := io.ReadAll(rc)
			rc.Close()
			sharedStrings = parseXlsxSharedStrings(xmlData)
			break
		}
	}

	var buf strings.Builder
	for _, f := range zipReader.File {
		if strings.HasPrefix(f.Name, "xl/worksheets/sheet") && strings.HasSuffix(f.Name, ".xml") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			xmlData, _ := io.ReadAll(rc)
			rc.Close()
			buf.WriteString(parseXlsxSheet(xmlData, sharedStrings))
		}
	}

	result := strings.TrimSpace(buf.String())
	if result == "" {
		return "", fmt.Errorf("no text content found in xlsx")
	}
	return result, nil
}

func parseXlsxSharedStrings(xmlData []byte) []string {
	var strings []string
	decoder := xml.NewDecoder(bytes.NewReader(xmlData))
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		if el, ok := tok.(xml.StartElement); ok && el.Name.Local == "t" {
			for {
				tok2, err := decoder.Token()
				if err != nil {
					break
				}
				if cd, ok := tok2.(xml.CharData); ok {
					strings = append(strings, string(cd))
				}
				if end, ok := tok2.(xml.EndElement); ok && end.Name.Local == "t" {
					break
				}
			}
		}
	}
	return strings
}

func parseXlsxSheet(xmlData []byte, sharedStrings []string) string {
	var buf strings.Builder
	decoder := xml.NewDecoder(bytes.NewReader(xmlData))
	var inValue bool

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch el := tok.(type) {
		case xml.StartElement:
			if el.Name.Local == "v" {
				inValue = true
			}
			if el.Name.Local == "t" {
				buf.WriteString(extractElementText(&el, decoder))
			}
		case xml.EndElement:
			if el.Name.Local == "v" {
				inValue = false
			}
			if el.Name.Local == "row" {
				buf.WriteString("\n")
			}
		case xml.CharData:
			if inValue {
				v := strings.TrimSpace(string(el))
				if v != "" {
					if idx, err := parseXlsxRef(v); err == nil && idx >= 0 && idx < len(sharedStrings) {
						buf.WriteString(sharedStrings[idx])
					}
				}
			}
		}
	}
	return buf.String()
}

func extractElementText(start *xml.StartElement, decoder *xml.Decoder) string {
	var buf strings.Builder
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		if end, ok := tok.(xml.EndElement); ok && end.Name.Local == start.Name.Local {
			break
		}
		if cd, ok := tok.(xml.CharData); ok {
			buf.Write(cd)
		}
	}
	return buf.String()
}

func parseXlsxRef(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not an integer: %s", s)
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// parsePptx extracts text from a PPTX file (OOXML ZIP package).
func parsePptx(data []byte) (string, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to open pptx as zip: %w", err)
	}

	var buf strings.Builder
	for _, f := range zipReader.File {
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			xmlData, _ := io.ReadAll(rc)
			rc.Close()

			decoder := xml.NewDecoder(bytes.NewReader(xmlData))
			for {
				tok, err := decoder.Token()
				if err != nil {
					break
				}
				if el, ok := tok.(xml.StartElement); ok && el.Name.Local == "t" {
					for {
						tok2, err := decoder.Token()
						if err != nil {
							break
						}
						if cd, ok := tok2.(xml.CharData); ok {
							buf.Write(cd)
						}
						if end, ok := tok2.(xml.EndElement); ok && end.Name.Local == "t" {
							break
						}
					}
					buf.WriteString("\n")
				}
			}
		}
	}

	result := strings.TrimSpace(buf.String())
	if result == "" {
		return "", fmt.Errorf("no text content found in pptx")
	}
	return result, nil
}

// extractPDFText performs basic PDF text extraction.
func extractPDFText(data []byte) string {
	text := string(data)
	var buf strings.Builder

	for {
		btIdx := strings.Index(text, "BT")
		if btIdx == -1 {
			break
		}
		etIdx := strings.Index(text[btIdx:], "ET")
		if etIdx == -1 {
			break
		}
		block := text[btIdx : btIdx+etIdx]

		for {
			left := strings.Index(block, "(")
			if left == -1 {
				break
			}
			block = block[left:]
			right := strings.IndexByte(block[1:], ')')
			if right == -1 {
				break
			}
			content := block[1 : 1+right]
			if strings.TrimSpace(content) != "" {
				buf.WriteString(content)
				buf.WriteString(" ")
			}
			block = block[1+right+1:]
		}
		text = text[btIdx+etIdx:]
	}

	return strings.TrimSpace(buf.String())
}

// extractJSONText flattens JSON into readable text for indexing.
func extractJSONText(data []byte) string {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return string(data)
	}
	return flattenJSON(v)
}

func flattenJSON(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%v", val)
	case bool:
		return fmt.Sprintf("%v", val)
	case []any:
		var pieces []string
		for _, item := range val {
			pieces = append(pieces, flattenJSON(item))
		}
		return strings.Join(pieces, " ")
	case map[string]any:
		var pieces []string
		for k, item := range val {
			pieces = append(pieces, k+": "+flattenJSON(item))
		}
		return strings.Join(pieces, " ")
	}
	return ""
}
