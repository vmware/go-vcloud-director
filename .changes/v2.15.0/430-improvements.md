* Changed private function `addMetadata()` to allow passing metadata *typedValue* instead of having
  hardcoded `MetadataStringValue`.

* Added to *VM* and *VApp* the methods `DeleteMetadataEntry`, `DeleteMetadataEntryAsync`, `AddMetadataEntry` and `AddMetadataEntryAsync`
  so it follows the same convention as the rest of entities that uses metadata.