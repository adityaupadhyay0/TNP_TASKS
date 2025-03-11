package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Certificate struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Course     string `json:"course"`
	IssuedTo   string `json:"issued_to"`
	IssueDate  string `json:"issue_date"`
	ExpiryDate string `json:"expiry_date"`
	Issuer     string `json:"issuer"`
	Content    string `json:"content"`
}

var (
	certificates []Certificate
	mutex        sync.Mutex
)

func sendJSONResponse(c *gin.Context, data interface{}, statusCode int) {
	c.JSON(statusCode, data)
}

func getCertificateByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		sendJSONResponse(c, gin.H{"error": "Invalid certificate ID"}, http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	for _, cert := range certificates {
		if cert.ID == id {
			sendJSONResponse(c, cert, http.StatusOK)
			return
		}
	}
	sendJSONResponse(c, gin.H{"error": "Certificate not found"}, http.StatusNotFound)
}

func createCertificate(c *gin.Context) {
	var cert Certificate
	if err := c.ShouldBindJSON(&cert); err != nil {
		sendJSONResponse(c, gin.H{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	cert.ID = len(certificates) + 1
	certificates = append(certificates, cert)

	sendJSONResponse(c, cert, http.StatusCreated)
}

func getAllCertificates(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	sendJSONResponse(c, certificates, http.StatusOK)
}

func updateCertificate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		sendJSONResponse(c, gin.H{"error": "Invalid certificate ID"}, http.StatusBadRequest)
		return
	}

	var updatedCert Certificate
	if err := c.ShouldBindJSON(&updatedCert); err != nil {
		sendJSONResponse(c, gin.H{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	for i, cert := range certificates {
		if cert.ID == id {
			updatedCert.ID = id
			certificates[i] = updatedCert
			sendJSONResponse(c, updatedCert, http.StatusOK)
			return
		}
	}
	sendJSONResponse(c, gin.H{"error": "Certificate not found"}, http.StatusNotFound)
}

func uploadCertificateData(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		sendJSONResponse(c, gin.H{"error": "File upload failed"}, http.StatusBadRequest)
		return
	}

	savePath := "./uploads/" + file.Filename
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		sendJSONResponse(c, gin.H{"error": "Failed to save file"}, http.StatusInternalServerError)
		return
	}

	data, err := readCSVData(savePath)
	if err != nil {
		sendJSONResponse(c, gin.H{"error": "Failed to read CSV file"}, http.StatusInternalServerError)
		return
	}

	filledTemplate, err := fillTemplate(data)
	if err != nil {
		sendJSONResponse(c, gin.H{"error": "Failed to process template"}, http.StatusInternalServerError)
		return
	}

	sendJSONResponse(c, gin.H{"filled_template": filledTemplate}, http.StatusOK)
}

func readCSVData(filePath string) ([]map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least one data row")
	}

	var data []map[string]string

	// Loop through each data row (starting from index 1)
	for _, row := range records[1:] {
		rowMap := make(map[string]string)
		for i, col := range records[0] { // Header row as keys
			if i < len(row) {
				rowMap[col] = row[i]
			}
		}
		data = append(data, rowMap)
	}

	return data, nil
}


func fillTemplates(dataList []map[string]string) ([]string, error) {
	templateFile := "./templates/sample_template.txt"
	content, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("template").Parse(string(content))
	if err != nil {
		return nil, err
	}

	var results []string

	for _, data := range dataList {
		var output strings.Builder
		err = tmpl.Execute(&output, data)
		if err != nil {
			return nil, err
		}
		results = append(results, output.String())
	}

	return results, nil
}

func main() {
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", os.ModePerm)
	}

	r := gin.Default()
	r.POST("/certificates/upload", uploadCertificateData)
	r.GET("/certificates/:id", getCertificateByID)
	r.POST("/certificates", createCertificate)
	r.GET("/certificates", getAllCertificates)
	r.PUT("/certificates/:id", updateCertificate)

	fmt.Println("Server running on port 8080")
	r.Run(":8080")
}
