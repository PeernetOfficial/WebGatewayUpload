package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/PeernetOfficial/core"
	"github.com/PeernetOfficial/core/btcec"
	"github.com/PeernetOfficial/core/webapi"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	limiter "github.com/julianshen/gin-limiter"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"
)

// Variables for the flags to get the address
var (
	// BackEndApiAddress Refers to the Peernet address ex: <address>:<port no>
	BackEndApiAddress *string
	// WebpageAddress Refers to the address for upload webgate server
	WebpageAddress *string
	// SSL To ensure SSL is required checks if SSL is required and subsequently
	// checks if the certificate is provided
	SSL *bool
	// Certificate SSL Certificate file
	Certificate *string
	// Key SSL Key file
	Key *string
	// BackendAddressWithHTTP ex: http://<address>:<port no>
	BackendAddressWithHTTP string
	// Production mode
	Production *bool
)

//  -------------------------------------------------------------------------------------------------------------
//  ---------------------------------------- Initialize flags and run Peernet  ---------------------------------
//  -------------------------------------------------------------------------------------------------------------

// init reading flags before any part of the code is executed
func init() {
	BackEndApiAddress = flag.String("BackEndApiAddress", "localhost:8088", "current environment")
	WebpageAddress = flag.String("WebpageAddress", "localhost:8098", "current environment")
	SSL = flag.Bool("SSL", false, "Flag to check if the SSL certificate is enabled or not")
	Certificate = flag.String("Certificate", "server.crt", "SSL Certificate file")
	Key = flag.String("Key", "server.key", "SSL Key file")
	Production = flag.Bool("Production", false, "Flag to check if required to run on production mode")
}

// InitPeernet Initializes Peernet backend
func InitPeernet() *core.Backend {
	backend, status, err := core.Init("Peernet Upload Application/1.0", "Config.yaml", nil, nil)
	if status != core.ExitSuccess {
		fmt.Printf("Error %d initializing backend: %s\n", status, err.Error())
		return nil
	}

	return backend
}

// RunPeernet Starts the WebAPI and peernet
func RunPeernet(backend *core.Backend) {
	webapi.Start(backend, []string{*BackEndApiAddress}, false, "", "", 10*time.Second, 10*time.Second, uuid.Nil)
	backend.Connect()

	for {

	}
}

//  -------------------------------------------------------------------------------------------------------------
//  ---------------------- Custom Structs required parse information from the Peernet backend API ---------------
//  -------------------------------------------------------------------------------------------------------------
//  1. Storing file in the warehouse
//  2. Storing file metadata in the blockchain

// -----------------------------------------------------------------------------------------
// ---------------------------------- Warehouse related structs ---------------------------

type WarehouseResult struct {
	Status int    `json:"status"`
	Hash   []byte `json:"hash"`
}

// -----------------------------------------------------------------------------------------
// -------------------------------- Blockchain related structs -----------------------------

// BlockchainRequest blockchain backend API request struct
type BlockchainRequest struct {
	Files []File `json:"files"`
}

type File struct {
	Hash []byte `json:"hash"`
	Type uint16 `json:"type"`
	Name string `json:"name"`
}

// BlockchainResponse blockchain backend API response struct
type BlockchainResponse struct {
	Status  int `json:"status"`
	Height  int `json:"height"`
	Version int `json:"version"`
}

// -----------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------

//  -------------------------------------------------------------------------------------------------------------
//  -------------------------------------- Functions to call Peernet Apis ---------------------------------------
//  -------------------------------------------------------------------------------------------------------------
//  1. Storing file in the warehouse (/warehouse/create)
//  2. Storing file metadata in the blockchain (/blockchain/file/add)

// -----------------------------------------------------------------------------------------
// ---------------------------------- Warehouse related ------------------------------------

// AddFileWarehouse API call for (Storing file in the warehouse)
func AddFileWarehouse(file io.Reader) *WarehouseResult {
	url := BackendAddressWithHTTP + "/warehouse/create"

	req, err := http.NewRequest("POST", url, file)
	//req.Header.Set("X-Custom-Header", "myvalue")
	//req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	var result WarehouseResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println(err)
	}

	return &result
}

// UploadFile Simple abstracted function to add files to peernet core
func UploadFile(backend *core.Backend, file *multipart.File, header *multipart.FileHeader) (*btcec.PublicKey, *WarehouseResult, error) {
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, *file); err != nil {
		return nil, nil, errors.New("io.Copy not successful")
	}

	// adds file to warehouse
	warehouseResult := AddFileWarehouse(buf)
	// current using default port for Peernet api which is 8080
	// First add file to warehouse

	// Adds the file to a blockchain
	Blockchainfo := AddFileToBlockchain(warehouseResult.Hash, header.Filename)
	if Blockchainfo == nil {
		return nil, nil, errors.New("add file to blockchain not successful")
	}

	_, publicKey := backend.ExportPrivateKey()

	return publicKey, warehouseResult, nil
}

// -----------------------------------------------------------------------------------------
// ----------------------------------- Blockchain related  ---------------------------------

// AddFileToBlockchain The follwoing function adds the filename and hash to the blockchain
func AddFileToBlockchain(hash []byte, filename string) *BlockchainResponse {
	url := BackendAddressWithHTTP + "/blockchain/file/add"

	// Get file type
	detectType, _, err := webapi.FileDetectType(filename)
	if err != nil {
		panic(err)
	}

	// Create file object for post
	var blockchainRequest BlockchainRequest
	var files File
	files.Name = filename
	files.Hash = hash
	files.Type = detectType
	blockchainRequest.Files = append(blockchainRequest.Files, files)

	Byte, err := json.Marshal(blockchainRequest)

	// convert bytes
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(Byte))
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var result BlockchainResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil
	}

	return &result
}

// -----------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------

//  -------------------------------------------------------------------------------------------------------------
//  --------------------------------------------- Main function -------------------------------------------------
//  -------------------------------------------------------------------------------------------------------------

func main() {
	// Parsing flags
	flag.Parse()

	// Start peernet
	backend := InitPeernet()
	go RunPeernet(backend)

	var r *gin.Engine
	if *Production {
		gin.SetMode(gin.ReleaseMode)
		r = gin.New()
	} else {
		r = gin.Default()
	}

	r.LoadHTMLGlob("templates/*.html")
	r.Static("/templates", "./templates")

	// --------------------------------- Middleware rate limiter -----------------------------------
	lm := limiter.NewRateLimiter(time.Minute, 10, func(ctx *gin.Context) (string, error) {
		return "", nil
	})
	// ---------------------------------------------------------------------------------------------

	// ---------------------------------------- Routes ---------------------------------------------
	// GET /upload to open upload page from webgateway
	r.GET("/upload", lm.Middleware(), func(c *gin.Context) {
		c.HTML(http.StatusOK, "upload.html", nil)
	})

	// POST /uploadFile Uploads file to peernet from Webgateway
	r.POST("/upload", lm.Middleware(), func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		defer file.Close()

		if err != nil {
			c.HTML(http.StatusBadRequest, "upload.html", gin.H{
				"error": err,
			})
			return
		}

		publicKey, warehouseResult, err := UploadFile(backend, &file, header)
		if err != nil {
			c.HTML(http.StatusBadRequest, "upload.html", gin.H{
				"error": err,
			})
			return
			//fmt.Println(err)
		}

		c.HTML(http.StatusOK, "upload.html", gin.H{
			"hash":     hex.EncodeToString(warehouseResult.Hash),
			"filename": header.Filename,
			"size":     header.Size,
			"link":     "https://peer.ae/" + hex.EncodeToString(publicKey.SerializeCompressed()) + "/" + hex.EncodeToString(warehouseResult.Hash),
			"address":  *WebpageAddress,
		})

	})

	// Implement CURL script to ensure linux users can upload directly
	// the Cli like https://bashupload.com
	// Ex: curl http://localhost:8080/uploadCurl -F add=@<file name>
	r.POST("/uploadCurl", lm.Middleware(), func(c *gin.Context) {
		file, header, err := c.Request.FormFile("add")
		defer file.Close()

		if err != nil {
			fmt.Println(err)
		}

		publicKey, warehouseResult, err := UploadFile(backend, &file, header)
		if err != nil {
			fmt.Println(err)
		}

		link := "https://peer.ae/" + hex.EncodeToString(publicKey.SerializeCompressed()) + "/" + hex.EncodeToString(warehouseResult.Hash)
		c.Data(http.StatusOK, "plain/text", []byte(link))
	})

	// ---------------------------------------------------------------------------------------------

	// ---------------------------------- Start Gin server -----------------------------------------
	// check if SSL is used or not
	if *SSL {
		BackendAddressWithHTTP = "https://" + *BackEndApiAddress
		r.RunTLS(*WebpageAddress, *Certificate, *Key)
		*WebpageAddress = "https://" + *WebpageAddress
	} else {
		BackendAddressWithHTTP = "http://" + *BackEndApiAddress
		r.Run(*WebpageAddress)
		*WebpageAddress = "http://" + *WebpageAddress
	}
	// ---------------------------------------------------------------------------------------------
}
