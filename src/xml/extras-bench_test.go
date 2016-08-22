// +build !race

package xml

import (
	"brainloop/pe/util/test/random"
	"encoding/base64"
	"fmt"
	"io"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCharDataReaderStreamed(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	const testDataSeed int64 = 42
	const testDataSize int64 = 40 * 1024 * 1024

	// get memory stats before performing the test
	runtime.GC()
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	r, w := io.Pipe()

	// produce a large chunk of XML
	go func(w io.WriteCloser) {
		fmt.Fprint(w, "<Message>")
		w64 := base64.NewEncoder(base64.StdEncoding, w)
		_, err := io.Copy(w64, random.NewBuf(testDataSeed, testDataSize))
		require.NoError(err)
		require.NoError(w64.Close())
		fmt.Fprint(w, "</Message>")
		require.NoError(w.Close())
	}(w)

	// decode the large chunk of XML
	d := NewDecoder(r)

	tok, err := d.Token() // start element
	require.NoError(err)
	assert.Equal(StartElement{Name: Name{Local: "Message"}, Attr: []Attr{}}, tok)

	r64 := base64.NewDecoder(base64.StdEncoding, d.CharDataReader())
	// note: the copy would fail if the output data does not match the input data
	_, err = io.Copy(random.NewBuf(testDataSeed, testDataSize), r64)
	require.NoError(err)

	tok, err = d.Token() // end element
	require.NoError(err)
	assert.Equal(EndElement{Name: Name{Local: "Message"}}, tok)

	// compare current memory stats with stats from before the test.
	// we expect a *LOT* less allocations than the amount of transferred data
	runtime.GC()
	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)
	bytesAllocated := memStatsAfter.TotalAlloc - memStatsBefore.TotalAlloc
	assert.InDelta(100*1024, bytesAllocated, 50*1024, "Allocated between 50 and 150KB")
}
