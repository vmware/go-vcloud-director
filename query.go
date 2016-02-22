/*
 * Copyright 2016 Skyscape Cloud Services.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	// "net/http/httputil"
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

// func debug(data []byte, err error) {
//     if err == nil {
//         fmt.Printf("%s\n\n", data)
//     } else {
//         log.Fatalf("%s\n\n", err)
//     }
// }

func (c *VCDClient) Query(params map[string]string) (Results, error) {

	req := c.Client.NewRequest(params, "GET", c.QueryHREF, nil)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version=5.5")
	// debug(httputil.DumpRequestOut(req, true))

	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return Results{}, fmt.Errorf("error retreiving query: %s", err)
	}
	// debug(httputil.DumpResponse(resp, true))

	results := NewResults(&c.Client)

	if err = decodeBody(resp, results.Results); err != nil {
		return Results{}, fmt.Errorf("error decoding query results: %s", err)
	}

	return *results, nil
}
