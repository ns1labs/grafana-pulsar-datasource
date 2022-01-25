/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import React, { PureComponent } from 'react';
import { Field, Input } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';

import { DataSource } from './datasource';
import { MetricType, QueryType, PulsarQuery, PulsarApp, AggType, Geo } from './types';
import { metricTypeDisplayName, aggTypeDisplayName, getGeoList } from './utils';

import { FieldRowGroup, Select } from './commons';

type Props = QueryEditorProps<DataSource, PulsarQuery>;

interface State {
  geoList: Geo[];
}

export class QueryEditor extends PureComponent<Props, State> {
  constructor(props: Props) {
    super(props);

    this.state = {
      geoList: getGeoList(),
    };
  }

  componentDidMount() {
    const { data, query, onChange, onRunQuery } = this.props;

    const appJobOptions = (data?.series && data.series[0]?.meta?.custom) as PulsarApp[] | undefined;

    // Check if the first fetch of apps/jobs was already done.
    // If not, a "first query" need to be triggered, because the apps/jobs come in the query result.
    // Usually, the dashboards automatically trigger a query on the page load or when a new QueryEditor is added.
    // It's needed for only the Explore tab and for the editing mode of the panel, because they don't automatically trigger the first query.
    if (!appJobOptions) {
      // The query is not really triggered by the plugin if the query object have only default keys filled.
      // So, we are adding a "queryType" just to have a new key in the query object. It can even be ignored by the backend.
      onChange({ ...query, queryType: QueryType.INITIAL_APPS_JOBS_FETCH });
      onRunQuery(); // it can be called just after the onChange (no need to wait a render cycle to have the props updated)
    }
  }

  componentDidUpdate(prevProps: Props) {
    const { query, data, onChange, onRunQuery } = this.props;

    const appJobOptions = (data?.series && data.series[0]?.meta?.custom) as PulsarApp[] | undefined;

    const foundedPulsarApp = appJobOptions?.find((app) => app.appid === query.appid);

    const foundedPulsarJob = foundedPulsarApp && foundedPulsarApp.jobs?.find((job) => job.jobid === query.jobid);

    // When the queryType is "initial fetch"
    if (query.queryType === QueryType.INITIAL_APPS_JOBS_FETCH) {
      // clear the type
      onChange({ ...query, queryType: undefined });
    }

    // When the list of apps/jobs is available (has been loaded)
    // The following operations should be performed only after the list exists
    if (appJobOptions) {
      // When the "selected app" is not in the "app options" anymore
      if (query.appid && !foundedPulsarApp) {
        // clear the app and job selection
        onChange({ ...query, appid: undefined, jobid: undefined });
      }

      // When the "selected job" is not in the "job options" anymore
      if (query.jobid && !foundedPulsarJob) {
        // clear the job selection
        onChange({ ...query, jobid: undefined });
      }
    }

    // When the 4 mandatory dropdowns are selected and at least one field had its value just changed
    if (
      query.appid &&
      query.jobid &&
      query.metricType &&
      query.agg &&
      (prevProps.query.appid !== query.appid ||
        prevProps.query.jobid !== query.jobid ||
        prevProps.query.metricType !== query.metricType ||
        prevProps.query.agg !== query.agg ||
        prevProps.query.geo !== query.geo ||
        prevProps.query.asn !== query.asn)
    ) {
      // run a new query
      onRunQuery();
    }
  }

  render() {
    const { query, data, onChange } = this.props;
    const { geoList } = this.state;

    const appJobOptions = (data?.series && data.series[0]?.meta?.custom) as PulsarApp[] | undefined;

    return (
      <div>
        <FieldRowGroup>
          <Field label="App">
            <Select
              placeholder="Select a Pulsar App"
              options={appJobOptions?.map((app) => ({
                label: `${app.name} (${app.appid})`,
                value: app.appid,
              }))}
              value={query.appid || null}
              onChange={(option) =>
                onChange({
                  ...query,
                  appid: option?.value,
                  jobid: query.appid !== option?.value ? undefined : query.jobid, // clear job if the app selection has changed
                })
              }
              isLoading={!appJobOptions}
            />
          </Field>
          <Field label="Job">
            <Select
              placeholder="Select a Pulsar Job"
              options={appJobOptions
                ?.find((app) => app.appid === query.appid)
                ?.jobs?.map((job) => ({
                  label: `${job.name} (${job.jobid})`,
                  value: job.jobid,
                }))}
              value={query.jobid || null}
              onChange={(option) => onChange({ ...query, jobid: option?.value })}
              isLoading={!appJobOptions}
            />
          </Field>
          <Field label="Metric">
            <Select
              placeholder="Select a metric type"
              options={Object.keys(metricTypeDisplayName).map((key) => ({
                label: metricTypeDisplayName[key as MetricType],
                value: key,
              }))}
              value={query.metricType || null}
              onChange={(option) => onChange({ ...query, metricType: option?.value as MetricType })}
            />
          </Field>
        </FieldRowGroup>
        <FieldRowGroup>
          <Field label="Aggregation">
            <Select
              placeholder="Select an agg"
              options={Object.keys(aggTypeDisplayName).map((key) => ({
                label: aggTypeDisplayName[key as AggType],
                value: key,
              }))}
              value={query.agg || null}
              onChange={(option) => onChange({ ...query, agg: option?.value as AggType })}
            />
          </Field>
          <Field label="Geo">
            <Select
              placeholder="Select a geo (leave it blank for all geo)"
              options={geoList.map((geo) => ({
                label: `${geo.flag} ${geo.name}`,
                value: geo.code,
              }))}
              value={query.geo || null}
              onChange={(option) =>
                onChange({
                  ...query,
                  geo: option?.value,
                  asn: option?.value ? query.asn : undefined,
                })
              }
              isClearable
            />
          </Field>
          <Field label="ASN" disabled={!query.geo}>
            <Input
              placeholder={!query.geo ? 'Select a geo to filter by ASN' : 'For all ASNs, leave it blank or with a *'}
              value={query.asn || ''}
              onChange={(event) => onChange({ ...query, asn: event.currentTarget.value || undefined })}
            />
          </Field>
        </FieldRowGroup>
      </div>
    );
  }
}
