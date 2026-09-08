//go:build unit || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

// receivedChunk is one PUT request as observed by the mock server.
type receivedChunk struct {
	start, end, total int64
	bodyLen           int64
}

// spawnUploadServer returns a server that accepts chunked PUT requests on
// /upload and records the Content-Range header and body size of each one. Any
// other path answers with 200 so that the session keepalive query issued by
// uploadPartFile does not fail the test.
func spawnUploadServer(t *testing.T, chunks *[]receivedChunk, body *[]byte) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT on /upload, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var chunk receivedChunk
		rangeHeader := r.Header.Get("Content-Range")
		if _, err := fmt.Sscanf(rangeHeader, "bytes %d-%d/%d", &chunk.start, &chunk.end, &chunk.total); err != nil {
			t.Errorf("could not parse Content-Range %q: %s", rangeHeader, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		received := make([]byte, 0, r.ContentLength)
		buf := make([]byte, 4096)
		for {
			n, err := r.Body.Read(buf)
			received = append(received, buf[:n]...)
			if err != nil {
				break
			}
		}

		chunk.bodyLen = int64(len(received))
		*chunks = append(*chunks, chunk)
		*body = append(*body, received...)

		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return httptest.NewServer(mux)
}

// writeTestFile creates a file of the given size whose content is a
// non-repeating byte pattern, so that reassembled chunks can be compared
// against the original.
func writeTestFile(t *testing.T, size int64) (string, []byte) {
	content := make([]byte, size)
	for i := range content {
		content[i] = byte(i % 251)
	}

	path := filepath.Join(t.TempDir(), "test-file.iso")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatalf("could not write test file: %s", err)
	}

	return path, content
}

// TestUploadFileChunkBoundaries checks that uploadFile transfers a file
// completely and in contiguous chunks. The cases marked as exact multiples are
// regression tests: such a file used to be reported as failed after every byte
// had already been transferred, because io.ReadFull signals a zero-length read
// with io.EOF rather than io.ErrUnexpectedEOF.
//
// Note that uploadPieceSize is only honoured when it is larger than 1024 and
// strictly smaller than the file size; otherwise defaultPieceSize applies. The
// expectedPieceSize column spells out which of the two a case exercises.
func TestUploadFileChunkBoundaries(t *testing.T) {
	const pieceSize int64 = 2048

	testCases := []struct {
		name              string
		fileSize          int64
		uploadPieceSize   int64
		expectedPieceSize int64
		expectedChunks    int
	}{
		{"ExactMultipleTwoChunks", 2 * pieceSize, pieceSize, pieceSize, 2},
		{"ExactMultipleThreeChunks", 3 * pieceSize, pieceSize, pieceSize, 3},
		{"MultiplePlusOneByte", 2*pieceSize + 1, pieceSize, pieceSize, 3},
		{"MultipleMinusOneByte", 2*pieceSize - 1, pieceSize, pieceSize, 2},
		{"PieceSizeEqualsFileSize", pieceSize, pieceSize, defaultPieceSize, 1},
		{"PieceSizeBelowMinimum", 4 * pieceSize, 512, defaultPieceSize, 1},
		{"SmallerThanPieceSize", 1024, pieceSize, defaultPieceSize, 1},
		{"SingleByte", 1, pieceSize, defaultPieceSize, 1},
		{"DefaultPieceSizeExactMultiple", defaultPieceSize, 0, defaultPieceSize, 1},
		{"DefaultPieceSizeTwoChunks", 2 * defaultPieceSize, 0, defaultPieceSize, 2},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var chunks []receivedChunk
			var receivedBody []byte

			server := spawnUploadServer(t, &chunks, &receivedBody)
			defer server.Close()

			serverUrl, err := url.ParseRequestURI(server.URL)
			if err != nil {
				t.Fatalf("could not parse server URL: %s", err)
			}

			filePath, content := writeTestFile(t, testCase.fileSize)

			var uploadError error
			var callbackValues []int64

			client := NewVCDClient(*serverUrl, true)
			uploadSize, err := uploadFile(&client.Client, filePath, uploadDetails{
				uploadLink:       server.URL + "/upload",
				fileSizeToUpload: testCase.fileSize,
				uploadPieceSize:  testCase.uploadPieceSize,
				allFilesSize:     testCase.fileSize,
				callBack: func(bytesUpload, totalSize int64) {
					callbackValues = append(callbackValues, bytesUpload)
				},
				uploadError: &uploadError,
			})

			if err != nil {
				t.Fatalf("upload of %d bytes failed: %s", testCase.fileSize, err)
			}
			if uploadError != nil {
				t.Errorf("uploadError was set although upload succeeded: %s", uploadError)
			}
			if uploadSize != testCase.fileSize {
				t.Errorf("uploadFile returned %d bytes, expected %d", uploadSize, testCase.fileSize)
			}

			// The number of chunks must not change with this fix: it is
			// determined by the piece size alone.
			if len(chunks) != testCase.expectedChunks {
				t.Errorf("server received %d chunks, expected %d", len(chunks), testCase.expectedChunks)
			}

			// Chunks must tile the file without gaps or overlaps.
			var expectedOffset int64
			for i, chunk := range chunks {
				if chunk.start != expectedOffset {
					t.Errorf("chunk %d starts at offset %d, expected %d", i, chunk.start, expectedOffset)
				}
				if chunk.end != chunk.start+chunk.bodyLen-1 {
					t.Errorf("chunk %d has range %d-%d but carries %d bytes", i, chunk.start, chunk.end, chunk.bodyLen)
				}
				if chunk.total != testCase.fileSize {
					t.Errorf("chunk %d reports total size %d, expected %d", i, chunk.total, testCase.fileSize)
				}
				if chunk.bodyLen > testCase.expectedPieceSize {
					t.Errorf("chunk %d carries %d bytes, more than the piece size %d", i, chunk.bodyLen, testCase.expectedPieceSize)
				}
				expectedOffset += chunk.bodyLen
			}
			if expectedOffset != testCase.fileSize {
				t.Errorf("chunks cover %d bytes, expected %d", expectedOffset, testCase.fileSize)
			}

			// The reassembled body must be the file, byte for byte.
			if len(receivedBody) != len(content) {
				t.Fatalf("server received %d bytes, expected %d", len(receivedBody), len(content))
			}
			for i := range content {
				if receivedBody[i] != content[i] {
					t.Fatalf("received body differs from source file at offset %d", i)
				}
			}

			// The progress callback must be strictly increasing and end at the
			// full file size.
			if len(callbackValues) != len(chunks) {
				t.Errorf("callback was invoked %d times, expected %d", len(callbackValues), len(chunks))
			}
			var previous int64
			for i, value := range callbackValues {
				if i > 0 && value <= previous {
					t.Errorf("callback value %d at index %d did not increase (previous %d)", value, i, previous)
				}
				previous = value
			}
			if len(callbackValues) > 0 && previous != testCase.fileSize {
				t.Errorf("last callback reported %d bytes, expected %d", previous, testCase.fileSize)
			}
		})
	}
}
