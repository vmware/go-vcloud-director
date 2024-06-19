/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"time"

	"github.com/vmware/go-vcloud-director/v2/govcd/internal/udf"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"sigs.k8s.io/yaml"
)

var slzAddOnRdeType = [3]string{"vmware", "solutions_add_on", "1.0.0"}

// SolutionAddOn is the main structure to handle Solution Add-Ons within Solution Landing Zone. It
// packs parent RDE and Solution Add-On entity itself
type SolutionAddOn struct {
	SolutionAddOnEntity *types.SolutionAddOn
	DefinedEntity       *DefinedEntity
	vcdClient           *VCDClient
}

// SolutionAddOnConfig defines configuration for Solution Add-On creation which is used for
// 'VCDClient.CreateSolutionAddOn'.
type SolutionAddOnConfig struct {
	IsoFilePath          string
	User                 string
	CatalogItemId        string
	AutoTrustCertificate bool
}

const (
	manifestKey     = "manifest"
	iconKey         = "icon"
	eulaKey         = "eula"
	certificateKey  = "certificate"
	manifestVersion = "version"
	manifestName    = "name"
	manifestVendor  = "vendor"
)

// wantedFiles are the files expected in solution Add-Ons
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

func createSolutionAddOnValidator(cfg SolutionAddOnConfig) error {
	if cfg.IsoFilePath == "" {
		return fmt.Errorf("'isoFilePath' must be specified")
	}

	if cfg.User == "" {
		return fmt.Errorf("'user' must be specified")
	}

	if cfg.CatalogItemId == "" {
		return fmt.Errorf("'catalogItemId' must be specified")
	}

	return nil
}

// CreateSolutionAddOn creates Solution Add-On instance in VCD based on given
// Requirements - the ISO image defined in `isoFilePath` must already be uploaded to a catalog and
// `catalogItemId` must reflect exactly the same image
//
// Order of operations that this method provides:
// * Get contents of Solution Add-On ISO file defined in 'isoFilePath'
// * Create the 'Entity' payload for creating RDE based on the given image
// * Get Solution Add-On RDE Name from the manifest within 'isoFilePath'
// * If 'autoTrustCertificate' is set to true - the code will check if VCD trusts the certificate
// and trust it if it wasn't already trusted
// * Lookup RDE type 'vmware:solutions_add_on:1.0.0'
// * Create an RDE entity with payload from the 'isoFilePath' contents
func (vcdClient *VCDClient) CreateSolutionAddOn(cfg SolutionAddOnConfig) (*SolutionAddOn, error) {
	err := createSolutionAddOnValidator(cfg)
	if err != nil {
		return nil, err
	}

	foundFiles, err := getContentsFromIsoFiles(cfg.IsoFilePath, wantedFiles)
	if err != nil {
		return nil, fmt.Errorf("error reading contents of '%s': %s", cfg.IsoFilePath, err)
	}

	solutionAddOnEntityRde, err := buildSolutionAddonRdeEntity(foundFiles, "administrator", cfg.CatalogItemId)
	if err != nil {
		return nil, fmt.Errorf("error building Solution Add-On RDE: %s", err)
	}

	rdeName, err := extractSolutionAddonRdeName(foundFiles)
	if err != nil {
		return nil, fmt.Errorf("error finding RDE Entity Name: %s", err)
	}

	if cfg.AutoTrustCertificate {
		certificateText, err := extractSolutionAddOnCertificate(foundFiles)
		if err != nil {
			return nil, fmt.Errorf("error extracting Certificate from Add-On image: %s", err)
		}
		isoFileName := filepath.Base(cfg.IsoFilePath)
		err = vcdClient.TrustAddOnImageCertificate(certificateText, isoFileName)
		if err != nil {
			return nil, fmt.Errorf("certificate trust was request, but it failed: %s", err)
		}
	}

	rdeType, err := vcdClient.GetRdeType(slzAddOnRdeType[0], slzAddOnRdeType[1], slzAddOnRdeType[2])
	if err != nil {
		return nil, fmt.Errorf("error retrieving RDE Type for Solution Add-On: %s", err)
	}

	unmarshalledRdeEntityJson, err := convertAnyToRdeEntity(solutionAddOnEntityRde)
	if err != nil {
		return nil, err
	}

	entityCfg := &types.DefinedEntity{
		EntityType: fmt.Sprintf("urn:vcloud:type:%s:%s:%s", slzAddOnRdeType[0], slzAddOnRdeType[1], slzAddOnRdeType[2]),
		Name:       rdeName,
		State:      addrOf("PRE_CREATED"),
		Entity:     unmarshalledRdeEntityJson,
	}

	createdRdeEntity, err := rdeType.CreateRde(*entityCfg, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating RDE entity: %s", err)
	}

	err = createdRdeEntity.Resolve()
	if err != nil {
		return nil, fmt.Errorf("error resolving Solutions Add-On after creating: %s", err)
	}

	result, err := convertRdeEntityToAny[types.SolutionAddOn](createdRdeEntity.DefinedEntity.Entity)
	if err != nil {
		return nil, err
	}

	returnType := SolutionAddOn{
		SolutionAddOnEntity: result,
		vcdClient:           vcdClient,
		DefinedEntity:       createdRdeEntity,
	}

	return &returnType, nil
}

// GetAllSolutionAddons retrieves all Solution Add-Ons with a given filter
func (vcdClient *VCDClient) GetAllSolutionAddons(queryParameters url.Values) ([]*SolutionAddOn, error) {
	allAddons, err := vcdClient.GetAllRdes(slzAddOnRdeType[0], slzAddOnRdeType[1], slzAddOnRdeType[2], queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all Solution Add-ons: %s", err)
	}

	results := make([]*SolutionAddOn, len(allAddons))
	for index, rde := range allAddons {
		addon, err := convertRdeEntityToAny[types.SolutionAddOn](rde.DefinedEntity.Entity)
		if err != nil {
			return nil, fmt.Errorf("error converting RDE to Solution Add-on: %s", err)
		}

		results[index] = &SolutionAddOn{
			vcdClient:           vcdClient,
			DefinedEntity:       rde,
			SolutionAddOnEntity: addon,
		}
	}

	return results, nil
}

// GetSolutionAddonById retrieves Solution Add-On by ID
func (vcdClient *VCDClient) GetSolutionAddonById(id string) (*SolutionAddOn, error) {
	if id == "" {
		return nil, fmt.Errorf("id must be specified")
	}
	rde, err := getRdeById(&vcdClient.Client, id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Solution Add-On by ID: %s", err)
	}

	result, err := convertRdeEntityToAny[types.SolutionAddOn](rde.DefinedEntity.Entity)
	if err != nil {
		return nil, err
	}

	packages := &SolutionAddOn{
		SolutionAddOnEntity: result,
		vcdClient:           vcdClient,
		DefinedEntity:       rde,
	}

	return packages, nil
}

// GetSolutionAddonByName retrieves Solution Add-Ons by name
// Example name: "vmware.ds-1.4.0-23376809"
func (vcdClient *VCDClient) GetSolutionAddonByName(name string) (*SolutionAddOn, error) {
	if name == "" {
		return nil, fmt.Errorf("name must be specified")
	}

	queryParams := url.Values{}
	queryParams.Add("filter", fmt.Sprintf("name==%s", name))
	results, err := vcdClient.GetAllSolutionAddons(queryParams)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Solution Add-Ons: %s", err)
	}

	return oneOrError("name", name, results)
}

func (s *SolutionAddOn) Update(saoCfg *types.SolutionAddOn) (*SolutionAddOn, error) {
	unmarshalledRdeEntityJson, err := convertAnyToRdeEntity(saoCfg)
	if err != nil {
		return nil, err
	}

	newStructure, err := s.vcdClient.GetSolutionAddonById(s.RdeId())
	if err != nil {
		return nil, fmt.Errorf("error creating a copy of Solution Add-On: %s", err)
	}

	newStructure.DefinedEntity.DefinedEntity.Entity = unmarshalledRdeEntityJson
	err = newStructure.DefinedEntity.Update(*newStructure.DefinedEntity.DefinedEntity)
	if err != nil {
		return nil, err
	}

	newStructure.SolutionAddOnEntity, err = convertRdeEntityToAny[types.SolutionAddOn](s.DefinedEntity.DefinedEntity.Entity)
	if err != nil {
		return nil, err
	}

	return newStructure, nil
}

func (s *SolutionAddOn) Delete() error {
	if s.DefinedEntity == nil {
		return fmt.Errorf("error - parent Defined Entity is nil")
	}
	return s.DefinedEntity.Delete()
}

// RdeId is a shortcut of SolutionEntity.DefinedEntity.DefinedEntity.ID
func (s *SolutionAddOn) RdeId() string {
	if s == nil || s.DefinedEntity == nil || s.DefinedEntity.DefinedEntity == nil {
		return ""
	}

	return s.DefinedEntity.DefinedEntity.ID
}

// TrustAddOnImageCertificate will check if a given certificate is trusted by VCD and trust it if it
// is not there yet
func (vcdClient *VCDClient) TrustAddOnImageCertificate(certificateText, source string) error {
	if certificateText == "" {
		return fmt.Errorf("certificate field is empty")
	}

	if source == "" {
		return fmt.Errorf("source field is empty")
	}

	foundCertificateInLibrary, err := vcdClient.Client.CountMatchingCertificates(certificateText)
	if err != nil {
		return err
	}
	if foundCertificateInLibrary == 0 {
		addonCertificateName := fmt.Sprintf("addon.%s_%s", source, time.Now().Format(time.RFC3339))

		certificateConfig := types.CertificateLibraryItem{
			Alias:       addonCertificateName,
			Certificate: certificateText,
			Description: "certificate retrieved from " + source,
		}
		_, err = vcdClient.Client.AddCertificateToLibrary(&certificateConfig)
		if err != nil {
			return fmt.Errorf("error adding certificate '%s' to library: %s", addonCertificateName, err)
		}
	}

	return nil
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

func buildSolutionAddonRdeEntity(foundFiles map[string]isoFileDef, user, catalogItemId string) (*types.SolutionAddOn, error) {
	iconContents := foundFiles[iconKey].contents
	iconText := base64.StdEncoding.EncodeToString(iconContents)
	licenseText := foundFiles[eulaKey].contents
	manifestText := foundFiles[manifestKey].contents

	jsonManifestText, err := yaml.YAMLToJSON(manifestText)
	if err != nil {
		return nil, fmt.Errorf("error converting manifest file '%s' to JSON: %s", foundFiles[manifestKey].foundFileName, err)
	}
	var manifestStruct map[string]any
	err = json.Unmarshal(jsonManifestText, &manifestStruct)
	if err != nil {
		return nil, fmt.Errorf("error encoding manifest file '%s' to JSON: %s", foundFiles[manifestKey].foundFileName, err)
	}

	iconEntry := ""
	fileSuffix := path.Ext(foundFiles[iconKey].foundFileName)

	iconDef, ok := iconTypes[fileSuffix]
	if !ok {
		return nil, fmt.Errorf("no icon definition found for file suffix '%s'", fileSuffix)
	}
	iconEntry = fmt.Sprintf("data:%s;base64,%s", iconDef, iconText)

	solutionEntity := &types.SolutionAddOn{
		Eula:     string(licenseText),
		Status:   "READY",
		Manifest: manifestStruct,
		Icon:     iconEntry,
		Origin: types.SolutionAddOnOrigin{
			Type:          "CATALOG",
			AcceptedBy:    user,
			AcceptedOn:    time.Now().UTC().Format(time.RFC3339),
			CatalogItemId: catalogItemId,
		},
	}

	return solutionEntity, nil
}

func extractSolutionAddOnCertificate(foundFiles map[string]isoFileDef) (string, error) {
	certificateText := foundFiles[certificateKey].contents
	certificateString := string(certificateText)

	if certificateString == "" {
		certFilter := wantedFiles[certificateKey].namePattern
		return "", fmt.Errorf("%s: no certificate file found based on name filter '%s'", ErrorEntityNotFound, certFilter)
	}

	return certificateString, nil
}

func extractSolutionAddonRdeName(foundFiles map[string]isoFileDef) (string, error) {
	manifestText := foundFiles[manifestKey].contents

	jsonManifestText, err := yaml.YAMLToJSON(manifestText)
	if err != nil {
		return "", fmt.Errorf("error converting manifest file '%s' to JSON: %s", foundFiles[manifestKey].foundFileName, err)
	}
	var manifestStruct map[string]any
	err = json.Unmarshal(jsonManifestText, &manifestStruct)
	if err != nil {
		return "", fmt.Errorf("error encoding manifest file '%s' to JSON: %s", foundFiles[manifestKey].foundFileName, err)
	}

	vendor, ok := manifestStruct[manifestVendor]
	if !ok {
		return "", fmt.Errorf("missing 'vendor' information from manifest file")
	}
	name, ok := manifestStruct[manifestName]
	if !ok {
		return "", fmt.Errorf("missing 'name' information from manifest file")
	}
	version, ok := manifestStruct[manifestVersion]
	if !ok {
		return "", fmt.Errorf("missing 'version' information from manifest file ")
	}
	solutionRdeName := fmt.Sprintf("%s.%s-%s", vendor, name, version)

	return solutionRdeName, nil
}
