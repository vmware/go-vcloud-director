// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
)

const (
	defaultPieceSize int64 = 1024 * 1024
)

type Catalog struct {
	Catalog *types.Catalog
	client  *Client
	parent  organization
}

func NewCatalog(client *Client) *Catalog {
	return &Catalog{
		Catalog: new(types.Catalog),
		client:  client,
	}
}

// Delete deletes the Catalog, returning an error if the vCD call fails.
// Link to API call: https://code.vmware.com/apis/1046/vmware-cloud-director/doc/doc/operations/DELETE-Catalog.html
func (catalog *Catalog) Delete(force, recursive bool) error {

	adminCatalogHREF := catalog.client.VCDHREF
	catalogID, err := getBareEntityUuid(catalog.Catalog.ID)
	if err != nil {
		return err
	}
	if catalogID == "" {
		return fmt.Errorf("empty ID returned for catalog %s", catalog.Catalog.Name)
	}
	adminCatalogHREF.Path += "/admin/catalog/" + catalogID

	if force && recursive {
		// A catalog cannot be removed if it has active tasks, or if any of its items have active tasks
		err = catalog.consumeTasks()
		if err != nil {
			return fmt.Errorf("error while consuming tasks from catalog %s: %s", catalog.Catalog.Name, err)
		}
	}

	req := catalog.client.NewRequest(map[string]string{
		"force":     strconv.FormatBool(force),
		"recursive": strconv.FormatBool(recursive),
	}, http.MethodDelete, adminCatalogHREF, nil)

	resp, err := checkResp(catalog.client.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error deleting Catalog %s: %s", catalog.Catalog.Name, err)
	}
	task := NewTask(catalog.client)
	if err = decodeBody(types.BodyTypeXML, resp, task.Task); err != nil {
		return fmt.Errorf("error decoding task response: %s", err)
	}
	if task.Task.Status == "error" {
		return errors.New(combinedTaskErrorMessage(task.Task, fmt.Errorf("catalog %s not properly destroyed", catalog.Catalog.Name)))
	}
	return task.WaitTaskCompletion()
}

// consumeTasks will cancel all catalog tasks and the ones related to its items
// 1. cancel all tasks associated with the catalog and keep them in a list
// 2. find a list of all catalog items
// 3. find a list of all tasks associated with the organization, with name = "syncCatalogItem" or "createCatalogItem"
// 4. loop through the tasks until we find the ones that belong to one of the items - add them to list in 1.
// 5. cancel all the filtered tasks
// 6. wait for the task list until all are finished
func (catalog *Catalog) consumeTasks() error {
	allTasks, err := catalog.client.QueryTaskList(map[string]string{
		"status": "running,preRunning,queued",
	})
	if err != nil {
		return fmt.Errorf("error getting task list from catalog %s: %s", catalog.Catalog.Name, err)
	}
	var taskList []string
	addTask := func(status, href string) {
		if status != "success" && status != "error" && status != "aborted" {
			quickTask := Task{
				client: catalog.client,
				Task: &types.Task{
					HREF: href,
				},
			}
			err = quickTask.CancelTask()
			if err != nil {
				util.Logger.Printf("[consumeTasks] error canceling task: %s\n", err)
			}
			taskList = append(taskList, extractUuid(href))
		}
	}
	if catalog.Catalog.Tasks != nil && len(catalog.Catalog.Tasks.Task) > 0 {
		for _, task := range catalog.Catalog.Tasks.Task {
			addTask(task.Status, task.HREF)
		}
	}
	catalogItemRefs, err := catalog.QueryCatalogItemList()
	if err != nil {
		return fmt.Errorf("error getting catalog %s items list: %s", catalog.Catalog.Name, err)
	}
	for _, task := range allTasks {
		for _, ref := range catalogItemRefs {
			catalogItemId := extractUuid(ref.HREF)
			if extractUuid(task.Object) == catalogItemId {
				addTask(task.Status, task.HREF)
				// No break here: the same object can have more than one task
			}
		}
	}
	_, err = catalog.client.WaitTaskListCompletion(taskList, true)
	if err != nil {
		return fmt.Errorf("error while waiting for task list completion for catalog %s: %s", catalog.Catalog.Name, err)
	}
	return nil
}

// Envelope is a ovf description root element. File contains information for vmdk files.
// Namespace: http://schemas.dmtf.org/ovf/envelope/1
// Description: Envelope is a ovf description root element. File contains information for vmdk files..
type Envelope struct {
	File []struct {
		HREF      string `xml:"href,attr"`
		ID        string `xml:"id,attr"`
		Size      int    `xml:"size,attr"`
		ChunkSize int    `xml:"chunkSize,attr"`
	} `xml:"References>File"`
}

// If catalog item is a valid CatalogItem and the call succeeds,
// then the function returns a CatalogItem. If the item does not
// exist, then it returns an empty CatalogItem. If the call fails
// at any point, it returns an error.
// Deprecated: use GetCatalogItemByName instead
func (cat *Catalog) FindCatalogItem(catalogItemName string) (CatalogItem, error) {
	for _, catalogItems := range cat.Catalog.CatalogItems {
		for _, catalogItem := range catalogItems.CatalogItem {
			if catalogItem.Name == catalogItemName && catalogItem.Type == "application/vnd.vmware.vcloud.catalogItem+xml" {

				cat := NewCatalogItem(cat.client)

				_, err := cat.client.ExecuteRequest(catalogItem.HREF, http.MethodGet,
					"", "error retrieving catalog: %s", nil, cat.CatalogItem)
				return *cat, err
			}
		}
	}

	return CatalogItem{}, nil
}

// UploadOvf uploads an ova/ovf file to a catalog. This method only uploads bits to vCD spool area.
// ovaFileName should be the path of OVA or OVF file(not ovf folder) itself. For OVF,
// user need to make sure all the files that OVF depends on exist and locate under the same folder.
// Returns errors if any occur during upload from vCD or upload process. On upload fail client may need to
// remove vCD catalog item which waits for files to be uploaded. Files from ova are extracted to system
// temp folder "govcd+random number" and left for inspection on error.
func (cat *Catalog) UploadOvf(ovaFileName, itemName, description string, uploadPieceSize int64) (UploadTask, error) {

	//	On a very high level the flow is as follows
	//	1. Makes a POST call to vCD to create the catalog item (also creates a transfer folder in the spool area and as result will give a sparse catalog item resource XML).
	//	2. Wait for the links to the transfer folder to appear in the resource representation of the catalog item.
	//	3. Start uploading bits to the transfer folder
	//	4. Wait on the import task to finish on vCD side -> task success = upload complete

	if *cat == (Catalog{}) {
		return UploadTask{}, errors.New("catalog can not be empty or nil")
	}

	ovaFileName, err := validateAndFixFilePath(ovaFileName)
	if err != nil {
		return UploadTask{}, err
	}

	for _, catalogItemName := range getExistingCatalogItems(cat) {
		if catalogItemName == itemName {
			return UploadTask{}, fmt.Errorf("catalog item '%s' already exists. Upload with different name", itemName)
		}
	}

	isOvf := false
	fileContentType, err := util.GetFileContentType(ovaFileName)
	if err != nil {
		return UploadTask{}, err
	}
	if strings.Contains(fileContentType, "text/xml") {
		isOvf = true
	}
	ovfFilePath := ovaFileName
	tmpDir := path.Dir(ovaFileName)
	filesAbsPaths := []string{ovfFilePath}
	if !isOvf {
		filesAbsPaths, tmpDir, err = util.Unpack(ovaFileName)
		if err != nil {
			return UploadTask{}, fmt.Errorf("%s. Unpacked files for checking are accessible in: %s", err, tmpDir)
		}
		ovfFilePath, err = getOvfPath(filesAbsPaths)
		if err != nil {
			return UploadTask{}, fmt.Errorf("%s. Unpacked files for checking are accessible in: %s", err, tmpDir)
		}
	}

	ovfFileDesc, err := getOvf(ovfFilePath)
	if err != nil {
		return UploadTask{}, fmt.Errorf("%s. OVF/Unpacked files for checking are accessible in: %s", err, tmpDir)
	}

	if !isOvf {
		err = validateOvaContent(filesAbsPaths, &ovfFileDesc, tmpDir)
		if err != nil {
			return UploadTask{}, fmt.Errorf("%s. Unpacked files for checking are accessible in: %s", err, tmpDir)
		}
	} else {
		dir := path.Dir(ovfFilePath)
		for _, fileItem := range ovfFileDesc.File {
			dependFile := path.Join(dir, fileItem.HREF)
			dependFile, err := validateAndFixFilePath(dependFile)
			if err != nil {
				return UploadTask{}, err
			}
			filesAbsPaths = append(filesAbsPaths, dependFile)
		}
	}

	catalogItemUploadURL, err := findCatalogItemUploadLink(cat, "application/vnd.vmware.vcloud.uploadVAppTemplateParams+xml")
	if err != nil {
		return UploadTask{}, err
	}

	vappTemplateUrl, err := createItemForUpload(cat.client, catalogItemUploadURL, itemName, description)
	if err != nil {
		return UploadTask{}, err
	}

	vappTemplate, err := queryVappTemplateAndVerifyTask(cat.client, vappTemplateUrl, itemName)
	if err != nil {
		return UploadTask{}, err
	}

	ovfUploadHref, err := getUploadLink(vappTemplate.Files)
	if err != nil {
		return UploadTask{}, err
	}

	err = uploadOvfDescription(cat.client, ovfFilePath, ovfUploadHref)
	if err != nil {
		removeCatalogItemOnError(cat.client, vappTemplateUrl, itemName)
		return UploadTask{}, err
	}

	vappTemplate, err = waitForTempUploadLinks(cat.client, vappTemplateUrl, itemName)
	if err != nil {
		removeCatalogItemOnError(cat.client, vappTemplateUrl, itemName)
		return UploadTask{}, err
	}

	progressCallBack, uploadProgress := getProgressCallBackFunction()

	uploadError := *new(error)

	// sending upload process to background, this allows not to lock and return task to client
	// The error should be captured in uploadError, but just in case, we add a logging for the
	// main error
	go func() {
		err = uploadFiles(cat.client, vappTemplate, &ovfFileDesc, tmpDir, filesAbsPaths, uploadPieceSize, progressCallBack, &uploadError, isOvf)
		if err != nil {
			util.Logger.Println(strings.Repeat("*", 80))
			util.Logger.Printf("*** [DEBUG - UploadOvf] error calling uploadFiles: %s\n", err)
			util.Logger.Println(strings.Repeat("*", 80))
		}
	}()

	var task Task
	for _, item := range vappTemplate.Tasks.Task {
		task, err = createTaskForVcdImport(cat.client, item.HREF)
		if err != nil {
			removeCatalogItemOnError(cat.client, vappTemplateUrl, itemName)
			return UploadTask{}, err
		}
		if task.Task.Status == "error" {
			removeCatalogItemOnError(cat.client, vappTemplateUrl, itemName)
			return UploadTask{}, fmt.Errorf("task did not complete succesfully: %s", task.Task.Description)
		}
	}

	uploadTask := NewUploadTask(&task, uploadProgress, &uploadError)

	util.Logger.Printf("[TRACE] Upload finished and task for vcd import created. \n")

	return *uploadTask, nil
}

// UploadOvfByLink uploads an OVF file to a catalog from remote URL.
// Returns errors if any occur during upload from VCD or upload process. On upload fail client may need to
// remove VCD catalog item which is in failed state.
func (cat *Catalog) UploadOvfByLink(ovfUrl, itemName, description string) (Task, error) {

	if *cat == (Catalog{}) {
		return Task{}, errors.New("catalog can not be empty or nil")
	}

	for _, catalogItemName := range getExistingCatalogItems(cat) {
		if catalogItemName == itemName {
			return Task{}, fmt.Errorf("catalog item '%s' already exists. Upload with different name", itemName)
		}
	}

	catalogItemUploadURL, err := findCatalogItemUploadLink(cat, "application/vnd.vmware.vcloud.uploadVAppTemplateParams+xml")
	if err != nil {
		return Task{}, err
	}

	vappTemplateUrl, err := createItemWithLink(cat.client, catalogItemUploadURL, itemName, description, ovfUrl)
	if err != nil {
		return Task{}, err
	}

	vappTemplate, err := fetchVappTemplate(cat.client, vappTemplateUrl)
	if err != nil {
		return Task{}, err
	}

	var task Task
	for _, item := range vappTemplate.Tasks.Task {
		task, err = createTaskForVcdImport(cat.client, item.HREF)
		if err != nil {
			removeCatalogItemOnError(cat.client, vappTemplateUrl, itemName)
			return Task{}, err
		}
		if task.Task.Status == "error" {
			removeCatalogItemOnError(cat.client, vappTemplateUrl, itemName)
			return Task{}, fmt.Errorf("task did not complete succesfully: %s", task.Task.Description)
		}
	}

	util.Logger.Printf("[TRACE] task for vcd import created. \n")

	return task, nil
}

// CaptureVappTemplate captures a vApp template from an existing vApp
func (cat *Catalog) CaptureVappTemplate(captureParams *types.CaptureVAppParams) (*VAppTemplate, error) {
	task, err := cat.CaptureVappTemplateAsync(captureParams)
	if err != nil {
		return nil, err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, err
	}

	if task.Task == nil || task.Task.Owner == nil || task.Task.Owner.HREF == "" {
		return nil, fmt.Errorf("task does not have Owner HREF populated: %#v", task)
	}

	// After the task is finished, owner field contains the resulting vApp template
	return cat.GetVappTemplateByHref(task.Task.Owner.HREF)
}

// CaptureVappTemplateAsync triggers vApp template capturing task and returns it
//
// Note. If 'CaptureVAppParams.CopyTpmOnInstantiate' is set, it will be unset for VCD versions
// before 10.4.2 as it would break API call
func (cat *Catalog) CaptureVappTemplateAsync(captureParams *types.CaptureVAppParams) (Task, error) {
	util.Logger.Printf("[TRACE] Capturing vApp template to catalog %s", cat.Catalog.Name)
	if captureParams == nil {
		return Task{}, fmt.Errorf("input CaptureVAppParams cannot be nil")
	}

	captureTemplateHref := cat.client.VCDHREF
	captureTemplateHref.Path += fmt.Sprintf("/catalog/%s/action/captureVApp", extractUuid(cat.Catalog.ID))

	captureParams.Xmlns = types.XMLNamespaceVCloud
	captureParams.XmlnsNs0 = types.XMLNamespaceOVF

	util.Logger.Printf("[TRACE] Url for capturing vApp template: %s", captureTemplateHref.String())

	if cat.client.APIVCDMaxVersionIs("< 37.2") {
		captureParams.CopyTpmOnInstantiate = nil
		util.Logger.Println("[TRACE] Explicitly unsetting 'CopyTpmOnInstantiate' because it was not supported before VCD 10.4.2")
	}

	return cat.client.ExecuteTaskRequest(captureTemplateHref.String(), http.MethodPost,
		types.MimeCaptureVappTemplateParams, "error capturing vApp Template: %s", captureParams)
}

// Upload files for vCD created upload links. Different approach then vmdk file are
// chunked (e.g. test.vmdk.000000000, test.vmdk.000000001 or test.vmdk). vmdk files are chunked if
// in description file attribute ChunkSize is not zero.
// params:
// client - client for requests
// vappTemplate - parsed from response vApp template
// ovfFileDesc - parsed from xml part containing ova files definition
// tempPath - path where extracted files are
// filesAbsPaths - array of extracted files
// uploadPieceSize - size of chunks in which the file will be uploaded to the catalog.
// callBack a function with signature //function(bytesUpload, totalSize) to let the caller monitor progress of the upload operation.
// uploadError - error to be ready be task
func uploadFiles(client *Client, vappTemplate *types.VAppTemplate, ovfFileDesc *Envelope, tempPath string, filesAbsPaths []string, uploadPieceSize int64, progressCallBack func(bytesUpload, totalSize int64), uploadError *error, isOvf bool) error {
	var uploadedBytes int64
	for _, item := range vappTemplate.Files.File {
		if item.BytesTransferred == 0 {
			number, err := getFileFromDescription(item.Name, ovfFileDesc)
			if err != nil {
				util.Logger.Printf("[Error] Error uploading files: %#v", err)
				*uploadError = err
				return err
			}
			if ovfFileDesc.File[number].ChunkSize != 0 {
				chunkFilePaths := getChunkedFilePaths(tempPath, ovfFileDesc.File[number].HREF, ovfFileDesc.File[number].Size, ovfFileDesc.File[number].ChunkSize)
				details := uploadDetails{
					uploadLink:               item.Link[0].HREF,
					uploadedBytes:            uploadedBytes,
					fileSizeToUpload:         int64(ovfFileDesc.File[number].Size),
					uploadPieceSize:          uploadPieceSize,
					uploadedBytesForCallback: uploadedBytes,
					allFilesSize:             getAllFileSizeSum(ovfFileDesc),
					callBack:                 progressCallBack,
					uploadError:              uploadError,
				}
				tempVar, err := uploadMultiPartFile(client, chunkFilePaths, details)
				if err != nil {
					util.Logger.Printf("[Error] Error uploading files: %#v", err)
					*uploadError = err
					return err
				}
				uploadedBytes += tempVar
			} else {
				details := uploadDetails{
					uploadLink:               item.Link[0].HREF,
					uploadedBytes:            0,
					fileSizeToUpload:         item.Size,
					uploadPieceSize:          uploadPieceSize,
					uploadedBytesForCallback: uploadedBytes,
					allFilesSize:             getAllFileSizeSum(ovfFileDesc),
					callBack:                 progressCallBack,
					uploadError:              uploadError,
				}
				tempVar, err := uploadFile(client, findFilePath(filesAbsPaths, item.Name), details)
				if err != nil {
					util.Logger.Printf("[Error] Error uploading files: %#v", err)
					*uploadError = err
					return err
				}
				uploadedBytes += tempVar
			}
		}
	}

	//remove extracted files with temp dir
	//If isOvf flag is true, means tempPath is origin OVF folder, not extracted, won't delete
	if !isOvf {
		err := os.RemoveAll(tempPath)
		if err != nil {
			util.Logger.Printf("[Error] Error removing temporary files: %#v", err)
			*uploadError = err
			return err
		}
	}
	uploadError = nil
	return nil
}

func getFileFromDescription(fileToFind string, ovfFileDesc *Envelope) (int, error) {
	for fileInArray, item := range ovfFileDesc.File {
		if item.HREF == fileToFind {
			util.Logger.Printf("[TRACE] getFileFromDescription - found matching file: %s in array: %d\n", fileToFind, fileInArray)
			return fileInArray, nil
		}
	}
	return -1, errors.New("file expected from vcd didn't match any description file")
}

func getAllFileSizeSum(ovfFileDesc *Envelope) (sizeSum int64) {
	sizeSum = 0
	for _, item := range ovfFileDesc.File {
		sizeSum += int64(item.Size)
	}
	return
}

// Uploads chunked ova file for vCD created upload link.
// params:
// client - client for requests
// vappTemplate - parsed from response vApp template
// filePaths - all chunked vmdk file paths
// uploadDetails - file upload settings and data
func uploadMultiPartFile(client *Client, filePaths []string, uDetails uploadDetails) (int64, error) {
	util.Logger.Printf("[TRACE] Upload multi part file: %v\n, href: %s, size: %v", filePaths, uDetails.uploadLink, uDetails.fileSizeToUpload)

	var uploadedBytes int64

	for i, filePath := range filePaths {
		util.Logger.Printf("[TRACE] Uploading file: %v\n", i+1)
		uDetails.uploadedBytesForCallback += uploadedBytes // previous files uploaded size plus current upload size
		uDetails.uploadedBytes = uploadedBytes
		tempVar, err := uploadFile(client, filePath, uDetails)
		if err != nil {
			return uploadedBytes, err
		}
		uploadedBytes += tempVar
	}
	return uploadedBytes, nil
}

// Function waits until vCD provides temporary file upload links.
func waitForTempUploadLinks(client *Client, vappTemplateUrl *url.URL, newItemName string) (*types.VAppTemplate, error) {
	var vAppTemplate *types.VAppTemplate
	var err error
	for {
		util.Logger.Printf("[TRACE] Sleep... for 5 seconds.\n")
		time.Sleep(time.Second * 5)
		vAppTemplate, err = queryVappTemplateAndVerifyTask(client, vappTemplateUrl, newItemName)
		if err != nil {
			return nil, err
		}
		if vAppTemplate.Files != nil && len(vAppTemplate.Files.File) > 1 {
			util.Logger.Printf("[TRACE] upload link prepared.\n")
			break
		}
	}
	return vAppTemplate, nil
}

func queryVappTemplateAndVerifyTask(client *Client, vappTemplateUrl *url.URL, newItemName string) (*types.VAppTemplate, error) {
	util.Logger.Printf("[TRACE] Querying vApp template: %s\n", vappTemplateUrl)

	vappTemplateParsed, err := fetchVappTemplate(client, vappTemplateUrl)
	if err != nil {
		return nil, err
	}

	if vappTemplateParsed.Tasks == nil {
		util.Logger.Printf("[ERROR] the vApp Template %s does not contain tasks, an error happened during upload: %v", vappTemplateUrl, vappTemplateParsed)
		return vappTemplateParsed, fmt.Errorf("the vApp Template %s does not contain tasks, an error happened during upload", vappTemplateUrl)
	}

	for _, task := range vappTemplateParsed.Tasks.Task {
		if task.Status == "error" && newItemName == task.Owner.Name {
			util.Logger.Printf("[Error] %#v", task.Error)
			return vappTemplateParsed, fmt.Errorf("error in vcd returned error code: %d, error: %s and message: %s ", task.Error.MajorErrorCode, task.Error.MinorErrorCode, task.Error.Message)
		}
	}

	return vappTemplateParsed, nil
}

func fetchVappTemplate(client *Client, vappTemplateUrl *url.URL) (*types.VAppTemplate, error) {
	util.Logger.Printf("[TRACE] Querying vApp template: %s\n", vappTemplateUrl)

	vappTemplateParsed := &types.VAppTemplate{}

	_, err := client.ExecuteRequest(vappTemplateUrl.String(), http.MethodGet,
		"", "error fetching vApp template: %s", nil, vappTemplateParsed)
	if err != nil {
		return nil, err
	}

	return vappTemplateParsed, nil
}

// Uploads ovf description file from unarchived provided ova file. As a result vCD will generate temporary upload links which has to be queried later.
// Function will return parsed part for upload files from description xml.
func uploadOvfDescription(client *Client, ovfFile string, ovfUploadUrl *url.URL) error {
	util.Logger.Printf("[TRACE] Uploding ovf description with file: %s and url: %s\n", ovfFile, ovfUploadUrl)
	openedFile, err := os.Open(filepath.Clean(ovfFile))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	ovfReader := io.TeeReader(openedFile, &buf)

	request := client.NewRequest(map[string]string{}, http.MethodPut, *ovfUploadUrl, ovfReader)
	request.Header.Add("Content-Type", "text/xml")

	_, err = checkResp(client.Http.Do(request))
	if err != nil {
		return err
	}

	err = openedFile.Close()
	if err != nil {
		util.Logger.Printf("[Error] Error closing file: %#v", err)
		return err
	}

	return nil
}

func parseOvfFileDesc(file *os.File, ovfFileDesc *Envelope) error {
	ovfXml, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(ovfXml, &ovfFileDesc)
	if err != nil {
		return err
	}
	return nil
}

func findCatalogItemUploadLink(catalog *Catalog, applicationType string) (*url.URL, error) {
	for _, item := range catalog.Catalog.Link {
		if item.Type == applicationType && item.Rel == "add" {
			util.Logger.Printf("[TRACE] Found Catalong link for upload: %s\n", item.HREF)

			uploadURL, err := url.ParseRequestURI(item.HREF)
			if err != nil {
				return nil, err
			}

			util.Logger.Printf("[TRACE] findCatalogItemUploadLink - catalog item upload url found: %s \n", uploadURL)
			return uploadURL, nil
		}
	}
	return nil, errors.New("catalog upload URL not found")
}

func getExistingCatalogItems(catalog *Catalog) (catalogItemNames []string) {
	for _, catalogItems := range catalog.Catalog.CatalogItems {
		for _, catalogItem := range catalogItems.CatalogItem {
			catalogItemNames = append(catalogItemNames, catalogItem.Name)
		}
	}
	return
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

// Initiates creation of item and returns ovf upload url for created item.
func createItemForUpload(client *Client, createHREF *url.URL, catalogItemName string, itemDescription string) (*url.URL, error) {
	util.Logger.Printf("[TRACE] createItemForUpload: %s, item name: %s, description: %s \n", createHREF, catalogItemName, itemDescription)
	reqBody := bytes.NewBufferString(
		"<UploadVAppTemplateParams xmlns=\"" + types.XMLNamespaceVCloud + "\" name=\"" + catalogItemName + "\" >" +
			"<Description>" + itemDescription + "</Description>" +
			"</UploadVAppTemplateParams>")

	request := client.NewRequest(map[string]string{}, http.MethodPost, *createHREF, reqBody)
	request.Header.Add("Content-Type", "application/vnd.vmware.vcloud.uploadVAppTemplateParams+xml")

	response, err := checkResp(client.Http.Do(request))
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			util.Logger.Printf("error closing response Body [createItemForUpload]: %s", err)
		}
	}(response.Body)

	catalogItemParsed := &types.CatalogItem{}
	if err = decodeBody(types.BodyTypeXML, response, catalogItemParsed); err != nil {
		return nil, err
	}

	util.Logger.Printf("[TRACE] Catalog item parsed: %#v\n", catalogItemParsed)

	ovfUploadUrl, err := url.ParseRequestURI(catalogItemParsed.Entity.HREF)
	if err != nil {
		return nil, err
	}

	return ovfUploadUrl, nil
}

// Initiates creation of item in catalog and returns vappTeamplate Url for created item.
func createItemWithLink(client *Client, createHREF *url.URL, catalogItemName, itemDescription, vappTemplateRemoteUrl string) (*url.URL, error) {
	util.Logger.Printf("[TRACE] createItemWithLink: %s, item name: %s, description: %s, vappTemplateRemoteUrl: %s \n",
		createHREF, catalogItemName, itemDescription, vappTemplateRemoteUrl)

	reqTemplate := `<UploadVAppTemplateParams xmlns="%s" name="%s" sourceHref="%s"><Description>%s</Description></UploadVAppTemplateParams>`
	reqBody := bytes.NewBufferString(fmt.Sprintf(reqTemplate, types.XMLNamespaceVCloud, catalogItemName, vappTemplateRemoteUrl, itemDescription))
	request := client.NewRequest(map[string]string{}, http.MethodPost, *createHREF, reqBody)
	request.Header.Add("Content-Type", "application/vnd.vmware.vcloud.uploadVAppTemplateParams+xml")

	response, err := checkResp(client.Http.Do(request))
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			util.Logger.Printf("error closing response Body [createItemWithLink]: %s", err)
		}
	}(response.Body)

	catalogItemParsed := &types.CatalogItem{}
	if err = decodeBody(types.BodyTypeXML, response, catalogItemParsed); err != nil {
		return nil, err
	}

	util.Logger.Printf("[TRACE] Catalog item parsed: %#v\n", catalogItemParsed)

	vappTemplateUrl, err := url.ParseRequestURI(catalogItemParsed.Entity.HREF)
	if err != nil {
		return nil, err
	}

	return vappTemplateUrl, nil
}

// Helper method to get path to multi-part files.
// For example a file called test.vmdk with total_file_size = 100 bytes and part_size = 40 bytes, implies the file is made of *3* part files.
//   - test.vmdk.000000000 = 40 bytes
//   - test.vmdk.000000001 = 40 bytes
//   - test.vmdk.000000002 = 20 bytes
//
// Say base_dir = /dummy_path/, and base_file_name = test.vmdk then
// the output of this function will be [/dummy_path/test.vmdk.000000000,
// /dummy_path/test.vmdk.000000001, /dummy_path/test.vmdk.000000002]
func getChunkedFilePaths(baseDir, baseFileName string, totalFileSize, partSize int) []string {
	var filePaths []string
	numbParts := math.Ceil(float64(totalFileSize) / float64(partSize))
	for i := 0; i < int(numbParts); i++ {
		temp := "000000000" + strconv.Itoa(i)
		postfix := temp[len(temp)-9:]
		filePath := path.Join(baseDir, baseFileName+"."+postfix)
		filePaths = append(filePaths, filePath)
	}

	util.Logger.Printf("[TRACE] Chunked files file paths: %s \n", filePaths)
	return filePaths
}

func getOvfPath(filesAbsPaths []string) (string, error) {
	for _, filePath := range filesAbsPaths {
		if filepath.Ext(filePath) == ".ovf" {
			return filePath, nil
		}
	}
	return "", errors.New("ova is not correct - missing ovf file")
}

func getOvf(ovfFilePath string) (Envelope, error) {
	openedFile, err := os.Open(filepath.Clean(ovfFilePath))
	if err != nil {
		return Envelope{}, err
	}

	var ovfFileDesc Envelope
	err = parseOvfFileDesc(openedFile, &ovfFileDesc)
	if err != nil {
		return Envelope{}, err
	}

	err = openedFile.Close()
	if err != nil {
		util.Logger.Printf("[Error] Error closing file: %#v", err)
		return Envelope{}, err
	}

	return ovfFileDesc, nil
}

func validateOvaContent(filesAbsPaths []string, ovfFileDesc *Envelope, tempPath string) error {
	for _, fileDescription := range ovfFileDesc.File {
		if fileDescription.ChunkSize == 0 {
			err := checkIfFileMatchesDescription(filesAbsPaths, fileDescription)
			if err != nil {
				return err
			}
			// check chunked ova content
		} else {
			chunkFilePaths := getChunkedFilePaths(tempPath, fileDescription.HREF, fileDescription.Size, fileDescription.ChunkSize)
			for part, chunkedFilePath := range chunkFilePaths {
				_, fileName := filepath.Split(chunkedFilePath)
				chunkedFileSize := fileDescription.Size - part*fileDescription.ChunkSize
				if chunkedFileSize > fileDescription.ChunkSize {
					chunkedFileSize = fileDescription.ChunkSize
				}
				chunkedFileDescription := struct {
					HREF      string `xml:"href,attr"`
					ID        string `xml:"id,attr"`
					Size      int    `xml:"size,attr"`
					ChunkSize int    `xml:"chunkSize,attr"`
				}{fileName, "", chunkedFileSize, fileDescription.ChunkSize}
				err := checkIfFileMatchesDescription(filesAbsPaths, chunkedFileDescription)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func checkIfFileMatchesDescription(filesAbsPaths []string, fileDescription struct {
	HREF      string `xml:"href,attr"`
	ID        string `xml:"id,attr"`
	Size      int    `xml:"size,attr"`
	ChunkSize int    `xml:"chunkSize,attr"`
}) error {
	filePath := findFilePath(filesAbsPaths, fileDescription.HREF)
	if filePath == "" {
		return fmt.Errorf("file '%s' described in ovf was not found in ova", fileDescription.HREF)
	}
	if fileInfo, err := os.Stat(filePath); err == nil {
		if fileDescription.Size > 0 && (fileInfo.Size() != int64(fileDescription.Size)) {
			return fmt.Errorf("file size didn't match described in ovf: %s", filePath)
		}
	} else {
		return err
	}
	return nil
}

func removeCatalogItemOnError(client *Client, vappTemplateLink *url.URL, itemName string) {
	if vappTemplateLink != nil {
		util.Logger.Printf("[TRACE] Deleting Catalog item %v", vappTemplateLink)

		// wait for task, cancel it and catalog item will be removed.
		var vAppTemplate *types.VAppTemplate
		var err error
		for {
			util.Logger.Printf("[TRACE] Sleep... for 5 seconds.\n")
			time.Sleep(time.Second * 5)
			vAppTemplate, err = queryVappTemplateAndVerifyTask(client, vappTemplateLink, itemName)
			if err != nil {
				util.Logger.Printf("[Error] Error deleting Catalog item %s: %s", vappTemplateLink, err)
			}
			if vAppTemplate.Tasks == nil {
				util.Logger.Printf("[Error] Error deleting Catalog item %s: it doesn't contain any task", vappTemplateLink)
				return
			}
			if vAppTemplate.Tasks != nil && len(vAppTemplate.Tasks.Task) > 0 {
				util.Logger.Printf("[TRACE] Task found. Will try to cancel.\n")
				break
			}
		}

		for _, taskItem := range vAppTemplate.Tasks.Task {
			if itemName == taskItem.Owner.Name {
				task := NewTask(client)
				task.Task = taskItem
				err = task.CancelTask()
				if err != nil {
					util.Logger.Printf("[ERROR] Error canceling task for catalog item upload %#v", err)
				}
			}
		}
	} else {
		util.Logger.Printf("[Error] Failed to delete catalog item created with error: %v", vappTemplateLink)
	}
}

// UploadMediaImage uploads a media image to the catalog
func (cat *Catalog) UploadMediaImage(mediaName, mediaDescription, filePath string, uploadPieceSize int64) (UploadTask, error) {
	return cat.UploadMediaFile(mediaName, mediaDescription, filePath, uploadPieceSize, true)
}

// UploadMediaFile uploads any file to the catalog.
// However, if checkFileIsIso is true, only .ISO are allowed.
func (cat *Catalog) UploadMediaFile(fileName, mediaDescription, filePath string, uploadPieceSize int64, checkFileIsIso bool) (UploadTask, error) {

	if *cat == (Catalog{}) {
		return UploadTask{}, errors.New("catalog can not be empty or nil")
	}

	mediaFilePath, err := validateAndFixFilePath(filePath)
	if err != nil {
		return UploadTask{}, err
	}

	if checkFileIsIso {
		isISOGood, err := verifyIso(mediaFilePath)
		if err != nil || !isISOGood {
			return UploadTask{}, fmt.Errorf("[ERROR] File %s isn't correct iso file: %#v", mediaFilePath, err)
		}
	}

	file, e := os.Stat(mediaFilePath)
	if e != nil {
		return UploadTask{}, fmt.Errorf("[ERROR] Issue finding file: %#v", e)
	}
	fileSize := file.Size()

	for _, catalogItemName := range getExistingCatalogItems(cat) {
		if catalogItemName == fileName {
			return UploadTask{}, fmt.Errorf("media item '%s' already exists. Upload with different name", fileName)
		}
	}

	catalogItemUploadURL, err := findCatalogItemUploadLink(cat, "application/vnd.vmware.vcloud.media+xml")
	if err != nil {
		return UploadTask{}, err
	}

	media, err := createMedia(cat.client, catalogItemUploadURL.String(), fileName, mediaDescription, fileSize)
	if err != nil {
		return UploadTask{}, fmt.Errorf("[ERROR] Issue creating media: %#v", err)
	}

	createdMedia, err := queryMedia(cat.client, media.Entity.HREF, fileName)
	if err != nil {
		return UploadTask{}, err
	}

	return executeUpload(cat.client, createdMedia, mediaFilePath, fileName, fileSize, uploadPieceSize)
}

// Refresh gets a fresh copy of the catalog from vCD
func (cat *Catalog) Refresh() error {
	if cat == nil || *cat == (Catalog{}) || cat.Catalog.HREF == "" {
		return fmt.Errorf("cannot refresh, Object is empty or HREF is empty")
	}

	refreshedCatalog := &types.Catalog{}

	_, err := cat.client.ExecuteRequest(cat.Catalog.HREF, http.MethodGet,
		"", "error refreshing VDC: %s", nil, refreshedCatalog)
	if err != nil {
		return err
	}
	cat.Catalog = refreshedCatalog

	return nil
}

// GetCatalogItemByHref finds a CatalogItem by HREF
// On success, returns a pointer to the CatalogItem structure and a nil error
// On failure, returns a nil pointer and an error
func (cat *Catalog) GetCatalogItemByHref(catalogItemHref string) (*CatalogItem, error) {

	catItem := NewCatalogItem(cat.client)

	_, err := cat.client.ExecuteRequest(catalogItemHref, http.MethodGet,
		"", "error retrieving catalog item: %s", nil, catItem.CatalogItem)
	if err != nil {
		return nil, err
	}
	return catItem, nil
}

// GetVappTemplateByHref finds a vApp template by HREF
// On success, returns a pointer to the vApp template structure and a nil error
// On failure, returns a nil pointer and an error
func (cat *Catalog) GetVappTemplateByHref(href string) (*VAppTemplate, error) {
	return getVAppTemplateByHref(cat.client, href)
}

// getVAppTemplateByHref finds a vApp template by HREF
// On success, returns a pointer to the vApp template structure and a nil error
// On failure, returns a nil pointer and an error
func getVAppTemplateByHref(client *Client, href string) (*VAppTemplate, error) {
	vappTemplate := NewVAppTemplate(client)

	_, err := client.ExecuteRequest(href, http.MethodGet, "", "error retrieving vApp Template: %s", nil, vappTemplate.VAppTemplate)
	if err != nil {
		return nil, err
	}
	return vappTemplate, nil
}

// GetCatalogItemByName finds a CatalogItem by Name
// On success, returns a pointer to the CatalogItem structure and a nil error
// On failure, returns a nil pointer and an error
func (cat *Catalog) GetCatalogItemByName(catalogItemName string, refresh bool) (*CatalogItem, error) {
	if refresh {
		err := cat.Refresh()
		if err != nil {
			return nil, err
		}
	}
	for _, catalogItems := range cat.Catalog.CatalogItems {
		for _, catalogItem := range catalogItems.CatalogItem {
			if catalogItem.Name == catalogItemName && catalogItem.Type == "application/vnd.vmware.vcloud.catalogItem+xml" {
				return cat.GetCatalogItemByHref(catalogItem.HREF)
			}
		}
	}
	return nil, ErrorEntityNotFound
}

// GetVAppTemplateByName finds a VAppTemplate by Name
// On success, returns a pointer to the VAppTemplate structure and a nil error
// On failure, returns a nil pointer and an error
func (cat *Catalog) GetVAppTemplateByName(vAppTemplateName string) (*VAppTemplate, error) {
	vAppTemplateQueryResult, err := cat.QueryVappTemplateWithName(vAppTemplateName)
	if err != nil {
		return nil, err
	}
	return cat.GetVappTemplateByHref(vAppTemplateQueryResult.HREF)
}

// GetCatalogItemById finds a Catalog Item by ID
// On success, returns a pointer to the CatalogItem structure and a nil error
// On failure, returns a nil pointer and an error
func (cat *Catalog) GetCatalogItemById(catalogItemId string, refresh bool) (*CatalogItem, error) {
	if refresh {
		err := cat.Refresh()
		if err != nil {
			return nil, err
		}
	}
	for _, catalogItems := range cat.Catalog.CatalogItems {
		for _, catalogItem := range catalogItems.CatalogItem {
			if equalIds(catalogItemId, catalogItem.ID, catalogItem.HREF) && catalogItem.Type == "application/vnd.vmware.vcloud.catalogItem+xml" {
				return cat.GetCatalogItemByHref(catalogItem.HREF)
			}
		}
	}
	return nil, ErrorEntityNotFound
}

// GetVAppTemplateById finds a vApp Template by ID.
// On success, returns a pointer to the VAppTemplate structure and a nil error.
// On failure, returns a nil pointer and an error.
func (cat *Catalog) GetVAppTemplateById(vAppTemplateId string) (*VAppTemplate, error) {
	return getVAppTemplateById(cat.client, vAppTemplateId)
}

// getVAppTemplateById finds a vApp Template by ID.
// On success, returns a pointer to the VAppTemplate structure and a nil error.
// On failure, returns a nil pointer and an error.
func getVAppTemplateById(client *Client, vAppTemplateId string) (*VAppTemplate, error) {
	vappTemplateHref := client.VCDHREF
	vappTemplateHref.Path += "/vAppTemplate/vappTemplate-" + extractUuid(vAppTemplateId)

	vappTemplate, err := getVAppTemplateByHref(client, vappTemplateHref.String())
	if err != nil {
		return nil, fmt.Errorf("could not find vApp Template with ID %s: %s", vAppTemplateId, err)
	}
	return vappTemplate, nil
}

// GetCatalogItemByNameOrId finds a Catalog Item by Name or ID.
// On success, returns a pointer to the CatalogItem structure and a nil error
// On failure, returns a nil pointer and an error
func (cat *Catalog) GetCatalogItemByNameOrId(identifier string, refresh bool) (*CatalogItem, error) {
	getByName := func(name string, refresh bool) (interface{}, error) { return cat.GetCatalogItemByName(name, refresh) }
	getById := func(id string, refresh bool) (interface{}, error) { return cat.GetCatalogItemById(id, refresh) }
	entity, err := getEntityByNameOrId(getByName, getById, identifier, refresh)
	if entity == nil {
		return nil, err
	}
	return entity.(*CatalogItem), err
}

// GetVAppTemplateByNameOrId finds a vApp Template by Name or ID.
// On success, returns a pointer to the VAppTemplate structure and a nil error
// On failure, returns a nil pointer and an error
func (cat *Catalog) GetVAppTemplateByNameOrId(identifier string, refresh bool) (*VAppTemplate, error) {
	getByName := func(name string, refresh bool) (interface{}, error) { return cat.GetVAppTemplateByName(name) }
	getById := func(id string, refresh bool) (interface{}, error) { return cat.GetVAppTemplateById(id) }
	entity, err := getEntityByNameOrIdSkipNonId(getByName, getById, identifier, refresh)
	if entity == nil {
		return nil, err
	}
	return entity.(*VAppTemplate), err
}

// QueryMediaList retrieves a list of media items for the catalog
func (catalog *Catalog) QueryMediaList() ([]*types.MediaRecordType, error) {
	typeMedia := "media"
	if catalog.client.IsSysAdmin {
		typeMedia = "adminMedia"
	}

	filter := fmt.Sprintf("catalog==%s", url.QueryEscape(catalog.Catalog.HREF))
	results, err := catalog.client.QueryWithNotEncodedParams(nil, map[string]string{"type": typeMedia, "filter": filter, "filterEncoded": "true"})
	if err != nil {
		return nil, fmt.Errorf("error querying medias: %s", err)
	}

	mediaResults := results.Results.MediaRecord
	if catalog.client.IsSysAdmin {
		mediaResults = results.Results.AdminMediaRecord
	}
	return mediaResults, nil
}

// getOrgInfo finds the organization to which the catalog belongs, and returns its name and ID
func (catalog *Catalog) getOrgInfo() (*TenantContext, error) {
	org := catalog.parent
	if org == nil {
		return nil, fmt.Errorf("no parent found for catalog %s", catalog.Catalog.Name)
	}

	return org.tenantContext()
}

func publishToExternalOrganizations(client *Client, url string, tenantContext *TenantContext, publishExternalCatalog types.PublishExternalCatalogParams) error {
	url = url + "/action/publishToExternalOrganizations"

	publishExternalCatalog.Xmlns = types.XMLNamespaceVCloud

	if tenantContext != nil {
		client.SetCustomHeader(getTenantContextHeader(tenantContext))
	}

	err := client.ExecuteRequestWithoutResponse(url, http.MethodPost,
		types.PublishExternalCatalog, "error publishing to external organization: %s", publishExternalCatalog)

	if tenantContext != nil {
		client.RemoveProvidedCustomHeaders(getTenantContextHeader(tenantContext))
	}

	return err
}

// PublishToExternalOrganizations publishes a catalog to external organizations.
func (cat *Catalog) PublishToExternalOrganizations(publishExternalCatalog types.PublishExternalCatalogParams) error {
	if cat.Catalog == nil {
		return fmt.Errorf("cannot publish to external organization, Object is empty")
	}

	catalogUrl := cat.Catalog.HREF
	if catalogUrl == "nil" || catalogUrl == "" {
		return fmt.Errorf("cannot publish to external organization, HREF is empty")
	}

	err := publishToExternalOrganizations(cat.client, catalogUrl, nil, publishExternalCatalog)
	if err != nil {
		return err
	}

	err = cat.Refresh()
	if err != nil {
		return err
	}

	return err
}

// elementSync is a low level function that synchronises a Catalog, AdminCatalog, CatalogItem, or Media item
func elementSync(client *Client, elementHref, label string) error {
	task, err := elementLaunchSync(client, elementHref, label)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// queryMediaList retrieves a list of media items for a given catalog or AdminCatalog
func queryMediaList(client *Client, catalogHref string) ([]*types.MediaRecordType, error) {
	typeMedia := "media"
	if client.IsSysAdmin {
		typeMedia = "adminMedia"
	}

	filter := fmt.Sprintf("catalog==%s", url.QueryEscape(catalogHref))
	results, err := client.QueryWithNotEncodedParams(nil, map[string]string{"type": typeMedia, "filter": filter, "filterEncoded": "true"})
	if err != nil {
		return nil, fmt.Errorf("error querying medias: %s", err)
	}

	mediaResults := results.Results.MediaRecord
	if client.IsSysAdmin {
		mediaResults = results.Results.AdminMediaRecord
	}
	return mediaResults, nil
}

// elementLaunchSync is a low level function that starts synchronisation for Catalog, AdminCatalog, CatalogItem, or Media item
func elementLaunchSync(client *Client, elementHref, label string) (*Task, error) {
	util.Logger.Printf("[TRACE] elementLaunchSync '%s' \n", label)
	href := elementHref + "/action/sync"
	syncTask, err := client.ExecuteTaskRequest(href, http.MethodPost,
		"", "error synchronizing "+label+": %s", nil)

	if err != nil {
		// This process may fail due to a possible race condition: a synchronisation process may start in background
		// after we check for existing tasks (in the function that called this one)
		// and before we run the request in this function.
		// In a Terraform vcd_subscribed_catalog operation, the completeness of the synchronisation
		// will be ensured at the next refresh.
		if strings.Contains(err.Error(), "LIBRARY_ITEM_SYNC") {
			util.Logger.Printf("[SYNC FAILURE] error when launching synchronisation: %s\n", err)
			return nil, nil
		}
		return nil, err
	}
	return &syncTask, nil
}

// QueryTaskList retrieves a list of tasks associated to the Catalog
func (catalog *Catalog) QueryTaskList(filter map[string]string) ([]*types.QueryResultTaskRecordType, error) {
	var newFilter = map[string]string{
		"object": catalog.Catalog.HREF,
	}
	for k, v := range filter {
		newFilter[k] = v
	}
	return catalog.client.QueryTaskList(newFilter)
}

// GetCatalogByHref allows retrieving a catalog from HREF, without a fully qualified Org object
func (client *Client) GetCatalogByHref(catalogHref string) (*Catalog, error) {
	catalogHref = strings.Replace(catalogHref, "/api/admin/catalog", "/api/catalog", 1)

	cat := NewCatalog(client)

	_, err := client.ExecuteRequest(catalogHref, http.MethodGet,
		"", "error retrieving catalog: %s", nil, cat.Catalog)

	if err != nil {
		return nil, err
	}
	// Setting the catalog parent, necessary to handle the tenant context
	org := NewOrg(client)
	for _, link := range cat.Catalog.Link {
		if link.Rel == "up" && link.Type == types.MimeOrg {
			_, err = client.ExecuteRequest(link.HREF, http.MethodGet,
				"", "error retrieving parent Org: %s", nil, org.Org)
			if err != nil {
				return nil, fmt.Errorf("error retrieving catalog parent: %s", err)
			}
			break
		}
	}
	cat.parent = org
	return cat, nil
}

// GetCatalogById allows retrieving a catalog from ID, without a fully qualified Org object
func (client *Client) GetCatalogById(catalogId string) (*Catalog, error) {
	href, err := url.JoinPath(client.VCDHREF.String(), "catalog", extractUuid(catalogId))
	if err != nil {
		return nil, err
	}
	return client.GetCatalogByHref(href)
}

// GetCatalogByName allows retrieving a catalog from name, without a fully qualified Org object
func (client *Client) GetCatalogByName(parentOrg, catalogName string) (*Catalog, error) {
	catalogs, err := queryCatalogList(client, nil)
	if err != nil {
		return nil, err
	}
	var parentOrgs []string
	for _, cat := range catalogs {
		if cat.Name == catalogName && cat.OrgName == parentOrg {
			return client.GetCatalogByHref(cat.HREF)
		}
		if cat.Name == catalogName {
			parentOrgs = append(parentOrgs, cat.OrgName)
		}
	}
	parents := ""
	if len(parentOrgs) > 0 {
		parents = fmt.Sprintf(" - Found catalog %s in Orgs %v", catalogName, parentOrgs)
	}
	return nil, fmt.Errorf("no catalog '%s' found in Org %s%s", catalogName, parentOrg, parents)
}

// WaitForTasks waits for the catalog's tasks to complete
func (cat *Catalog) WaitForTasks() error {
	if ResourceInProgress(cat.Catalog.Tasks) {
		err := WaitResource(func() (*types.TasksInProgress, error) {
			err := cat.Refresh()
			if err != nil {
				return nil, err
			}
			return cat.Catalog.Tasks, nil
		})
		return err
	}
	return nil
}
