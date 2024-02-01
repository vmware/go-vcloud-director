/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/govcd/internal/udf"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sigs.k8s.io/yaml"
	"time"
)

const (
	manifestKey     = "manifest"
	iconKey         = "icon"
	eulaKey         = "eula"
	certificateKey  = "certificate"
	manifestVersion = "version"
	manifestName    = "name"
	manifestVendor  = "vendor"
)

// wantedFiles are the files expected for data solution 1.3
var wantedFiles = map[string]isoFileDef{
	eulaKey:        {namePattern: `(?i)^eula\.txt$`},
	manifestKey:    {namePattern: `^manifest\.yaml$`},
	iconKey:        {namePattern: `^icon\.(png|jpg|jpeg|jfif|svg|bmp|ico)$`},
	certificateKey: {namePattern: `^certificate\.pem$`},
}

// isoFileDef describes a file wanted from the .ISO container
type isoFileDef struct {
	namePattern   string         // name expression seeked
	nameRegexp    *regexp.Regexp // regular expression later used to check if the file name matches
	foundFileName string         // Name of the file matched by the regular expression
	contents      []byte         // contents of the found file
}

// iconTypes defines the type of the icon as retrieved from the .ISO file
var iconTypes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".jfif": "image/jpeg",
	".svg":  "image/svg+xml",
	".bmp":  "image/bmp",
	".ico":  "image/x-icon",
}

// TODO: move SolutionEntity and SolutionsOrigin to types

// SolutionEntity is the definition of a solution RDE
type SolutionEntity struct {
	Eula     string          `json:"eula"`
	Icon     string          `json:"icon"`
	Manifest map[string]any  `json:"manifest"`
	Origin   SolutionsOrigin `json:"origin"`
	Status   string          `json:"status"`
}

type SolutionsOrigin struct {
	Type          string `json:"type"`
	AcceptedBy    string `json:"acceptedBy"`
	AcceptedOn    string `json:"acceptedOn"`
	CatalogItemId string `json:"catalogItemId"`
}

func getContentsFromIsoFiles(isoFileName string, wanted map[string]isoFileDef) (map[string]isoFileDef, error) {
	if stat, err := os.Stat(isoFileName); err != nil || stat.IsDir() {
		return nil, fmt.Errorf("file %s does not exist", isoFileName)
	}
	var err error
	var result = make(map[string]isoFileDef)
	for key, elem := range wanted {
		if elem.nameRegexp == nil {
			elem.nameRegexp, err = regexp.Compile(elem.namePattern)
			if err != nil {
				return nil, fmt.Errorf("error compiling regular expression '%s': %s", elem.namePattern, err)
			}
		}
		result[key] = elem
	}

	file, err := os.Open(filepath.Clean(isoFileName))
	if err != nil {
		return nil, err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			panic(err)
		}
	}()

	reader, err := udf.Open(file)
	if err != nil {
		return nil, err
	}

	rootDir, err := reader.RootDir()
	if err != nil {
		return nil, err
	}

	children, err := reader.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	matchItem := func(name string) string {
		for k, elem := range result {
			if elem.nameRegexp.MatchString(name) {
				return k
			}
		}
		return ""
	}
	for idx := range children {
		child := &children[idx]
		if itemId := matchItem(child.Name()); itemId != "" {
			fileReader, err := reader.NewFileReader(child)

			if err != nil {
				return nil, err
			}
			var fileContent []byte = make([]byte, child.Size())
			if _, err = fileReader.Read(fileContent); err != nil {
				return nil, err
			}

			elem := result[itemId]
			elem.contents = fileContent
			elem.foundFileName = child.Name()
			result[itemId] = elem
		}
	}
	var missing []string
	var zeroContents []string

	// Making sure that all the wanted elements were retrieved
	for key, elem := range result {
		if elem.foundFileName == "" {
			missing = append(missing, key)
		}
		if len(elem.contents) == 0 {
			zeroContents = append(zeroContents, key)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("elements %v not found in .ISO '%s'", missing, isoFileName)
	}
	if len(zeroContents) > 0 {
		return nil, fmt.Errorf("elements %v have zero contents in .ISO '%s'", zeroContents, isoFileName)
	}

	return result, nil
}

func (client Client) buildLandingZoneRDE(isoFileName, user, catalogItemId string) (*SolutionEntity, string, error) {

	// ACTIONS:
	// 1. retrieve files from .ISO
	// 2. check certificates
	// 3. if certificate is not in directory, add it
	// 4. encode file contents
	// 5. return the SolutionEntity

	foundFiles, err := getContentsFromIsoFiles(isoFileName, wantedFiles)
	if err != nil {
		return nil, "", fmt.Errorf("error retrieving files from .ISO '%s': %s", isoFileName, err)
	}

	iconContents := foundFiles[iconKey].contents
	iconText := base64.StdEncoding.EncodeToString(iconContents)
	certificateText := foundFiles[certificateKey].contents
	licenseText := foundFiles[eulaKey].contents
	manifestText := foundFiles[manifestKey].contents

	certificates, err := client.GetAllCertificatesFromLibrary(nil)
	if err != nil {
		return nil, "", err
	}

	jsonManifestText, err := yaml.YAMLToJSON(manifestText)
	if err != nil {
		return nil, "", fmt.Errorf("error converting manifest file '%s' to JSON: %s", foundFiles[manifestKey].foundFileName, err)
	}
	var manifestStruct map[string]any
	err = json.Unmarshal(jsonManifestText, &manifestStruct)
	if err != nil {
		return nil, "", fmt.Errorf("error encoding manifest file '%s' to JSON: %s", foundFiles[manifestKey].foundFileName, err)
	}

	vendor, ok := manifestStruct[manifestVendor]
	if !ok {
		return nil, "", fmt.Errorf("missing 'vendor' information from manifest file retrieved from '%s'", isoFileName)
	}
	name, ok := manifestStruct[manifestName]
	if !ok {
		return nil, "", fmt.Errorf("missing 'name' information from manifest file retrieved from '%s'", isoFileName)
	}
	version, ok := manifestStruct[manifestVersion]
	if !ok {
		return nil, "", fmt.Errorf("missing 'version' information from manifest file retrieved from '%s'", isoFileName)
	}
	foundCertificate := false
	for _, existingCert := range certificates {
		if existingCert.CertificateLibrary.Certificate == string(certificateText) {
			foundCertificate = true
		}
	}
	if !foundCertificate {
		addonCertificateName := fmt.Sprintf("addon.%s_%s", vendor, time.Now().Format("2006-01-02T15:04:05.000Z"))

		certificateConfig := types.CertificateLibraryItem{
			Alias:       addonCertificateName,
			Id:          "",
			Certificate: string(certificateText),
			Description: "certificate retrieved from file " + isoFileName,
		}
		_, err = client.AddCertificateToLibrary(&certificateConfig)
		if err != nil {
			return nil, "", fmt.Errorf("error adding certificate '%s' to library", addonCertificateName)
		}
	}

	iconEntry := ""
	fileSuffix := path.Ext(foundFiles[iconKey].foundFileName)

	iconDef, ok := iconTypes[fileSuffix]
	if !ok {
		return nil, "", fmt.Errorf("no icon definition found for file suffix '%s'", fileSuffix)
	}
	iconEntry = fmt.Sprintf("data:%s;base64,%s", iconDef, iconText)

	solutionEntity := &SolutionEntity{
		Eula:     string(licenseText),
		Status:   "READY",
		Manifest: manifestStruct,
		Icon:     iconEntry,
		Origin: SolutionsOrigin{
			Type:          "CATALOG",
			AcceptedBy:    user,
			AcceptedOn:    time.Now().UTC().Format("2000-01-01T10:10:10.000Z"),
			CatalogItemId: catalogItemId,
		},
	}

	solutionRdeName := fmt.Sprintf("%s.%s-%s", vendor, name, version)
	//fmt.Println(iconKey, iconText)
	//fmt.Println(certificateKey, string(certificateText))
	//fmt.Println(eulaKey, string(licenseText))
	//fmt.Println("json manifest", string(jsonManifestText))

	return solutionEntity, solutionRdeName, nil
}

func (client Client) CreateLandingZoneRde(isoFileName, user, catalogId string) (*DefinedEntity, error) {

	//solutionEntity, rdeName, err := client.buildLandingZoneRDE(isoFileName, user, catalogId)
	//if err != nil {
	//	return nil, err
	//}
	// TODO: prepare RDE using solution entity as payload

	return nil, fmt.Errorf("WIP")
}
