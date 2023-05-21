package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/Akilan1999/p2p-rendering-computation/p2p/frp"
	"github.com/PeernetOfficial/core"
	"github.com/PeernetOfficial/core/btcec"
	"github.com/PeernetOfficial/core/webapi"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	limiter "github.com/julianshen/gin-limiter"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
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
	// P2PRC mode
	P2PRC *bool
	// P2PRCRootNode Root node of P2PRC
	P2PRCRootNode *string
	// P2PRCExposePort Public exposed port
	P2PRCExposePort *string
)

//  -------------------------------------------------------------------------------------------------------------
//  ---------------------------------------- Initialize flags and run Peernet  ---------------------------------
//  -------------------------------------------------------------------------------------------------------------

// init reading flags before any part of the code is executed
func init() {
	BackEndApiAddress = flag.String("BackEndApiAddress", "localhost:8081", "current environment")
	WebpageAddress = flag.String("WebpageAddress", "localhost:8098", "current environment")
	SSL = flag.Bool("SSL", false, "Flag to check if the SSL certificate is enabled or not")
	Certificate = flag.String("Certificate", "server.crt", "SSL Certificate file")
	Key = flag.String("Key", "server.key", "SSL Key file")
	Production = flag.Bool("Production", false, "Flag to check if required to run on production mode")
	P2PRC = flag.Bool("P2PRC", false, "Run P2PRC mode to selfnode node")
	P2PRCRootNode = flag.String("P2PRCHost", "", "Run P2PRC mode to selfnode node")
	P2PRCExposePort = flag.String("P2PRCExposedPort", "", "Port exposed externally for P2PRC")
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
	webapi.Start(backend, []string{*BackEndApiAddress}, false, "", "", 200000*time.Second, 200000*time.Second, uuid.Nil)
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
//func AddFileWarehouse(file io.Reader, id string) *WarehouseResult {
//    route := BackendAddressWithHTTP + "/warehouse/create"
//
//    form := url.Values{}
//    form.Add("file", file)
//    form.Add("ip", c.ip)
//    form.Add("ua", c.ua)
//    req.PostForm = form
//
//    req, err := http.NewRequest("POST", route, file)
//    //req.Header.Set("X-Custom-Header", "myvalue")
//    req.Header.Set("Content-Type", "multipart/form-data")
//
//    client := &http.Client{}
//    resp, err := client.Do(req)
//    if err != nil {
//        panic(err)
//    }
//    defer resp.Body.Close()
//
//    body, err := ioutil.ReadAll(resp.Body)
//    if err != nil {
//        fmt.Println(err)
//    }
//    var result WarehouseResult
//    err = json.Unmarshal(body, &result)
//    if err != nil {
//        fmt.Println(err)
//    }
//
//    return &result
//}

func AddFileWarehouse(File io.Reader, uuid string) (resp *WarehouseResult, err error) {
	//// Prepare a form that you will submit to that URL.
	//var b bytes.Buffer
	//w := multipart.NewWriter(&b)
	//for key, r := range values {
	//	var fw io.Writer
	//	if x, ok := r.(io.Closer); ok {
	//		defer x.Close()
	//	}
	//	// Add an image file
	//	if x, ok := r.(*os.File); ok {
	//		if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
	//			return
	//		}
	//	} else {
	//		// Add other fields
	//		if fw, err = w.CreateFormField(key); err != nil {
	//			return
	//		}
	//	}
	//	if _, err = io.Copy(fw, r); err != nil {
	//		return nil, err
	//	}
	//
	//}
	//// Don't forget to close the multipart writer.
	//// If you don't close it, your request will be missing the terminating boundary.
	//w.Close()
	//
	//// Now that you have a form, you can submit it to your handler.
	//req, err := http.NewRequest("POST", BackendAddressWithHTTP+"/warehouse/create", &b)
	//if err != nil {
	//	return
	//}
	//// Don't forget to set the content type, this will contain the boundary.
	//req.Header.Set("Content-Type", w.FormDataContentType())
	//
	//client := &http.Client{}
	//
	//// Submit the request
	//res, err := client.Do(req)
	//if err != nil {
	//	return
	//}
	//
	//// Check the response
	//if res.StatusCode != http.StatusOK {
	//	err = fmt.Errorf("bad status: %s", res.Status)
	//}
	//
	//body, err := ioutil.ReadAll(res.Body)
	//if err != nil {
	//	return
	//}
	//
	//err = json.Unmarshal(body, &resp)
	//if err != nil {
	//	return
	//}
	//
	//fmt.Println(resp)
	//
	//return

	url := BackendAddressWithHTTP + "/warehouse/create"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("id", uuid)
	part2, errFile2 := writer.CreateFormFile("File", "test")
	_, errFile2 = io.Copy(part2, File)
	if errFile2 != nil {
		fmt.Println(errFile2)
		return
	}
	err = writer.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	return
}

// UploadFile Simple abstracted function to add files to peernet core
func UploadFile(backend *core.Backend, file *multipart.File, header *multipart.FileHeader, uuid string) (*btcec.PublicKey, *WarehouseResult, error) {
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, *file); err != nil {
		return nil, nil, errors.New("io.Copy not successful")
	}

	// adds file to warehouse
	warehouseResult, err := AddFileWarehouse(buf, uuid)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil, err
	}
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

	fmt.Println(result)

	return &result
}

// ---------------------------------------------------------------------------------------------
// ----------------------------------- Get status upload file  ---------------------------------

func GetStatusUploadFile(uuid string) (uploadStatus *webapi.UploadStatus) {
	url := BackendAddressWithHTTP + "/create/track/uploadID?id=" + uuid

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	err = json.Unmarshal(body, uploadStatus)
	if err != nil {
		return nil
	}

	fmt.Println(uploadStatus.Progress.Percentage)

	return
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

	r.Use(CORSMiddleware())

	r.LoadHTMLGlob("templates/*.html")
	r.Static("/templates", "./templates")

	// --------------------------------- Middleware rate limiter -----------------------------------
	lm := limiter.NewRateLimiter(time.Minute, 10, func(ctx *gin.Context) (string, error) {
		return "", nil
	})
	// ---------------------------------------------------------------------------------------------

	// ---------------------------------------- Routes ---------------------------------------------
	// GET /upload to open upload page from webgateway
	r.GET("/", lm.Middleware(), func(c *gin.Context) {
		c.HTML(http.StatusOK, "upload.html", gin.H{"backendUrl": *BackEndApiAddress})
	})

	// POST /uploadFile Uploads file to peernet from Webgateway
	r.POST("/upload", lm.Middleware(), func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		defer file.Close()

		uuid := c.Request.FormValue("uuid")

		if err != nil {
			c.HTML(http.StatusBadRequest, "upload.html", gin.H{
				"error": err,
			})
			return
		}

		publicKey, warehouseResult, err := UploadFile(backend, &file, header, uuid)
		if err != nil {
			c.HTML(http.StatusBadRequest, "upload.html", gin.H{
				"backendUrl": *BackEndApiAddress,
				"error":      err,
			})
			return
			//fmt.Println(err)
		}

		c.JSON(http.StatusOK, gin.H{
			"hash":       hex.EncodeToString(warehouseResult.Hash),
			"filename":   header.Filename,
			"size":       header.Size,
			"link":       "http://164.90.177.167:8889/" + hex.EncodeToString(publicKey.SerializeCompressed()) + "/" + hex.EncodeToString(warehouseResult.Hash) + "/?filename=" + header.Filename,
			"address":    *WebpageAddress,
			"backendUrl": *BackEndApiAddress,
		})

	})

	r.GET("/uploadStatus", lm.Middleware(), func(c *gin.Context) {

		uuid := c.Query("uuid")

		c.JSON(http.StatusOK, GetStatusUploadFile(uuid))

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

		publicKey, warehouseResult, err := UploadFile(backend, &file, header, "")
		if err != nil {
			fmt.Println(err)
		}

		link := "http://peer.ae/" + hex.EncodeToString(publicKey.SerializeCompressed()) + "/" + hex.EncodeToString(warehouseResult.Hash)
		c.Data(http.StatusOK, "plain/text", []byte(link))
	})

	// ---------------------------------------------------------------------------------------------

	// ----------------------- Check if P2PRC mode is selected for hosting ------------------------
	if *P2PRC && *P2PRCRootNode != "" && *P2PRCExposePort != "" {
		EscapeNATWebGateway()
	}

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

func EscapeNATWebGateway() (ExposePortP2PRC string, err error) {
	host, port, err := net.SplitHostPort(*P2PRCRootNode)
	if err != nil {
		return
	}

	serverPort, err := frp.GetFRPServerPort("http://" + host + ":" + port)

	if err != nil {
		return
	}

	time.Sleep(1 * time.Second)

	_, port, err = net.SplitHostPort(*WebpageAddress)
	if err != nil {
		return
	}

	//port for the barrierKVM server
	ExposePortP2PRC, err = frp.StartFRPClientForServer(host, serverPort, port, *P2PRCExposePort)
	if err != nil {
		return
	}

	return
}

// CORSMiddleware Use Cors middleware
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
