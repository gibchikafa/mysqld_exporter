// Copyright 2019, 2020 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Scrape `ndbinfo.counters.tc`

package collector

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

const ndbinfoCountersTCQuery = `
	SELECT node_id, counter_name, sum(val)
        FROM ndbinfo.counters
        WHERE block_name = "DBTC" and counter_name != "ATTRINFO"
        GROUP BY node_id, counter_name
	`
var (
	ndbinfoCountersTCDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "tc_counter"),
		"Event counters for simple operations",
		[]string{"nodeID", "counterName"}, nil,
	)
)

// ScrapeNdbinfoCountersTC collects for `ndbinfo.counters.tc`
type ScrapeNdbinfoCountersTC struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNdbinfoCountersTC) Name() string {
	return "ndbinfo.counters.tc"
}

// Help describes the role of the Scraper
func (ScrapeNdbinfoCountersTC) Help() string {
	return "Collect metrics from ndbinfo.counters.tc"
}

// Version of MySQL from which scraper is available
func (ScrapeNdbinfoCountersTC) Version() float64 {
	return 5.6
}

// Scrape collects data from database connection and sends it over channel as prometheus metric
func (ScrapeNdbinfoCountersTC) Scrape(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	ndbinfoCountersTCRows, err := db.QueryContext(ctx, ndbinfoCountersTCQuery)
	if err != nil {
		return err
	}
	defer ndbinfoCountersTCRows.Close()

	var (
		nodeID, val                         uint64
		counter_name                        string
	)

	// Iterate over the memory settings
	for ndbinfoCountersTCRows.Next() {
		if err := ndbinfoCountersTCRows.Scan(
			&nodeID, &counter_name, &val); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoCountersTCDesc, prometheus.GaugeValue, float64(val),
			strconv.FormatUint(nodeID, 10), counter_name)
	}
	return nil
}