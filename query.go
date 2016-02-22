/*
 * Copyright 2014 Skyscape Cloud Services.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	// "net/url"
	// "net/http"
	"net/http/httputil"
	"log"

	types "github.com/hmrc/vmware-govcd/types/v56"
)

type Results struct {
	Results *types.QueryResultRecordsType
	c       *Client	
}

func NewResults(c *Client) *Results {
	return &Results{
		Results: new(types.QueryResultRecordsType),
		c:       c,
	}
}

func debug(data []byte, err error) {
    if err == nil {
        fmt.Printf("%s\n\n", data)
    } else {
        log.Fatalf("%s\n\n", err)
    }
}

func (c *VCDClient) Query(params map[string]string) (Results, error) {

	req := c.Client.NewRequest(params, "GET", c.QueryHREF, nil)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version=5.5")
	debug(httputil.DumpRequestOut(req, true))

	// TODO: wrap into checkresp to parse error
	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return Results{}, fmt.Errorf("error retreiving query: %s", err)
	}
	debug(httputil.DumpResponse(resp, true))

	// org := NewOrg(&c.Client)
	results := NewResults(&c.Client)

	if err = decodeBody(resp, results.Results); err != nil {
		return Results{}, fmt.Errorf("error decoding query results: %s", err)
	}

	// // Get the VDC ref from the Org
	// for _, s := range org.Org.Link {
	// 	if s.Type == "application/vnd.vmware.vcloud.vdc+xml" && s.Rel == "down" {
	// 		if vcdname != "" && s.Name != vcdname {
	// 			continue
	// 		}
	// 		u, err := url.Parse(s.HREF)
	// 		if err != nil {
	// 			return Org{}, err
	// 		}
	// 		c.Client.VCDVDCHREF = *u
	// 	}
	// }

	// if &c.Client.VCDVDCHREF == nil {
	// 	return Org{}, fmt.Errorf("error finding the organization VDC HREF")
	// }

	return *results, nil
}
