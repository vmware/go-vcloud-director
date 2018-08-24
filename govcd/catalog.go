/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"bytes"
	"encoding/xml"
	"errors"
	"github.com/vmware/go-vcloud-director/types/v56"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Catalog struct {
	Catalog *types.Catalog
	c       *Client
}

func NewCatalog(c *Client) *Catalog {
	return &Catalog{
		Catalog: new(types.Catalog),
		c:       c,
	}
}

func (c *Catalog) FindCatalogItem(catalogitem string) (CatalogItem, error) {

	for _, cis := range c.Catalog.CatalogItems {
		for _, ci := range cis.CatalogItem {
			if ci.Name == catalogitem && ci.Type == "application/vnd.vmware.vcloud.catalogItem+xml" {
				u, err := url.ParseRequestURI(ci.HREF)

				if err != nil {
					return CatalogItem{}, fmt.Errorf("error decoding catalog response: %s", err)
				}

				req := c.c.NewRequest(map[string]string{}, "GET", *u, nil)

				resp, err := checkResp(c.c.Http.Do(req))
				if err != nil {
					return CatalogItem{}, fmt.Errorf("error retreiving catalog: %s", err)
				}

				cat := NewCatalogItem(c.c)

				if err = decodeBody(resp, cat.CatalogItem); err != nil {
					return CatalogItem{}, fmt.Errorf("error decoding catalog response: %s", err)
				}

				// The request was successful
				return *cat, nil
			}
		}
	}

	return CatalogItem{}, fmt.Errorf("can't find catalog item: %s", catalogitem)
}

//add callback or channel
func (c *Catalog) UploadOvf(ovaFileName, itemName, description string, chunkSize int) error {

	catalogItemUploadURL, err := findCatalogItemUploadLink(c)
	if err != nil {
		return err
	}

	vappTemplateUrl, err := createItemForUpload(c.c, catalogItemUploadURL, itemName, description)
	if err != nil {
		return err
	}

	vAppTemplate, err := queryVappTemplate(c.c, vappTemplateUrl)
	if err != nil {
		return err
	}

	ovfUploadHref, err := getOvfUploadlink(vAppTemplate)
	if err != nil {
		return err
	}

	fmt.Println("AAA:", ovfUploadHref)

	filesAbsPaths, err := Unpack(ovaFileName)
	if err != nil {
		return err
	}

	for _, path := range filesAbsPaths {
		if filepath.Ext(path) == ".ovf" {
			err := uploadOvfDescription(c.c, path, ovfUploadHref)
			if err != nil {
				return err
			}
		}
	}

	for {
		log.Printf("[TRACE] Sleep... for 5 seconds.\n")
		time.Sleep(time.Second * 5)
		vAppTemplate, err = queryVappTemplate(c.c, vappTemplateUrl)
		if err != nil {
			return err
		}
		if len(vAppTemplate.Files.File) > 1 {
			log.Printf("[TRACE] upload link prepared.\n")
			break
		}
	}

	err = uploadFiles(c.c, filesAbsPaths, *vAppTemplate)
	if err != nil {
		return err
	}

	log.Printf("[TRACE] End ^^..^^ \n")
	return nil
}

func getOvfUploadlink(vappTemplate *types.VAppTemplate) (*url.URL, error) {
	log.Printf("[TRACE] Parsing ofv upload link: %#v\n", vappTemplate)

	ovfUploadHref, err := url.ParseRequestURI(vappTemplate.Files.File[0].Link[0].HREF)
	if err != nil {
		return nil, err
	}

	return ovfUploadHref, nil
}

func queryVappTemplate(client *Client, vappTemplateUrl *url.URL) (*types.VAppTemplate, error) {
	log.Printf("[TRACE] Qeurying vapp template: %s\n", vappTemplateUrl)
	request := client.NewRequest(map[string]string{}, "GET", *vappTemplateUrl, nil)
	response, err := client.Http.Do(request)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	vAppTemplateParsed := &types.VAppTemplate{}

	err = xml.Unmarshal(body, &vAppTemplateParsed)
	if err != nil {
		return nil, err
	}

	log.Printf("[TRACE] Response: %#v\n", response)
	log.Printf("[TRACE] Response body: %s\n", string(body[:]))
	log.Printf("[TRACE] Response body: %#v\n", vAppTemplateParsed)
	return vAppTemplateParsed, nil
}

func uploadOvfDescription(client *Client, ovfFile string, ovfUploadUrl *url.URL) error {
	log.Printf("[TRACE] Uploding ovf description with file: %s and url: %s\n", ovfFile, ovfUploadUrl)
	ovfReader, err := os.Open(ovfFile)
	if err != nil {
		return err
	}

	request := client.NewRequest(map[string]string{}, "PUT", *ovfUploadUrl, ovfReader)
	request.Header.Add("Content-Type", "text/xml")

	response, err := client.Http.Do(request)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	log.Printf("[TRACE] Response: %#v\n", response)
	log.Printf("[TRACE] Response body: %s\n", string(body[:]))
	return nil
}

func findCatalogItemUploadLink(catalog *Catalog) (*url.URL, error) {
	for _, item := range catalog.Catalog.Link {
		if item.Type == "application/vnd.vmware.vcloud.uploadVAppTemplateParams+xml" && item.Rel == "add" {
			log.Printf("[TRACE] Found Catalong link for uplaod: %s\n", item.HREF)

			uploadURL, err := url.ParseRequestURI(item.HREF)
			if err != nil {
				return nil, err
			}

			return uploadURL, nil
		}
	}
	return nil, errors.New("catalog upload url isn't found")
}

func uploadFiles(client *Client, filesAbsPaths []string, parsedVAppTemplate types.VAppTemplate) error {
	for _, item := range parsedVAppTemplate.Files.File {
		if item.BytesTransferred == 0 {
			log.Printf("[TRACE] Found file. Starting uploading: %s\n", item.Name)

			file, err := os.Open(findFilePath(filesAbsPaths, item.Name))
			if err != nil {
				return err
			}
			fileInfo, err := file.Stat()
			if err != nil {
				return err
			}
			fileSize := fileInfo.Size()
			defer file.Close()

			request, err := newFileUploadRequest(item.Link[0].HREF, item.Name, file, fileSize)
			if err != nil {
				return err
			}

			response, err := client.Http.Do(request)
			if err != nil {
				return err
			}

			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return err
			}
			log.Printf("[TRACE] Response: %#v\n", response)
			log.Printf("[TRACE] Response body: %s\n", string(body[:]))

			if response.StatusCode != http.StatusOK {
				return errors.New("File " + item.Name + " upload failed. Err: " + fmt.Sprintf("%#v", response) + " :: " + string(body[:]) + "\n")
			}
		}
	}
	return nil
}

func findFilePath(filesAbsPaths []string, fileName string) string {
	for _, item := range filesAbsPaths {
		_, file := filepath.Split(item)
		if file == fileName {
			return item
		}
	}
	return ""
}

func createItemForUpload(client *Client, createHREF *url.URL, catalogItemName string, itemDescription string) (*url.URL, error) {

	reqBody := bytes.NewBufferString(
		"<UploadVAppTemplateParams xmlns=\"http://www.vmware.com/vcloud/v1.5\" name=\"" + catalogItemName + "\" >" +
			"<Description>" + itemDescription + "</Description>" +
			"</UploadVAppTemplateParams>")

	request := client.NewRequest(map[string]string{}, "POST", *createHREF, reqBody)
	request.Header.Add("Content-Type", "application/vnd.vmware.vcloud.uploadVAppTemplateParams+xml")

	response, err := client.Http.Do(request)
	resBody, err := ioutil.ReadAll(response.Body)

	// Unmarshal the XML.
	catalogItemParsed := &types.CatalogItem{}
	err = xml.Unmarshal(resBody, &catalogItemParsed)
	if err != nil {
		return nil, err
	}

	log.Printf("[TRACE] Response: %#v \n", response)
	log.Printf("[TRACE] Response body: %s\n", string(resBody[:]))
	log.Printf("[TRACE] Catalog item href to query vaapTemplate: %s\n", catalogItemParsed.Entity.HREF)

	ovfUploadUrl, err := url.ParseRequestURI(catalogItemParsed.Entity.HREF)
	if err != nil {
		return nil, err
	}

	return ovfUploadUrl, nil
}

// Creates a new file upload http request with optional extra params
func newFileUploadRequest(requestUrl, paramName string, file io.Reader, fileSize int64) (*http.Request, error) {
	log.Printf("[TRACE] Creating file part upload: %s, %s, %s \n", requestUrl, paramName, file)

	uploadReq, err := http.NewRequest("PUT", requestUrl, file)
	if err != nil {
		return nil, err
	}

	uploadReq.ContentLength = int64(fileSize)
	uploadReq.Header.Set("Content-Length", strconv.FormatInt(uploadReq.ContentLength, 10))

	rangeExpression := "bytes 0-" + strconv.FormatInt(int64(fileSize-1), 10) + "/" + strconv.FormatInt(int64(fileSize), 10)
	uploadReq.Header.Set("Content-Range", rangeExpression)

	for key, value := range uploadReq.Header {
		log.Printf("[TRACE] Header: %s :%s \n", key, value)
	}

	return uploadReq, nil
}
