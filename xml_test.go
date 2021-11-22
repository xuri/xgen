package xgen

import (
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	schema "github.com/xuri/xgen/test/go"
	"io/ioutil"
	"path/filepath"
	"testing"
)

// TestGeneratedGo runs through test cases to validate Go generated structs. Each test case
// requires a xml fixture file to unmarshal into the receiving struct. Validate first validates
// that the file can be unmarshaled as the receiving struct and then remarshals the content
// to make sure the marshaling is symmetrical
func TestGeneratedGo(t *testing.T) {
	testCases := []struct {
		// xmlFileName is the path to the xml fixture file to unmarshal into the receiving struct
		xmlFileName     string
		// receivingStruct is a pointer to the struct to unmarshal the xml file content into. It should match
		// the type of the top level element present in that file
		receivingStruct interface{}
	}{
		{
			xmlFileName:     "base64.xml",
			receivingStruct: &schema.TopLevel{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.xmlFileName, func(t *testing.T) {
			fullPath := filepath.Join("xmlFixtures", tc.xmlFileName)

			input, err := ioutil.ReadFile(fullPath)
			require.NoError(t, err)

			err = xml.Unmarshal(input, tc.receivingStruct)
			require.NoError(t, err)

			// Validate that decoding resulted in a non-zero value
			assert.NotEmpty(t, tc.receivingStruct)

			// Remarshal the parsed content to compare it with the original and make sure that the parsing/encoding
			// is symmetrical
			remarshaled, err := xml.MarshalIndent(tc.receivingStruct, "", "    ")
			require.NoError(t, err)

			XMLEqual(t, input, remarshaled)
		})
	}
}

// XMLEqual checks that two inputs of raw XML represent the same logical data disregarding formatting differences
func XMLEqual(t *testing.T, expected []byte, actual []byte) {
	var parsedRaw interface{}
	err := xml.Unmarshal(expected, &parsedRaw)
	require.NoError(t, err)

	var parsedRemarshaled interface{}
	err = xml.Unmarshal(actual, &parsedRemarshaled)
	require.NoError(t, err)

	assert.Equal(t, parsedRaw, parsedRemarshaled)
}
