/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
	"strings"
)

// OpenApiMetadataEntry is a wrapper object for types.OpenApiMetadataEntry
type OpenApiMetadataEntry struct {
	MetadataEntry  *types.OpenApiMetadataEntry
	client         *Client
	Etag           string // Allows concurrent operations with metadata
	href           string // This is the HREF of the given metadata entry
	parentEndpoint string // This is the endpoint of the object that has the metadata entries
}

// ---------------------------------------------------------------------------------------------------------------------
// Specific objects compatible with metadata
// ---------------------------------------------------------------------------------------------------------------------

// GetMetadata returns all the metadata from a DefinedEntity.
// NOTE: The obtained metadata doesn't have ETags, use GetMetadataById or GetMetadataByKey to obtain a ETag for a specific entry.
func (rde *DefinedEntity) GetMetadata() ([]*OpenApiMetadataEntry, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntities
	return getAllOpenApiMetadata(rde.client, endpoint, rde.DefinedEntity.ID, rde.DefinedEntity.Name, "entity", nil)
}

// GetMetadataByKey returns a unique DefinedEntity metadata entry corresponding to the given domain, namespace and key.
// The domain and namespace are only needed when there's more than one entry with the same key.
// This is a more costly operation than GetMetadataById due to ETags, so use that preferred option whenever possible.
func (rde *DefinedEntity) GetMetadataByKey(domain, namespace, key string) (*OpenApiMetadataEntry, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntities
	return getOpenApiMetadataByKey(rde.client, endpoint, rde.DefinedEntity.ID, rde.DefinedEntity.Name, "entity", domain, namespace, key)
}

// GetMetadataById returns a unique DefinedEntity metadata entry corresponding to the given domain, namespace and key.
// The domain and namespace are only needed when there's more than one entry with the same key.
func (rde *DefinedEntity) GetMetadataById(id string) (*OpenApiMetadataEntry, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntities
	return getOpenApiMetadataById(rde.client, endpoint, rde.DefinedEntity.ID, rde.DefinedEntity.Name, "entity", id)
}

// AddMetadata adds metadata to the receiver DefinedEntity.
func (rde *DefinedEntity) AddMetadata(metadataEntry types.OpenApiMetadataEntry) (*OpenApiMetadataEntry, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntities
	return addOpenApiMetadata(rde.client, endpoint, rde.DefinedEntity.ID, metadataEntry)
}

// ---------------------------------------------------------------------------------------------------------------------
// Metadata Entry methods for OpenAPI metadata
// ---------------------------------------------------------------------------------------------------------------------

// Update updates the metadata value from the receiver entry.
// Only the value and persistence of the entry can be updated. Re-create the entry in case you want to modify any of the other fields.
func (entry *OpenApiMetadataEntry) Update(value interface{}, persistent bool) error {
	if entry.MetadataEntry.ID == "" {
		return fmt.Errorf("ID of the receiver metadata entry is empty")
	}

	payload := types.OpenApiMetadataEntry{
		ID:           entry.MetadataEntry.ID,
		IsPersistent: persistent,
		IsReadOnly:   entry.MetadataEntry.IsReadOnly,
		KeyValue: types.OpenApiMetadataKeyValue{
			Domain: entry.MetadataEntry.KeyValue.Domain,
			Key:    entry.MetadataEntry.KeyValue.Key,
			Value: types.OpenApiMetadataTypedValue{
				Value: value,
				Type:  entry.MetadataEntry.KeyValue.Value.Type,
			},
			Namespace: entry.MetadataEntry.KeyValue.Namespace,
		},
	}

	apiVersion, err := entry.client.getOpenApiHighestElevatedVersion(entry.parentEndpoint)
	if err != nil {
		return err
	}

	urlRef, err := url.ParseRequestURI(entry.href)
	if err != nil {
		return err
	}

	headers, err := entry.client.OpenApiPutItemAndGetHeaders(apiVersion, urlRef, nil, payload, entry.MetadataEntry, map[string]string{"If-Match": entry.Etag})
	if err != nil {
		return err
	}
	entry.Etag = headers.Get("Etag")
	return nil
}

// Delete deletes the receiver metadata entry.
func (entry *OpenApiMetadataEntry) Delete() error {
	if entry.MetadataEntry.ID == "" {
		return fmt.Errorf("ID of the receiver metadata entry is empty")
	}

	apiVersion, err := entry.client.getOpenApiHighestElevatedVersion(entry.parentEndpoint)
	if err != nil {
		return err
	}

	urlRef, err := url.ParseRequestURI(entry.href)
	if err != nil {
		return err
	}

	err = entry.client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return err
	}

	entry.Etag = ""
	entry.parentEndpoint = ""
	entry.href = ""
	entry.MetadataEntry = &types.OpenApiMetadataEntry{}
	return nil
}

// ---------------------------------------------------------------------------------------------------------------------
// OpenAPI Metadata private functions
// ---------------------------------------------------------------------------------------------------------------------

// getAllOpenApiMetadata is a generic function to retrieve all metadata from any VCD object using its ID and the given OpenAPI endpoint.
// It supports query parameters to input, for example, filtering options.
func getAllOpenApiMetadata(client *Client, endpoint, objectId, objectName, objectType string, queryParameters url.Values) ([]*OpenApiMetadataEntry, error) {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, fmt.Sprintf("%s/metadata", objectId))
	if err != nil {
		return nil, err
	}

	allMetadata := []*types.OpenApiMetadataEntry{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &allMetadata, nil)
	if err != nil {
		return nil, err
	}

	var filteredMetadata []*types.OpenApiMetadataEntry
	for _, entry := range allMetadata {
		_, err = filterSingleOpenApiMetadataEntry(objectType, objectName, entry, client.IgnoredMetadata)
		if err != nil {
			if strings.Contains(err.Error(), "is being ignored") {
				continue
			}
			return nil, err
		}
		filteredMetadata = append(filteredMetadata, entry)
	}

	// Wrap all type.OpenApiMetadataEntry into OpenApiMetadataEntry types with client
	results := make([]*OpenApiMetadataEntry, len(filteredMetadata))
	for i := range filteredMetadata {
		results[i] = &OpenApiMetadataEntry{
			MetadataEntry:  filteredMetadata[i],
			client:         client,
			href:           fmt.Sprintf("%s/%s", urlRef.String(), filteredMetadata[i].ID),
			parentEndpoint: endpoint,
		}
	}

	return results, nil
}

// getOpenApiMetadataByKey is a generic function to retrieve a unique metadata entry from any VCD object using its domain, namespace and key.
// The domain and namespace are only needed when there's more than one entry with the same key.
func getOpenApiMetadataByKey(client *Client, endpoint, objectId, objectName, objectType string, domain, namespace, key string) (*OpenApiMetadataEntry, error) {
	queryParameters := url.Values{}
	// As for now, the filter only supports filtering by key
	queryParameters.Add("filter", fmt.Sprintf("keyValue.key==%s", key))
	metadata, err := getAllOpenApiMetadata(client, endpoint, objectId, objectName, objectType, queryParameters)
	if err != nil {
		return nil, err
	}

	if len(metadata) == 0 {
		return nil, fmt.Errorf("%s could not find the metadata associated to object %s", ErrorEntityNotFound, objectId)
	}

	// There's more than one entry with same key, the namespace and domain need to be compared to be able to filter.
	if len(metadata) > 1 {
		var filteredMetadata []*OpenApiMetadataEntry
		for _, entry := range metadata {
			if entry.MetadataEntry.KeyValue.Namespace == namespace && entry.MetadataEntry.KeyValue.Domain == domain {
				filteredMetadata = append(filteredMetadata, entry)
			}
		}
		if len(filteredMetadata) > 1 {
			return nil, fmt.Errorf("found %d metadata entries associated to object %s", len(filteredMetadata), objectId)
		}
		// Required to retrieve an ETag
		return getOpenApiMetadataById(client, endpoint, objectId, objectName, objectType, filteredMetadata[0].MetadataEntry.ID)
	}

	// Required to retrieve an ETag
	return getOpenApiMetadataById(client, endpoint, objectId, objectName, objectType, metadata[0].MetadataEntry.ID)
}

// getOpenApiMetadataById is a generic function to retrieve a unique metadata entry from any VCD object using its unique ID.
func getOpenApiMetadataById(client *Client, endpoint, objectId, objectName, objectType, metadataId string) (*OpenApiMetadataEntry, error) {
	if metadataId == "" {
		return nil, fmt.Errorf("input metadata entry ID is empty")
	}

	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, fmt.Sprintf("%s/metadata/%s", objectId, metadataId))
	if err != nil {
		return nil, err
	}

	response := &OpenApiMetadataEntry{
		MetadataEntry:  &types.OpenApiMetadataEntry{},
		client:         client,
		href:           urlRef.String(),
		parentEndpoint: endpoint,
	}

	headers, err := client.OpenApiGetItemAndHeaders(apiVersion, urlRef, nil, response.MetadataEntry, nil)
	if err != nil {
		return nil, err
	}

	_, err = filterSingleOpenApiMetadataEntry(objectType, objectName, response.MetadataEntry, client.IgnoredMetadata)
	if err != nil {
		return nil, err
	}

	response.Etag = headers.Get("Etag")
	return response, nil
}

// addOpenApiMetadata adds one metadata entry to the VCD object with given ID
func addOpenApiMetadata(client *Client, endpoint, objectId string, metadataEntry types.OpenApiMetadataEntry) (*OpenApiMetadataEntry, error) {
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, fmt.Sprintf("%s/metadata", objectId))
	if err != nil {
		return nil, err
	}

	response := &OpenApiMetadataEntry{
		client:         client,
		MetadataEntry:  &types.OpenApiMetadataEntry{},
		parentEndpoint: endpoint,
	}
	headers, err := client.OpenApiPostItemAndGetHeaders(apiVersion, urlRef, nil, metadataEntry, response.MetadataEntry, nil)
	if err != nil {
		return nil, err
	}
	response.Etag = headers.Get("Etag")
	response.href = fmt.Sprintf("%s/%s", urlRef.String(), response.MetadataEntry.ID)
	return response, nil
}

// ---------------------------------------------------------------------------------------------------------------------
// Ignore OpenAPI Metadata feature
// ---------------------------------------------------------------------------------------------------------------------

// normaliseOpenApiMetadata transforms OpenAPI metadata into a normalised structure
func normaliseOpenApiMetadata(objectType, name string, metadataEntry *types.OpenApiMetadataEntry) (*normalisedMetadata, error) {
	return &normalisedMetadata{
		ObjectType: objectType,
		ObjectName: name,
		Key:        metadataEntry.KeyValue.Key,
		Value:      fmt.Sprintf("%v", metadataEntry.KeyValue.Value.Value),
	}, nil
}

// filterSingleOpenApiMetadataEntry filters a single OpenAPI metadata entry depending on the contents of the input ignored metadata slice.
func filterSingleOpenApiMetadataEntry(objectType, objectName string, metadataEntry *types.OpenApiMetadataEntry, metadataToIgnore []IgnoredMetadata) (*types.OpenApiMetadataEntry, error) {
	normalisedEntry, err := normaliseOpenApiMetadata(objectType, objectName, metadataEntry)
	if err != nil {
		return nil, err
	}
	isFiltered := filterSingleGenericMetadataEntry(normalisedEntry, metadataToIgnore)
	if isFiltered {
		return nil, fmt.Errorf("the metadata entry with key '%s' and value '%v' is being ignored", metadataEntry.KeyValue.Key, metadataEntry.KeyValue.Value.Value)
	}
	return metadataEntry, nil
}
