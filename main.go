package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/durian-client-go/durian"
	"github.com/gorilla/mux"
	capnp "zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/rpc"
)

type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
}

var Blockchain []Block

func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func generateBlock(oldBlock Block, BPM int) (Block, error) {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
}

func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

func replaceChain(newBlocks []Block) {
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
}

func run() error {
	mux := makeMuxRouter()
	httpAddr := os.Getenv("ADDR")
	log.Println("Listening on ", os.Getenv("ADDR"))
	s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	return muxRouter
}

func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(Blockchain, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

type Message struct {
	BPM int
}

func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	var m Message

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()

	newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], m.BPM)
	if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, m)
		return
	}
	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
		newBlockchain := append(Blockchain, newBlock)
		replaceChain(newBlockchain)
		spew.Dump(Blockchain)
	}

	respondWithJSON(w, r, http.StatusCreated, newBlock)

}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

func main() {
	c, err := net.Dial("tcp", "0.0.0.0:3333")
	if err != nil {
		panic(err)
	}

	conn := rpc.NewConn(rpc.StreamTransport(c))
	defer conn.Close()
	ctx := context.Background()

	// Create a connection that we can use to get the HashFactory.
	bz, err := ioutil.ReadFile("token.wasm")
	if err != nil {
		panic(err)
	}

	_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		panic(err)
	}
	tx, err := durian.NewTransaction(seg)
	if err != nil {
		panic(err)
	}
	tx.SetSender([]byte("abcdeabcdeabcdeabcde"))
	tx.SetGas([]byte{1, 2, 1, 1, 1, 1, 1})
	tx.SetGasPrice([]byte{0})
	tx.SetValue([]byte{0})
	tx.SetArgs([]byte("abcdeabcdeabcdeabcdeabcdeabcdeababcdeabcdeabcdeabcdeabcdeabcdeab"))
	action := tx.Action().Create()
	action.SetCode(bz)
	action.SetSalt([]byte("abcdeabcdeabcdeabcdeabcdeabcdeab"))

	//tx.Set
	executor := durian.Executor{Client: conn.Bootstrap(ctx)}
	var provider Provider

	s := executor.Execute(ctx, func(p durian.Executor_execute_Params) error {
		p.SetTransaction(tx)
		p.SetProvider(durian.Provider_ServerToClient(provider))
		return nil
	})

	//s.Client()
	fmt.Println(s)

	// err = capnp.NewEncoder(os.Stdout).Encode(msg)
	// if err != nil {
	// 	panic(err)
	// }
	res, err := s.ResultData().Struct()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Println(res.Data())

	return
}
