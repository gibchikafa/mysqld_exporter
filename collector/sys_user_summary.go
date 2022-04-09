// Copyright 2022 Eduardo J. Ortega U.
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

package collector

import (
	"context"
	"database/sql"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const sysUserSummaryQuery = `
	SELECT
		user,
		statement,
        total,
		total_latency,
        max_latency,
        lock_latency,
        rows_sent,
        rows_examined,
        rows_affected,
        full_scans
	FROM
		` + sysSchema + `.x$user_summary_by_statement_type
`

var (
	sysUserSummaryTotalStatements = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sysSchema, "user_total_statements"),
		" The total number of statements for the user",
		[]string{"user", "statement"}, nil)
	sysUserSummaryStatementTotalLatency = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sysSchema, "user_statement_total_latency"),
		"The total wait time of timed statements for the user",
		[]string{"user", "statement"}, nil)
	sysUserSummaryStatementMaxLatency = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sysSchema, "user_statement_max_latency"),
		"The maximum wait time per timed statement for the user",
		[]string{"user", "statement"}, nil)
	sysUserSummaryStatementLockLatency = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sysSchema, "user_statement_lock_latency"),
		"The maximum wait time per timed statement for the user",
		[]string{"user", "statement"}, nil)
	sysUserSummaryStatementRowsSent = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sysSchema, "rows_sent_by_user"),
		"The number of rows sent per timed statement for the user",
		[]string{"user", "statement"}, nil)
	sysUserSummaryStatementRowsExamined = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sysSchema, "rows_examined_by_user"),
		"The number of rows examined per timed statement for the user",
		[]string{"user", "statement"}, nil)
	sysUserSummaryStatementRowsAffected = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sysSchema, "rows_affected_by_user"),
		"The number of rows affected per timed statement for the user",
		[]string{"user", "statement"}, nil)
	sysUserSummaryStatementFullScans = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sysSchema, "full_scans_by_user"),
		"The total full scans per timed statement for the user",
		[]string{"user", "statement"}, nil)
)

type ScrapeSysUserSummary struct{}

// Name of the Scraper. Should be unique.
func (ScrapeSysUserSummary) Name() string {
	return sysSchema + ".user_summary"
}

// Help describes the role of the Scraper.
func (ScrapeSysUserSummary) Help() string {
	return "Collect per user metrics from sys.x$user_summary. See https://dev.mysql.com/doc/refman/5.7/en/sys-user-summary.html for details"
}

// Version of MySQL from which scraper is available.
func (ScrapeSysUserSummary) Version() float64 {
	return 5.7
}

// Scrape the information from sys.user_summary, creating a metric for each value of each row, labeled with the user
func (ScrapeSysUserSummary) Scrape(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric, logger log.Logger) error {

	userSummaryRows, err := db.QueryContext(ctx, sysUserSummaryQuery)
	if err != nil {
		return err
	}
	defer userSummaryRows.Close()

	var (
		user          string
		statement     string
		total         uint64
		total_latency uint64
		max_latency   uint64
		lock_latency  uint64
		rows_sent     uint64
		rows_examined uint64
		rows_affected uint64
		full_scans    uint64
	)

	for userSummaryRows.Next() {
		err = userSummaryRows.Scan(
			&user,
			&statement,
			&total,
			&total_latency,
			&max_latency,
			&lock_latency,
			&rows_sent,
			&rows_examined,
			&rows_affected,
			&full_scans,
		)
		if err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(sysUserSummaryTotalStatements, prometheus.CounterValue, float64(total), user, statement)
		ch <- prometheus.MustNewConstMetric(sysUserSummaryStatementTotalLatency, prometheus.CounterValue, float64(total_latency), user, statement)
		ch <- prometheus.MustNewConstMetric(sysUserSummaryStatementMaxLatency, prometheus.CounterValue, float64(max_latency), user, statement)
		ch <- prometheus.MustNewConstMetric(sysUserSummaryStatementLockLatency, prometheus.CounterValue, float64(lock_latency), user, statement)
		ch <- prometheus.MustNewConstMetric(sysUserSummaryStatementRowsSent, prometheus.CounterValue, float64(rows_sent), user, statement)
		ch <- prometheus.MustNewConstMetric(sysUserSummaryStatementRowsExamined, prometheus.CounterValue, float64(rows_examined), user, statement)
		ch <- prometheus.MustNewConstMetric(sysUserSummaryStatementRowsAffected, prometheus.CounterValue, float64(rows_affected), user, statement)
		ch <- prometheus.MustNewConstMetric(sysUserSummaryStatementFullScans, prometheus.CounterValue, float64(full_scans), user, statement)
	}
	return nil
}

var _ Scraper = ScrapeSysUserSummary{}
