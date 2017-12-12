//  Copyright 2017 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package bqext includes generally useful abstractions for simplifying
// interactions with bigquery.
// Production extensions should go here, but test facilities should go
// in a separate bqtest package.
package bqext

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Dataset provides extensions to the bigquery Dataset and Dataset
// objects to streamline common actions.
// It encapsulates the Client and Dataset to simplify methods.
type Dataset struct {
	BqClient *bigquery.Client
	Dataset  *bigquery.Dataset
}

// NewDataset creates a Dataset for a project.
// httpClient is used to inject mocks for the bigquery client.
// if httpClient is nil, a suitable default client is used.
// Additional bigquery ClientOptions may be optionally passed as final
//   clientOpts argument.  This is useful for testing credentials.
func NewDataset(project, dataset string, clientOpts ...option.ClientOption) (Dataset, error) {
	ctx := context.Background()
	var bqClient *bigquery.Client
	var err error
	bqClient, err = bigquery.NewClient(ctx, project, clientOpts...)

	if err != nil {
		return Dataset{}, err
	}

	return Dataset{bqClient, bqClient.Dataset(dataset)}, nil
}

// ResultQuery constructs a query with common QueryConfig settings for
// writing results to a table.
// Generally, may need to change WriteDisposition.
func (dsExt *Dataset) ResultQuery(query string, dryRun bool) *bigquery.Query {
	q := dsExt.BqClient.Query(query)
	q.QueryConfig.DryRun = dryRun
	if strings.HasPrefix(query, "#legacySQL") {
		q.QueryConfig.UseLegacySQL = true
	}
	// Default for unqualified table names in the query.
	q.QueryConfig.DefaultProjectID = dsExt.Dataset.ProjectID
	q.QueryConfig.DefaultDatasetID = dsExt.Dataset.DatasetID
	return q
}

///////////////////////////////////////////////////////////////////
// Code to execute a single query and parse single row result.
///////////////////////////////////////////////////////////////////

// QueryAndParse executes a query that should return a single row, with
// all struct fields that match query columns filled in.
// The caller must pass in the *address* of an appropriate struct.
// TODO - extend this to also handle multirow results, by passing
// slice of structs.
func (dsExt *Dataset) QueryAndParse(q string, structPtr interface{}) error {
	typeInfo := reflect.ValueOf(structPtr)

	if typeInfo.Type().Kind() != reflect.Ptr {
		return errors.New("Argument should be ptr to struct")
	}
	if reflect.Indirect(typeInfo).Kind() != reflect.Struct {
		return errors.New("Argument should be ptr to struct")
	}

	query := dsExt.ResultQuery(q, false)
	it, err := query.Read(context.Background())
	if err != nil {
		return err
	}

	// We expect a single result row, so proceed accordingly.
	err = it.Next(structPtr)
	if err != nil {
		return err
	}
	var row map[string]bigquery.Value
	// If there are more rows, then something is wrong.
	err = it.Next(&row)
	if err != iterator.Done {
		return errors.New("multiple row data")
	}
	return nil
}

// TableInfo contains the critical stats for a specific table
// or partition.
type TableInfo struct {
	Name             string
	IsPartitioned    bool
	NumBytes         int64
	NumRows          uint64
	CreationTime     time.Time
	LastModifiedTime time.Time
}

// PartitionInfo provides basic information about a partition.
type PartitionInfo struct {
	PartitionID  string
	CreationTime time.Time
	LastModified time.Time
}

// GetPartitionInfo provides basic information about a partition.
func (dsExt Dataset) GetPartitionInfo(table string, partition string) (PartitionInfo, error) {
	// This uses legacy, because PARTITION_SUMMARY is not supported in standard.
	queryString := fmt.Sprintf(
		`#legacySQL
		SELECT
		  partition_id as PartitionID,
		  msec_to_timestamp(creation_time) AS CreationTime,
		  msec_to_timestamp(last_modified_time) AS LastModified
		FROM
		  [%s$__PARTITIONS_SUMMARY__]
		where partition_id = "%s" `, table, partition)
	pi := PartitionInfo{}

	err := dsExt.QueryAndParse(queryString, &pi)
	if err != nil {
		log.Println(err)
		return PartitionInfo{}, err
	}
	return pi, nil
}

// DestinationQuery constructs a query with common Config settings for
// writing results to a table.
// Generally, may need to change WriteDisposition.
func (dsExt *Dataset) DestinationQuery(query string, dest *bigquery.Table) *bigquery.Query {
	q := dsExt.BqClient.Query(query)
	if dest != nil {
		q.QueryConfig.Dst = dest
	} else {
		q.QueryConfig.DryRun = true
	}
	q.QueryConfig.AllowLargeResults = true
	// Default for unqualified table names in the query.
	q.QueryConfig.DefaultProjectID = dsExt.Dataset.ProjectID
	q.QueryConfig.DefaultDatasetID = dsExt.Dataset.DatasetID
	q.QueryConfig.DisableFlattenedResults = true
	return q
}

///////////////////////////////////////////////////////////////////
// Specific queries.
///////////////////////////////////////////////////////////////////

// TODO - really should take the one that was parsed last, instead
// of random
var dedupTemplate = `
	#standardSQL
	# Delete all duplicate rows based on test_id
	SELECT * except (row_number)
	FROM (
	  SELECT *, ROW_NUMBER() OVER (PARTITION BY %s) row_number
	  FROM ` + "`%s`" + `)
	WHERE row_number = 1`

// Dedup executes a query that dedups and writes to an appropriate
// partition.
// src is relative to the project:dataset of dsExt.
// dedupOn names the field to be used for dedupping.
// project, dataset, table specify the table to write into.
// NOTE: destination table must include the partition suffix.  This
// avoids accidentally overwriting TODAY's partition.
func (dsExt *Dataset) Dedup(src string, dedupOn string, overwrite bool, project, dataset, table string) (*bigquery.JobStatus, error) {
	if !strings.Contains(table, "$") {
		return nil, errors.New("Destination table does not specify partition")
	}
	queryString := fmt.Sprintf(dedupTemplate, dedupOn, src)
	dest := dsExt.BqClient.DatasetInProject(project, dataset)
	q := dsExt.DestinationQuery(queryString, dest.Table(table))
	if overwrite {
		q.QueryConfig.WriteDisposition = bigquery.WriteTruncate
	}
	log.Printf("Removing dups (of %s) and writing to %s\n", dedupOn, table)
	job, err := q.Run(context.Background())
	if err != nil {
		return nil, err
	}
	log.Println("JobID:", job.ID())
	status, err := job.Wait(context.Background())
	if err != nil {
		return status, err
	}
	return status, nil
}
