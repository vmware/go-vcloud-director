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

200 - ok

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

This operation requires System administrator privileges

```
POST https://example.eng.vmware.com/network/firewall/vdc/616d7175-e538-431a-a659-f28a30e21540
```

Response: 201 - created

## Disable distributed firewall

This operation requires System administrator privileges

```
DELETE https://example.eng.vmware.com/network/firewall/vdc/616d7175-e538-431a-a659-f28a30e21540
```

Response: 204 - no content

## Add a new rule

```
PUT https://example.com.eng.vmware.com/network/firewall/globalroot-0/config/layer3sections/616d7175-e538-431a-a659-f28a30e21540
```

Payload 1 rule

```xml

<section id="1043" name="616d7175-e538-431a-a659-f28a30e21540" generationNumber="1669055079107">
  <rule logged="false">
    <name>aaaa</name>
    <action>allow</action>
    <appliedToList/>
    <sources excluded="false">
      <source>
        <name>net-datacloud-r</name>
        <value>urn:vcloud:network:895c70bc-c19e-4401-9c9f-42c6bc0f1342</value>
        <type>Network</type>
      </source>
    </sources>
    <services/>
    <direction>in</direction>
    <packetType>any</packetType>
  </rule>
  <rule id="1047" disabled="false" logged="false">
    <name>Default Allow Rule</name>
    <action>deny</action>
    <appliedToList>
      <appliedTo>
        <name>vdc-datacloud</name>
        <value>616d7175-e538-431a-a659-f28a30e21540</value>
        <type>VDC</type>
        <isValid>true</isValid>
      </appliedTo>
    </appliedToList>
    <sectionId>1043</sectionId>
    <direction>inout</direction>
    <packetType>any</packetType>
  </rule>
</section>
```

Payload 3 rules

```xml

<section id="1049" name="616d7175-e538-431a-a659-f28a30e21540" generationNumber="1669270562214">
  <rule id="1055" disabled="false" logged="false">
    <name>abcd</name>
    <action>allow</action>
    <appliedToList>
      <appliedTo>
        <name>gw-datacloud</name>
        <value>urn:vcloud:gateway:01e3934a-a02f-4607-8ce2-a6d3e4e2c45e</value>
        <type>Edge</type>
        <isValid>true</isValid>
      </appliedTo>
      <appliedTo>
        <name>vdc-datacloud</name>
        <value>616d7175-e538-431a-a659-f28a30e21540</value>
        <type>VDC</type>
        <isValid>true</isValid>
      </appliedTo>
    </appliedToList>
    <sectionId>1049</sectionId>
    <sources excluded="false">
      <source>
        <name>net-datacloud-r</name>
        <value>urn:vcloud:network:895c70bc-c19e-4401-9c9f-42c6bc0f1342</value>
        <type>Network</type>
        <isValid>true</isValid>
      </source>
      <source>
        <name>sys-gen-empty-ipset-edge-fw</name>
        <value>616d7175-e538-431a-a659-f28a30e21540:ipset-1</value>
        <type>IPSet</type>
        <isValid>true</isValid>
      </source>
      <source>
        <name>TestVm</name>
        <value>urn:vcloud:vm:1aaf7c5f-cc54-413b-a4f5-96999f664a29</value>
        <type>VirtualMachine</type>
        <isValid>true</isValid>
      </source>
    </sources>
    <destinations excluded="false">
      <destination>
        <name>net-datacloud-r</name>
        <value>urn:vcloud:network:895c70bc-c19e-4401-9c9f-42c6bc0f1342</value>
        <type>Network</type>
        <isValid>true</isValid>
      </destination>
    </destinations>
    <services>
      <service>
        <sourcePort>500</sourcePort>
        <destinationPort>800</destinationPort>
        <protocol>6</protocol>
        <protocolName>TCP</protocolName>
        <isValid>true</isValid>
      </service>
      <service>
        <sourcePort>5000</sourcePort>
        <destinationPort>8000</destinationPort>
        <protocol>6</protocol>
        <protocolName>TCP</protocolName>
        <isValid>true</isValid>
      </service>
      <service>
        <name>PostgreSQL</name>
        <value>application-259</value>
        <type>Application</type>
        <isValid>true</isValid>
      </service>
    </services>
    <direction>inout</direction>
    <packetType>any</packetType>
  </rule>
  <rule id="1056" disabled="false" logged="false">
    <name>xyz</name>
    <action>allow</action>
    <appliedToList>
      <appliedTo>
        <name>vdc-datacloud</name>
        <value>616d7175-e538-431a-a659-f28a30e21540</value>
        <type>VDC</type>
        <isValid>true</isValid>
      </appliedTo>
    </appliedToList>
    <sectionId>1049</sectionId>
    <sources excluded="false">
      <source>
        <value>10.10.10.10</value>
        <type>Ipv4Address</type>
        <isValid>true</isValid>
      </source>
    </sources>
    <direction>in</direction>
    <packetType>ipv6</packetType>
  </rule>
  <rule id="1054" disabled="false" logged="false">
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
    <sectionId>1049</sectionId>
    <direction>inout</direction>
    <packetType>any</packetType>
  </rule>
</section>
```

Response: 200 - ok

1 rule

```xml
<?xml version="1.0" encoding="UTF-8"?>
<section id="1043" name="616d7175-e538-431a-a659-f28a30e21540" generationNumber="1669055519826"
         timestamp="1669055519826" tcpStrict="false" stateless="false" useSid="false" type="LAYER3">
  <rule id="1048" disabled="false" logged="false">
    <name>aaaa</name>
    <action>allow</action>
    <appliedToList>
      <appliedTo>
        <name>vdc-datacloud</name>
        <value>616d7175-e538-431a-a659-f28a30e21540</value>
        <type>VDC</type>
        <isValid>true</isValid>
      </appliedTo>
    </appliedToList>
    <sectionId>1043</sectionId>
    <sources excluded="false">
      <source>
        <name>net-datacloud-1</name>
        <value>urn:vcloud:network:deadbeef-c19e-4401-9c9f-42c6bc0f1342</value>
        <type>Network</type>
        <isValid>true</isValid>
      </source>
      <source>
        <name>net-datacloud-r</name>
        <value>urn:vcloud:network:895c70bc-c19e-4401-9c9f-42c6bc0f1342</value>
        <type>Network</type>
        <isValid>true</isValid>
      </source>
    </sources>
    <direction>in</direction>
    <packetType>any</packetType>
    <tag>616d7175-e538-431a-a659-f28a30</tag>
  </rule>
  <rule id="1047" disabled="false" logged="false">
    <name>Default Allow Rule</name>
    <action>deny</action>
    <appliedToList>
      <appliedTo>
        <name>vdc-datacloud</name>
        <value>616d7175-e538-431a-a659-f28a30e21540</value>
        <type>VDC</type>
        <isValid>true</isValid>
      </appliedTo>
      <appliedTo>
        <name>vdc-datacloud1</name>
        <value>deadbeef-e538-431a-a659-f28a30e21540</value>
        <type>VDC</type>
        <isValid>true</isValid>
      </appliedTo>
    </appliedToList>
    <sectionId>1043</sectionId>
    <direction>inout</direction>
    <packetType>any</packetType>
    <tag>616d7175-e538-431a-a659-f28a30</tag>
  </rule>
</section>
```

3 rules

```xml
<?xml version="1.0" encoding="UTF-8"?>
<firewallConfiguration timestamp="0">
  <contextId>globalroot-0</contextId>
  <layer3Sections>
    <section id="1049" name="616d7175-e538-431a-a659-f28a30e21540" generationNumber="1669270562214"
             timestamp="1669270562214" tcpStrict="false" stateless="false" useSid="false" type="LAYER3">
      <rule id="1055" disabled="false" logged="false">
        <name>abcd</name>
        <action>allow</action>
        <appliedToList>
          <appliedTo>
            <name>gw-datacloud</name>
            <value>urn:vcloud:gateway:01e3934a-a02f-4607-8ce2-a6d3e4e2c45e</value>
            <type>Edge</type>
            <isValid>true</isValid>
          </appliedTo>
          <appliedTo>
            <name>vdc-datacloud</name>
            <value>616d7175-e538-431a-a659-f28a30e21540</value>
            <type>VDC</type>
            <isValid>true</isValid>
          </appliedTo>
        </appliedToList>
        <sectionId>1049</sectionId>
        <sources excluded="false">
          <source>
            <name>net-datacloud-r</name>
            <value>urn:vcloud:network:895c70bc-c19e-4401-9c9f-42c6bc0f1342</value>
            <type>Network</type>
            <isValid>true</isValid>
          </source>
          <source>
            <name>sys-gen-empty-ipset-edge-fw</name>
            <value>616d7175-e538-431a-a659-f28a30e21540:ipset-1</value>
            <type>IPSet</type>
            <isValid>true</isValid>
          </source>
          <source>
            <name>TestVm</name>
            <value>urn:vcloud:vm:1aaf7c5f-cc54-413b-a4f5-96999f664a29</value>
            <type>VirtualMachine</type>
            <isValid>true</isValid>
          </source>
        </sources>
        <destinations excluded="false">
          <destination>
            <name>net-datacloud-r</name>
            <value>urn:vcloud:network:895c70bc-c19e-4401-9c9f-42c6bc0f1342</value>
            <type>Network</type>
            <isValid>true</isValid>
          </destination>
        </destinations>
        <services>
          <service>
            <isValid>true</isValid>
            <sourcePort>500</sourcePort>
            <destinationPort>800</destinationPort>
            <protocol>6</protocol>
            <protocolName>TCP</protocolName>
          </service>
          <service>
            <isValid>true</isValid>
            <sourcePort>5000</sourcePort>
            <destinationPort>8000</destinationPort>
            <protocol>6</protocol>
            <protocolName>TCP</protocolName>
          </service>
          <service>
            <name>PostgreSQL</name>
            <value>application-259</value>
            <type>Application</type>
            <isValid>true</isValid>
          </service>
        </services>
        <direction>inout</direction>
        <packetType>any</packetType>
      </rule>
      <rule id="1056" disabled="false" logged="false">
        <name>xyz</name>
        <action>allow</action>
        <appliedToList>
          <appliedTo>
            <name>vdc-datacloud</name>
            <value>616d7175-e538-431a-a659-f28a30e21540</value>
            <type>VDC</type>
            <isValid>true</isValid>
          </appliedTo>
        </appliedToList>
        <sectionId>1049</sectionId>
        <direction>in</direction>
        <packetType>ipv6</packetType>
        <tag>616d7175-e538-431a-a659-f28a30</tag>
      </rule>
      <rule id="1054" disabled="false" logged="false">
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
        <sectionId>1049</sectionId>
        <direction>inout</direction>
        <packetType>any</packetType>
        <tag>616d7175-e538-431a-a659-f28a30</tag>
      </rule>
    </section>
  </layer3Sections>
  <layer2Sections>
    <section id="1048" name="616d7175-e538-431a-a659-f28a30e21540" generationNumber="1669268780141"
             timestamp="1669268780141" tcpStrict="false" stateless="true" useSid="false" type="LAYER2">
      <rule id="1053" disabled="false" logged="false">
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
        <sectionId>1048</sectionId>
        <direction>inout</direction>
        <packetType>any</packetType>
        <tag>616d7175-e538-431a-a659-f28a30</tag>
      </rule>
    </section>
  </layer2Sections>
</firewallConfiguration>
```

## Retrieve list of services

```
GET https://example.eng.vmware.com/network/services/application/scope/616d7175-e538-431a-a659-f28a30e21540
```
