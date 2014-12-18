package functionaltests

import (
	"fmt"
	"testing"
	"time"
	"github.com/couchbase/indexing/secondary/tests/framework/datautility"
	"github.com/couchbase/indexing/secondary/tests/framework/kvutility"
	"github.com/couchbase/indexing/secondary/tests/framework/secondaryindex"
	// tc "github.com/couchbase/indexing/secondary/tests/framework/common"
	tv "github.com/couchbase/indexing/secondary/tests/framework/validation"
)

var docs []kvutility.KeyValue
var defaultlimit int64 = 10000000

func init() {
	fmt.Println("In init()")
	secondaryindex.DropAllSecondaryIndexes()
	
	// Working with Users10k dataset.
	dataFilePath := "../testdata/Users10k.txt.gz"
	// tc.DownloadDataFile(tc.IndexTypesStaticJSONData, dataFilePath)
	keyValues := datautility.LoadJSONFromCompressedFile(dataFilePath, "docid")
	kvutility.SetKeyValues(keyValues, "default", "", "127.0.0.1")
	docs = keyValues
}

func TestDummy(t *testing.T) {
	fmt.Println("In TestDummy()")
}

// Test for single index field of data type float64
func TestSimpleIndex_FloatDataType(t *testing.T) {
	fmt.Println("In TestSimpleIndex_FloatDataType()")
	var indexName = "index_age"
	var bucketName = "default"
	
	err := secondaryindex.CreateSecondaryIndex(indexName, bucketName, []string {"age"}, true)
	FailTestIfError(err, "Error in creating the index", t)
	
	// Wait, else results in "Index not ready"
	time.Sleep(1 * time.Second)
	
	docScanResults := datautility.ExpectedScanResponse_float64(docs, "age", 35, 40, 1)
	scanResults, err := secondaryindex.Range(indexName, bucketName, []interface{} {35}, []interface{} {40}, 1, true, defaultlimit)
	FailTestIfError(err, "Error in scan", t)
	tv.Validate(docScanResults, scanResults)
}

// Test for single index field of data type string
func TestSimpleIndex_StringDataType(t *testing.T) {
	fmt.Println("In TestSimpleIndex_StringDataType()")
	var indexName = "index_company"
	var bucketName = "default"
	
	err := secondaryindex.CreateSecondaryIndex(indexName, bucketName, []string {"company"}, true)
	FailTestIfError(err, "Error in creating the index", t)
	time.Sleep(1 * time.Second)
	
	docScanResults := datautility.ExpectedScanResponse_string(docs, "company", "G", "M", 1)
	scanResults, err := secondaryindex.Range(indexName, bucketName, []interface{} {"G"}, []interface{} {"M"}, 1, true, defaultlimit)
	FailTestIfError(err, "Error in scan 1", t)
	tv.Validate(docScanResults, scanResults)
	
	docScanResults = datautility.ExpectedScanResponse_string(docs, "company", "BIOSPAN", "ZILLANET", 1)
	scanResults, err = secondaryindex.Range(indexName, bucketName, []interface{} {"BIOSPAN"}, []interface{} {"ZILLANET"}, 1, true, defaultlimit)
	FailTestIfError(err, "Error in scan 2", t)
	tv.Validate(docScanResults, scanResults)
}

// Test for case sensitivity of index values
func TestSimpleIndex_StringCaseSensitivity(t *testing.T) {
	fmt.Println("In TestSimpleIndex_StringCaseSensitivity()")
	var indexName = "index_company"
	var bucketName = "default"
	
	err := secondaryindex.CreateSecondaryIndex(indexName, bucketName, []string {"company"}, true)
	FailTestIfError(err, "Error in creating the index", t)
	time.Sleep(1 * time.Second)
	
	docScanResults := datautility.ExpectedScanResponse_string(docs, "company", "B", "C", 1)
	scanResults, err := secondaryindex.Range(indexName, bucketName, []interface{} {"B"}, []interface{} {"C"}, 1, true, defaultlimit)
	FailTestIfError(err, "Error in scan 1", t)
	tv.Validate(docScanResults, scanResults)
	
	docScanResults = datautility.ExpectedScanResponse_string(docs, "company", "B", "c", 1)
	scanResults, err = secondaryindex.Range(indexName, bucketName, []interface{} {"B"}, []interface{} {"c"}, 1, true, defaultlimit)
	FailTestIfError(err, "Error in scan 2", t)
	tv.Validate(docScanResults, scanResults)
}

// Test for single index field of data type bool
func TestSimpleIndex_BoolDataType(t *testing.T) {
	fmt.Println("In TestSimpleIndex_BoolDataType()")
	var indexName = "index_isActive"
	var bucketName = "default"
	
	err := secondaryindex.CreateSecondaryIndex(indexName, bucketName, []string {"isActive"}, true)
	FailTestIfError(err, "Error in creating the index", t)
	
	time.Sleep(1 * time.Second)
	
	docScanResults := datautility.ExpectedScanResponse_bool(docs, "isActive", true, 3)
	scanResults, err:= secondaryindex.Range(indexName, bucketName, []interface{} { true }, []interface{} { true }, 3, true, defaultlimit)
	FailTestIfError(err, "Error in scan", t)
	tv.Validate(docScanResults, scanResults)
}

func TestBasicLookup(t *testing.T) {
	fmt.Println("In TestBasicLookup()")
	var indexName = "index_company"
	var bucketName = "default"
	
	err := secondaryindex.CreateSecondaryIndex(indexName, bucketName, []string {"company"}, true)
	FailTestIfError(err, "Error in creating the index", t)
	time.Sleep(1 * time.Second)
	
	docScanResults := datautility.ExpectedScanResponse_string(docs, "company", "BIOSPAN", "BIOSPAN", 3)
	scanResults, err := secondaryindex.Lookup(indexName, bucketName, []interface{} {"BIOSPAN"}, true, 10000000)
	FailTestIfError(err, "Error in scan", t)
	tv.Validate(docScanResults, scanResults)
}

func TestIndexOnNonExistentField(t *testing.T) {
	fmt.Println("In TestIndexOnNonExistentField()")
	var indexName = "index_height"
	var bucketName = "default"
	
	err := secondaryindex.CreateSecondaryIndex(indexName, bucketName, []string {"height"}, true)
	FailTestIfError(err, "Error in creating the index", t)
	time.Sleep(1 * time.Second)
	
	docScanResults := datautility.ExpectedScanResponse_float64(docs, "height", 6.0, 6.5, 1)
	scanResults, err := secondaryindex.Range(indexName, bucketName, []interface{} { 6.0 }, []interface{} { 6.5 }, 1, true, defaultlimit)
	FailTestIfError(err, "Error in scan", t)
	tv.Validate(docScanResults, scanResults)
}

func TestIndexPartiallyMissingField(t *testing.T) {
	fmt.Println("In TestIndexPartiallyMissingField()")
	var indexName = "index_nationality"
	var bucketName = "default"
	
	err := secondaryindex.CreateSecondaryIndex(indexName, bucketName, []string {"nationality"}, true)
	FailTestIfError(err, "Error in creating the index", t)
	time.Sleep(5 * time.Second)
	
	docScanResults := datautility.ExpectedScanResponse_string(docs, "nationality", "A", "z", 1)
	scanResults, err := secondaryindex.Range(indexName, bucketName, []interface{} { "A" }, []interface{} { "z" }, 1, true, defaultlimit)
	FailTestIfError(err, "Error in scan", t)
	tv.Validate(docScanResults, scanResults)
}

// Index field is float but scan for string
func TestScanNonMatchingDatatype(t *testing.T) {
	fmt.Println("In TestScanNonMatchingDatatype()")
	var indexName = "index_age"
	var bucketName = "default"
	
	err := secondaryindex.CreateSecondaryIndex(indexName, bucketName, []string {"age"}, true)
	FailTestIfError(err, "Error in creating the index", t)
	
	// Wait, else results in "Index not ready"
	time.Sleep(5 * time.Second)
	
	docScanResults := datautility.ExpectedScanResponse_string(docs, "age", "35", "40", 1)
	scanResults, err := secondaryindex.Range(indexName, bucketName, []interface{} { "35" }, []interface{} { "40" }, 1, true, defaultlimit)
	FailTestIfError(err, "Error in scan", t)
	tv.Validate(docScanResults, scanResults)
}

// Inclusion tests

// Inclusion 0
func TestInclusionNeither(t *testing.T) {
	fmt.Println("In TestInclusionNeither()")
	var indexName = "index_age"
	var bucketName = "default"
	
	err := secondaryindex.CreateSecondaryIndex(indexName, bucketName, []string {"age"}, true)
	FailTestIfError(err, "Error in creating the index", t)
	
	// Wait, else results in "Index not ready"
	time.Sleep(3 * time.Second)
	
	docScanResults := datautility.ExpectedScanResponse_float64(docs, "age", 32, 36, 0)
	scanResults, err := secondaryindex.Range(indexName, bucketName, []interface{} { 32 }, []interface{} { 36 }, 0, true, defaultlimit)
	FailTestIfError(err, "Error in scan", t)
	tv.Validate(docScanResults, scanResults)
}

// Inclusion 1
func TestInclusionLow(t *testing.T) {
	fmt.Println("In TestInclusionLow()")
	var indexName = "index_age"
	var bucketName = "default"
	
	err := secondaryindex.CreateSecondaryIndex(indexName, bucketName, []string {"age"}, true)
	FailTestIfError(err, "Error in creating the index", t)
	
	// Wait, else results in "Index not ready"
	time.Sleep(3 * time.Second)
	
	docScanResults := datautility.ExpectedScanResponse_float64(docs, "age", 32, 36, 1)
	scanResults, err := secondaryindex.Range(indexName, bucketName, []interface{} { 32 }, []interface{} { 36 }, 1, true, defaultlimit)
	FailTestIfError(err, "Error in scan", t)
	tv.Validate(docScanResults, scanResults)
}

// Inclusion 2
func TestInclusionHigh(t *testing.T) {
	fmt.Println("In TestInclusionHigh()")
	var indexName = "index_age"
	var bucketName = "default"
	
	err := secondaryindex.CreateSecondaryIndex(indexName, bucketName, []string {"age"}, true)
	FailTestIfError(err, "Error in creating the index", t)
	
	// Wait, else results in "Index not ready"
	time.Sleep(3 * time.Second)
	
	docScanResults := datautility.ExpectedScanResponse_float64(docs, "age", 32, 36, 2)
	scanResults, err := secondaryindex.Range(indexName, bucketName, []interface{} { 32 }, []interface{} { 36 }, 2, true, defaultlimit)
	FailTestIfError(err, "Error in scan", t)
	tv.Validate(docScanResults, scanResults)
}

// Inclusion 3
func TestInclusionBoth(t *testing.T) {
	fmt.Println("In TestInclusionBoth()")
	var indexName = "index_age"
	var bucketName = "default"
	
	err := secondaryindex.CreateSecondaryIndex(indexName, bucketName, []string {"age"}, true)
	FailTestIfError(err, "Error in creating the index", t)
	
	// Wait, else results in "Index not ready"
	time.Sleep(3 * time.Second)
	
	docScanResults := datautility.ExpectedScanResponse_float64(docs, "age", 32, 36, 3)
	scanResults, err := secondaryindex.Range(indexName, bucketName, []interface{} { 32 }, []interface{} { 36 }, 3, true, defaultlimit)
	FailTestIfError(err, "Error in scan", t)
	tv.Validate(docScanResults, scanResults)
}

func TestNestedIndex_String(t *testing.T) {
	fmt.Println("In TestNestedIndex_String()")
	var indexName = "index_streetname"
	var bucketName = "default"
	
	err := secondaryindex.CreateSecondaryIndex(indexName, bucketName, []string {"address.streetaddress.streetname"}, true)
	FailTestIfError(err, "Error in creating the index", t)
	
	// Wait, else results in "Index not ready"
	time.Sleep(3 * time.Second)
	
	docScanResults := datautility.ExpectedScanResponse_string(docs, "address.streetaddress.streetname", "A", "z", 3)
	scanResults, err := secondaryindex.Range(indexName, bucketName, []interface{} { "A" }, []interface{} { "z" }, 3, true, defaultlimit)
	FailTestIfError(err, "Error in scan", t)
	tv.Validate(docScanResults, scanResults)
}

func TestNestedIndex_Float(t *testing.T) {
	fmt.Println("In TestNestedIndex_Float()")
	var indexName = "index_floor"
	var bucketName = "default"
	
	err := secondaryindex.CreateSecondaryIndex(indexName, bucketName, []string {"address.streetaddress.floor"}, true)
	FailTestIfError(err, "Error in creating the index", t)
	
	// Wait, else results in "Index not ready"
	time.Sleep(3 * time.Second)
	
	docScanResults := datautility.ExpectedScanResponse_float64(docs, "address.streetaddress.floor", 3, 6, 3)
	scanResults, err := secondaryindex.Range(indexName, bucketName, []interface{} { 3 }, []interface{} { 6 }, 3, true, defaultlimit)
	FailTestIfError(err, "Error in scan", t)
	tv.Validate(docScanResults, scanResults)
}

func TestNestedIndex_Bool(t *testing.T) {
	fmt.Println("In TestNestedIndex_Bool()")
	var indexName = "index_isresidential"
	var bucketName = "default"
	
	err := secondaryindex.CreateSecondaryIndex(indexName, bucketName, []string {"address.isresidential"}, true)
	FailTestIfError(err, "Error in creating the index", t)
	
	// Wait, else results in "Index not ready"
	time.Sleep(3 * time.Second)
	
	docScanResults := datautility.ExpectedScanResponse_bool(docs, "address.isresidential", false, 3)
	scanResults, err := secondaryindex.Range(indexName, bucketName, []interface{} { false }, []interface{} { false }, 3, true, defaultlimit)
	FailTestIfError(err, "Error in scan", t)
	tv.Validate(docScanResults, scanResults)
}

func FailTestIfError(err error, msg string, t *testing.T) {
	if err != nil {
		t.Fatal("%v: %v\n", msg, err)
	}
}

