//go:build ALL || s3 || functional || objectstorage

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"net/http"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_CreateBucket(check *C) {
	// Test case 1: Create a bucket with a valid name
	bucketName := "my-bucket"
	region := "us-west-2"
	response, err := vcd.Client.S3ApiCreateBucket(bucketName, region, nil)
	check.Assert(err, IsNil)
	check.Assert(response.StatusCode, Equals, http.StatusOK)

	// Test case 2: Create a bucket with an invalid name
	invalidBucketName := "invalid-bucket-name"
	invalidRegion := "invalid-region"
	_, err = vcd.Client.S3ApiCreateBucket(invalidBucketName, invalidRegion, nil)
	check.Assert(err, NotNil)

	// Test case 3: Create a bucket with additional headers
	bucketNameWithHeaders := "bucket-with-headers"
	regionWithHeaders := "us-east-1"
	additionalHeaders := map[string]string{
		"X-Custom-Header": "CustomValue",
	}
	response, err = vcd.Client.S3ApiCreateBucket(bucketNameWithHeaders, regionWithHeaders, additionalHeaders)
	check.Assert(err, IsNil)
	check.Assert(response.StatusCode, Equals, http.StatusOK)

	// Test case 4: Create a bucket in a different region
	bucketNameDifferentRegion := "bucket-different-region"
	regionDifferentRegion := "eu-central-1"
	response, err = vcd.Client.S3ApiCreateBucket(bucketNameDifferentRegion, regionDifferentRegion, nil)
	check.Assert(err, IsNil)
	check.Assert(response.StatusCode, Equals, http.StatusOK)
}

func (vcd *TestVCD) Test_DeleteBucket(check *C) {
	// Test case 1: Delete a bucket that exists
	bucketName := "my-bucket"
	region := "us-west-2"
	response, err := vcd.Client.S3ApiCreateBucket(bucketName, region, nil)
	check.Assert(err, IsNil)
	check.Assert(response.StatusCode, Equals, http.StatusOK)

	err = vcd.Client.S3ApiDeleteBucket(bucketName)
	check.Assert(err, IsNil)

	// Test case 2: Delete a bucket that does not exist
	invalidBucketName := "invalid-bucket-name"
	err = vcd.Client.S3ApiDeleteBucket(invalidBucketName)
	check.Assert(err, NotNil)
}

func (vcd *TestVCD) Test_ListBuckets(check *C) {
	// Test case 1: List buckets when there are no buckets
	buckets, err := vcd.Client.S3ApiListBuckets()
	check.Assert(err, IsNil)
	check.Assert(len(buckets), Equals, 0)

	// Test case 2: List buckets when there are buckets
	bucketName := "my-bucket"
	region := "us-west-2"
	response, err := vcd.Client.S3ApiCreateBucket(bucketName, region, nil)
	check.Assert(err, IsNil)
	check.Assert(response.StatusCode, Equals, http.StatusOK)

	buckets, err = vcd.Client.S3ApiListBuckets()
	check.Assert(err, IsNil)
	check.Assert(len(buckets), Equals, 1)
	check.Assert(buckets[0], Equals, bucketName)
}

func (vcd *TestVCD) Test_GetBucket(check *C) {
	// Test case 1: Get a bucket that exists
	bucketName := "my-bucket"
	region := "us-west-2"
	response, err := vcd.Client.S3ApiCreateBucket(bucketName, region, nil)
	check.Assert(err, IsNil)
	check.Assert(response.StatusCode, Equals, http.StatusOK)

	bucket, err := vcd.Client.S3ApiGetBucket(bucketName)
	check.Assert(err, IsNil)
	check.Assert(bucket, Equals, bucketName)

	// Test case 2: Get a bucket that does not exist
	invalidBucketName := "invalid-bucket-name"
	_, err = vcd.Client.S3ApiGetBucket(invalidBucketName)
	check.Assert(err, NotNil)
}
