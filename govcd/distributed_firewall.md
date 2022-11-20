# Distributed firewall for NSX-V Virtual Data Centers

**WORK IN PROGRESS**

___
Notes on the implementation. Will be removed upon completion of the PR.
___

## Request distributed firewall status

```
GET https://example.eng.vmware.com/network/firewall/globalroot-0/config?vdc=616d7175-e538-431a-a659-f28a30e21540
```

### Response when disabled

```xml
<?xml version="1.0" encoding="UTF-8"?>
<error>
  <errorCode>400</errorCode>
  <details>Distributed Firewall with section name 616d7175-e538-431a-a659-f28a30e21540 not found; Please see your
    admin.
  </details>
  <rootCauseString>Distributed Firewall with section name 616d7175-e538-431a-a659-f28a30e21540 not found; Please see
    your admin.
  </rootCauseString>
</error>
```

### Response when using a non NSX-V VDC

```xml
<?xml version="1.0" encoding="UTF-8"?>
<error>
  <errorCode>400</errorCode>
  <details>Backing NSX manager example.eng.vmware.com is not a NSX-V manager. Cannot be used for Nsx-V
    api calls.
  </details>
  <rootCauseString>Backing NSX manager example.eng.vmware.com is not a NSX-V manager. Cannot be used for
    Nsx-V api calls.
  </rootCauseString>
</error>
```

### Response when enabled

```xml
<?xml version="1.0" encoding="UTF-8"?>
<firewallConfiguration timestamp="0">
  <contextId>globalroot-0</contextId>
  <layer3Sections>
    <section id="1019" name="616d7175-e538-431a-a659-f28a30e21540" generationNumber="1668964944629"
             timestamp="1668964944629" tcpStrict="false" stateless="false" useSid="false" type="LAYER3">
      <rule id="1023" disabled="false" logged="false">
        <name>Default Allow Rule</name>
        <action>allow</action>
        <appliedToList>
          <appliedTo>
            <name>vdc-datacloud</name>
            <value>616d7175-e538-431a-a659-f28a30e21540</value>
            <type>VDC</type>
            <isValid>true</isValid>
          </appliedTo>
        </appliedToList>
        <sectionId>1019</sectionId>
        <direction>inout</direction>
        <packetType>any</packetType>
        <tag>616d7175-e538-431a-a659-f28a30</tag>
      </rule>
    </section>
  </layer3Sections>
  <layer2Sections>
    <section id="1018" name="616d7175-e538-431a-a659-f28a30e21540" generationNumber="1668964944159"
             timestamp="1668964944159" tcpStrict="false" stateless="true" useSid="false" type="LAYER2">
      <rule id="1022" disabled="false" logged="false">
        <name>Default Allow Rule</name>
        <action>allow</action>
        <appliedToList>
          <appliedTo>
            <name>vdc-datacloud</name>
            <value>616d7175-e538-431a-a659-f28a30e21540</value>
            <type>VDC</type>
            <isValid>true</isValid>
          </appliedTo>
        </appliedToList>
        <sectionId>1018</sectionId>
        <direction>inout</direction>
        <packetType>any</packetType>
        <tag>616d7175-e538-431a-a659-f28a30</tag>
      </rule>
    </section>
  </layer2Sections>
</firewallConfiguration>
```

## Enable distributed firewall

```
POST https://example.eng.vmware.com/network/firewall/vdc/616d7175-e538-431a-a659-f28a30e21540
```

Response: 201 - created

## Disable distributed firewall

```
DELETE https://example.eng.vmware.com/network/firewall/vdc/616d7175-e538-431a-a659-f28a30e21540
```

Response: 204 - no content
