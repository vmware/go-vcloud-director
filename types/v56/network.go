package types

type FirewallConfiguration struct {
	ContextID      string `xml:"contextId"`
	Layer2Sections struct {
		Section struct {
			GenerationNumber int    `xml:"generationNumber,attr"`
			ID               int    `xml:"id,attr"`
			Name             string `xml:"name,attr"`
			Stateless        bool   `xml:"stateless,attr"`
			TcpStrict        bool   `xml:"tcpStrict,attr"`
			Timestamp        int    `xml:"timestamp,attr"`
			Type             string `xml:"type,attr"`
			UseSid           bool   `xml:"useSid,attr"`
			Rule             struct {
				Disabled      bool   `xml:"disabled,attr"`
				ID            int    `xml:"id,attr"`
				Logged        bool   `xml:"logged,attr"`
				Action        string `xml:"action"`
				AppliedToList struct {
					AppliedTo struct {
						IsValid bool   `xml:"isValid"`
						Name    string `xml:"name"`
						Type    string `xml:"type"`
						Value   string `xml:"value"`
					} `xml:"appliedTo"`
				} `xml:"appliedToList"`
				Direction  string `xml:"direction"`
				Name       string `xml:"name"`
				PacketType string `xml:"packetType"`
				SectionID  int    `xml:"sectionId"`
				Tag        string `xml:"tag"`
			} `xml:"rule"`
		} `xml:"section"`
	} `xml:"layer2Sections"`
	Layer3Sections struct {
		Section struct {
			GenerationNumber int    `xml:"generationNumber,attr"`
			ID               int    `xml:"id,attr"`
			Name             string `xml:"name,attr"`
			Stateless        bool   `xml:"stateless,attr"`
			TcpStrict        bool   `xml:"tcpStrict,attr"`
			Timestamp        int    `xml:"timestamp,attr"`
			Type             string `xml:"type,attr"`
			UseSid           bool   `xml:"useSid,attr"`
			Rule             []struct {
				Disabled      bool   `xml:"disabled,attr"`
				ID            int    `xml:"id,attr"`
				Logged        bool   `xml:"logged,attr"`
				Action        string `xml:"action"`
				AppliedToList struct {
					AppliedTo []struct {
						IsValid bool   `xml:"isValid"`
						Name    string `xml:"name"`
						Type    string `xml:"type"`
						Value   string `xml:"value"`
					} `xml:"appliedTo"`
				} `xml:"appliedToList"`
				Destinations *struct {
					Excluded    bool `xml:"excluded,attr"`
					Destination struct {
						IsValid bool   `xml:"isValid"`
						Name    string `xml:"name"`
						Type    string `xml:"type"`
						Value   string `xml:"value"`
					} `xml:"destination"`
				} `xml:"destinations"`
				Direction  string `xml:"direction"`
				Name       string `xml:"name"`
				PacketType string `xml:"packetType"`
				SectionID  int    `xml:"sectionId"`
				Services   *struct {
					Service []struct {
						DestinationPort *int    `xml:"destinationPort"`
						IsValid         bool    `xml:"isValid"`
						Name            string  `xml:"name"`
						Protocol        *int    `xml:"protocol"`
						ProtocolName    *string `xml:"protocolName"`
						SourcePort      *int    `xml:"sourcePort"`
						Type            string  `xml:"type"`
						Value           string  `xml:"value"`
					} `xml:"service"`
				} `xml:"services"`
				Sources *struct {
					Excluded bool `xml:"excluded,attr"`
					Source   []struct {
						IsValid bool   `xml:"isValid"`
						Name    string `xml:"name"`
						Type    string `xml:"type"`
						Value   string `xml:"value"`
					} `xml:"source"`
				} `xml:"sources"`
				Tag string `xml:"tag"`
			} `xml:"rule"`
		} `xml:"section"`
	} `xml:"layer3Sections"`
}

/*
type DWAppliedTo struct {
	Name    string `xml:"name"`
	Value   string `xml:"value"`
	Type    string `xml:"type"`
	IsValid string `xml:"isValid"`
}

type DWAppliedToList struct {
	AppliedTo []DWAppliedTo `xml:"appliedTo"`
}

type DWSource struct {
	Name    string `xml:"name"`
	Value   string `xml:"value"`
	Type    string `xml:"type"`
	IsValid string `xml:"isValid"`
}

type DWService struct {
}
type DWSources struct {
	Source []DWSource `xml:"source"`
}

type DWServices struct {
	Source []DWService `xml:"service"`
}
type DWRule struct {
	ID            string           `xml:"id,attr"`       // The rule identifier - it is usually a bare number, presented as a string
	Name          string           `xml:"name"`          // The name of the rule, as provided by the user
	Disabled      bool             `xml:"disabled,attr"` // If true, the rule is preserved, but not used
	Logged        bool             `xml:"logged,attr"`   // If true, the rule usage is logged
	Action        string           `xml:"action"`        // allow or deny
	AppliedToList *DWAppliedToList `xml:"appliedToList"` // To which objects the rule applies
	SectionId     string           `xml:"sectionId"`     // To which section the rule belongs
	Sources       *DWSources       `xml:"sources"`       // List of the sources for this rule
	Services      *DWServices      `xml:"services"`      // List of the services for this rule
	Direction     string           `xml:"direction"`     // in, out, or inout
	PacketType    string           `xml:"packetType"`    // any, IPV4, IPV6
	Tag           string           `xml:"tag"`
}

type DWSection struct {
	XMLName          xml.Name `xml:"section"`
	ID               string   `xml:"id,attr"`               // The section identifier
	Name             string   `xml:"name,attr"`             // The name of the section - It is the UUID of the VDC ID
	GenerationNumber string   `xml:"generationNumber,attr"` // read-only : it's the Etag of the latest operation
	Timestamp        string   `xml:"timestamp,attr"`        // read-only - the Unix timestamp of the rule
	Type             string   `xml:"type,attr"`             // either LAYER2 or LAYER3
	TcpStrict        bool     `xml:"tcpStrict,attr"`
	Stateless        bool     `xml:"stateless,attr"`
	UseSid           bool     `xml:"useSid,attr"`
	Rule             []DWRule `xml:"rule"`
}

type Section struct {
	Rule []struct {
		Disabled      bool   `xml:"disabled,attr"`
		ID            int    `xml:"id,attr"`
		Logged        bool   `xml:"logged,attr"`
		Action        string `xml:"action"`
		AppliedToList struct {
			AppliedTo []struct {
				IsValid bool   `xml:"isValid"`
				Name    string `xml:"name"`
				Type    string `xml:"type"`
				Value   string `xml:"value"`
			} `xml:"appliedTo"`
		} `xml:"appliedToList"`
		Direction  string `xml:"direction"`
		Name       string `xml:"name"`
		PacketType string `xml:"packetType"`
		SectionID  int    `xml:"sectionId"`
		Sources    *struct {
			Excluded bool `xml:"excluded,attr"`
			Source   []struct {
				IsValid bool   `xml:"isValid"`
				Name    string `xml:"name"`
				Type    string `xml:"type"`
				Value   string `xml:"value"`
			} `xml:"source"`
		} `xml:"sources"`
		Tag string `xml:"tag"`
	} `xml:"rule"`
}


*/
