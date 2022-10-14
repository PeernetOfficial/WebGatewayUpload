package main

import (
    "bytes"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/PeernetOfficial/core"
    "github.com/PeernetOfficial/core/btcec"
    "github.com/PeernetOfficial/core/webapi"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "io"
    "io/ioutil"
    "mime/multipart"
    "net/http"
    "time"
    "flag"
)

// Variables for the flags to get the address
var (
    BackEndApiAddress *string
    WebpageAddress    *string
    SSL               *bool
)

func init() {
    BackEndApiAddress = flag.String("BackEndApiAddress", "0.0.0.0:8088", "current environment")
    WebpageAddress = flag.String("WebpageAddress", "0.0.0.0:8098", "current environment")
    SSL = flag.Bool("SSL", false, "Flag to check if the SSL certificate is enabled or not")
}

// Initializes Peernet 
func InitPeernet() *core.Backend {
    backend, status, err := core.Init("Your application/1.0", "Config.yaml", nil, nil)
    if status != core.ExitSuccess {
        fmt.Printf("Error %d initializing backend: %s\n", status, err.Error())
        return nil
    }

    return backend
}

// Starts the WebAPI and peernet 
func RunPeernet(backend *core.Backend) {
    webapi.Start(backend, []string{*BackEndApiAddress}, false, "", "", 10*time.Second, 10*time.Second, uuid.Nil)
    backend.Connect()

    for {

    }
}

type WarehouseResult struct {
    Status int    `json:"status"`
    Hash   []byte `json:"hash"`
}

type BlockchainRequest struct {
    Files []File `json:"files"`
}

type File struct {
    Hash []byte `json:"hash"`
    Type int    `json:"type"`
    Name string `json:"name"`
}

func AddFileWarehouse(file io.Reader) *WarehouseResult {
    url := *BackEndApiAddress + "/warehouse/create"

    req, err := http.NewRequest("POST", url, file)
    //req.Header.Set("X-Custom-Header", "myvalue")
    //req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    //fmt.Println("response Status:", resp.Status)
    //fmt.Println("response Headers:", resp.Header)
    body, _ := ioutil.ReadAll(resp.Body)
    //fmt.Println("response Body:", string(body))
    var result WarehouseResult
    err = json.Unmarshal(body, &result)
    if err != nil {
        fmt.Println(err)
    }

    return &result
}

type BlockchainResponse struct {
    Status  int `json:"status"`
    Height  int `json:"height"`
    Version int `json:"version"`
}

// The follwoing function adds the filename and hash to the blockchain  
func AddFileToBlockchain(hash []byte, filename string) *BlockchainResponse {
    url := *BackEndApiAddress + "/blockchain/file/add"

    // Create file object for post
    var blockchainRequest BlockchainRequest
    var files File
    files.Name = filename
    files.Hash = hash
    files.Type = 0
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

    //fmt.Println("response Status:", resp.Status)
    //fmt.Println("response Headers:", resp.Header)
    body, _ := ioutil.ReadAll(resp.Body)
    //fmt.Println("response Body:", string(body))
    var result BlockchainResponse
    err = json.Unmarshal(body, &result)
    if err != nil {
        return nil
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
    fmt.Println(warehouseResult.Hash)
    // current using default port for Peernet api which is 8080
    // First add file to warehouse

    // Adds the file to a blockchain
    Blockchainfo := AddFileToBlockchain(warehouseResult.Hash, header.Filename)
    if Blockchainfo == nil {
        return nil, nil, errors.New("add file to blockchain not successful")
    }

    _, publicKey := backend.ExportPrivateKey()
    fmt.Println(hex.EncodeToString(publicKey.SerializeCompressed()))

    return publicKey, warehouseResult, nil
}

// Add files
func main() {

    // check if SSL is used or not
    if *SSL {
        *BackEndApiAddress = "https://" + *BackEndApiAddress
    } else {
        *BackEndApiAddress = "http://" + *BackEndApiAddress
    }

    // Start peernet
    backend := InitPeernet()
    go RunPeernet(backend)

    r := gin.Default()
    r.LoadHTMLGlob("templates/*.html")
    r.Static("/templates", "./templates")
    r.GET("/upload", func(c *gin.Context) {
        c.HTML(http.StatusOK, "upload2.html", nil)
    })

    r.POST("/uploadFile", func(c *gin.Context) {
        file, header, err := c.Request.FormFile("file")
        defer file.Close()

        if err != nil {
            fmt.Println(err)
        }

        publicKey, warehouseResult, err := UploadFile(backend, &file, header)
        if err != nil {
            fmt.Println(err)
        }

        c.HTML(http.StatusOK, "upload2.html", gin.H{
            "hash":     hex.EncodeToString(warehouseResult.Hash),
            "filename": header.Filename,
            "size":     header.Size,
            "link":     "https://peer.ae/" + hex.EncodeToString(publicKey.SerializeCompressed()) + "/" + hex.EncodeToString(warehouseResult.Hash),
        })

    })

    // Implement CURL script to ensure linux users can upload directly
    // the Cli like https://bashupload.com
    // Ex: curl peer.ae/upload -T your_file.txt

    r.POST("/uploadCurl", func(c *gin.Context) {
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
        //    c.JSON(http.StatusOK, gin.H{
        //    "hash":     hex.EncodeToString(warehouseResult.Hash),
        //    "filename": header.Filename,
        //    "size":     header.Size,
        //    "link":     "https://peer.ae/" + hex.EncodeToString(publicKey.SerializeCompressed()) + "/" + hex.EncodeToString(warehouseResult.Hash),
        //})
    })

    r.Run(*WebpageAddress)
}
