* Added new field `HostName` in `QueryResultVMRecordType` struct:
    * String field containing the hostName value from the XML VMAdminRecord body retrieved
    * This info can help to link VM with a cluster since hypervisor is known [GH-503]