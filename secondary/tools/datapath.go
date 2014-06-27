package main

import (
	"code.google.com/p/goprotobuf/proto"
	ap "github.com/couchbase/indexing/secondary/adminport"
	c "github.com/couchbase/indexing/secondary/common"
	"github.com/couchbase/indexing/secondary/indexer"
	"github.com/couchbase/indexing/secondary/projector"
	"github.com/couchbase/indexing/secondary/protobuf"
	"log"
	"time"
)

var cluster = "localhost:9000"
var pooln = "default"
var bucketn = "beer-sample"
var kvaddrs = []string{"127.0.0.1:12000"}
var adminport = "localhost:9999"
var endpoint = "localhost:9998"
var coordEndpoint = "localhost:9997"
var vbnos = []uint16{0, 1, 2, 3, 4, 5, 6, 7}
var done = make(chan bool)

var instances = []uint64{0x11, 0x12}

func main() {
	c.SetLogLevel(c.LogLevelTrace)
	projector.NewProjector(cluster, kvaddrs, adminport)
	aport := ap.NewHTTPClient("http://"+adminport, "/adminport/")
	fReq := protobuf.FailoverLogRequest{
		Pool:   proto.String(pooln),
		Bucket: proto.String(bucketn),
		Vbnos:  c.Vbno16to32(vbnos),
	}
	fRes := protobuf.FailoverLogResponse{}
	if err := aport.Request(&fReq, &fRes); err != nil {
		log.Fatal(err)
	}
	vbuuids := make([]uint64, 0)
	for _, flog := range fRes.GetLogs() {
		vbuuids = append(vbuuids, flog.Vbuuids[len(flog.Vbuuids)-1])
	}

	go endpointServer(endpoint)
	go endpointServer(coordEndpoint)
	time.Sleep(100 * time.Millisecond)

	mReq := makeStartRequest(vbuuids)
	mRes := protobuf.MutationStreamResponse{}
	if err := aport.Request(mReq, &mRes); err != nil {
		log.Fatal(err)
	}
	<-done
	<-done
}

func makeStartRequest(vbuuids []uint64) *protobuf.MutationStreamRequest {
	bTs := &protobuf.BranchTimestamp{
		Bucket:  proto.String(bucketn),
		Vbnos:   []uint32{0, 1, 2, 3, 4, 5, 6, 7},
		Seqnos:  []uint64{0, 0, 0, 0, 0, 0, 0, 0},
		Vbuuids: vbuuids,
	}
	req := protobuf.MutationStreamRequest{
		Topic:             proto.String("maintanence"),
		Pools:             []string{pooln},
		Buckets:           []string{bucketn},
		RestartTimestamps: []*protobuf.BranchTimestamp{bTs},
		Instances:         makeIndexInstances(),
	}
	req.SetStartFlag()
	return &req
}

func makeIndexInstances() []*protobuf.IndexInst {
	sExprs := []string{`{"type":"property","path":"name"}`,
		`{"type":"property","path":"abv"}`}
	defn1 := &protobuf.IndexDefn{
		DefnID:          proto.Uint64(instances[0]),
		Bucket:          proto.String(bucketn),
		IsPrimary:       proto.Bool(false),
		Name:            proto.String("index1"),
		Using:           protobuf.StorageType_View.Enum(),
		ExprType:        protobuf.ExprType_N1QL.Enum(),
		SecExpressions:  sExprs,
		PartitionScheme: protobuf.PartitionScheme_TEST.Enum(),
		PartnExpression: proto.String(`{"type":"property","path":"type"}`),
	}
	instance1 := &protobuf.IndexInst{
		InstId:     proto.Uint64(0x1),
		State:      protobuf.IndexState_IndexInitial.Enum(),
		Definition: defn1,
		Tp: &protobuf.TestPartition{
			CoordEndpoint: proto.String(coordEndpoint),
			Endpoints:     []string{endpoint},
		},
	}

	defn2 := &protobuf.IndexDefn{
		DefnID:          proto.Uint64(instances[1]),
		Bucket:          proto.String(bucketn),
		IsPrimary:       proto.Bool(false),
		Name:            proto.String("index2"),
		Using:           protobuf.StorageType_View.Enum(),
		ExprType:        protobuf.ExprType_N1QL.Enum(),
		SecExpressions:  []string{`{"type":"property","path":"city"}`},
		PartitionScheme: protobuf.PartitionScheme_TEST.Enum(),
		PartnExpression: proto.String(`{"type":"property","path":"type"}`),
	}
	instance2 := &protobuf.IndexInst{
		InstId:     proto.Uint64(0x2),
		State:      protobuf.IndexState_IndexInitial.Enum(),
		Definition: defn2,
		Tp: &protobuf.TestPartition{
			CoordEndpoint: proto.String(coordEndpoint),
			Endpoints:     []string{endpoint},
		},
	}
	return []*protobuf.IndexInst{instance1, instance2}
}

func endpointServer(addr string) {
	mutChanSize := 100
	mutch := make(chan []*protobuf.VbKeyVersions, mutChanSize)
	sbch := make(chan interface{}, 100)
	_, err := indexer.NewMutationStream(addr, mutch, sbch)
	if err != nil {
		log.Fatal(err)
	}
	mutations, messages := 0, 0
	commandWise := make(map[byte]int)
	keys := map[uint64]map[string][]string{
		0x11: make(map[string][]string),
		0x12: make(map[string][]string),
	}

	printTm := time.Tick(100 * time.Millisecond)

loop:
	for {
		select {
		case vbs, ok := <-mutch:
			if ok {
				mutations += gatherKeys(vbs, commandWise, keys)
			} else {
				break loop
			}
		case s, ok := <-sbch:
			if ok {
				switch v := s.(type) {
				case []*indexer.RestartVbuckets:
					printRestartVbuckets(addr, v)
				}
				messages++
			} else {
				break loop
			}
		case <-time.After(4 * time.Second):
			break loop
		case <-printTm:
			log.Println(addr, "-- mutations", mutations)
			log.Println(addr, "-- commandWise", commandWise)
		}
	}

	log.Println(addr, "-- mutations", mutations, "-- messages", messages)
	log.Println(addr, "-- commandWise", commandWise)
	ks, ds := countKeysAndDocs(keys[0x11])
	log.Printf("%v -- for instance 0x11, %v unique keys found in %v docs\n", addr, ks, ds)
	ks, ds = countKeysAndDocs(keys[0x12])
	log.Printf("%v -- for instance 0x12, %v unique keys found in %v docs\n", addr, ks, ds)
	done <- true
}

func printRestartVbuckets(addr string, rs []*indexer.RestartVbuckets) {
	for _, r := range rs {
		log.Printf("restart: %s, %v %v\n", addr, r.Bucket, r.Vbuckets)
	}
}

func gatherKeys(
	vbs []*protobuf.VbKeyVersions,
	commandWise map[byte]int,
	keys map[uint64]map[string][]string,
) int {

	mutations := 0
	for _, vb := range vbs {
		kvs := vb.GetKvs()
		for _, kv := range kvs {
			mutations++
			docid := string(kv.GetDocid())
			uuids, seckeys := kv.GetUuids(), kv.GetKeys()
			for i, command := range kv.GetCommands() {
				cmd := byte(command)
				if _, ok := commandWise[cmd]; !ok {
					commandWise[cmd] = 0
				}
				commandWise[cmd]++

				if cmd == c.StreamBegin {
					continue
				}

				if cmd == c.Upsert {
					uuid := uuids[i]
					key := string(seckeys[i])
					if _, ok := keys[uuid][key]; !ok {
						keys[uuid][key] = make([]string, 0)
					}
					keys[uuid][key] = append(keys[uuid][key], docid)
				}
			}
		}
	}
	return mutations
}

func countKeysAndDocs(keys map[string][]string) (int, int) {
	countKs, countDs := 0, 0
	for _, docs := range keys {
		countKs++
		countDs += len(docs)
	}
	return countKs, countDs
}