import React, { PureComponent } from 'react';
import { css } from '@emotion/css';
import { Field, Select } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from './datasource';
import { MetricType, PulsarQuery, PulsarApp } from './types';
import { metricTypeDisplayName } from './utils';

type Props = QueryEditorProps<DataSource, PulsarQuery>;

export class QueryEditor extends PureComponent<Props> {
  componentDidMount() {
    const { data, onRunQuery } = this.props;

    const appJobOptions = data?.series && data?.series[0]?.meta?.custom as PulsarApp[];

    // Run the first query to get the initial apps/jobs options
    // Check if the data already exist, because the query could have been triggered on the page load
    if (!appJobOptions) {
      onRunQuery();
    }
  }

  componentDidUpdate(prevProps: Props) {
    const { query, data, onChange, onRunQuery } = this.props;

    const appJobOptions = data?.series && data?.series[0]?.meta?.custom as PulsarApp[];

    const foundedPulsarApp = appJobOptions?.find((app) => app.appid === query.appid);
    
    const foundedPulsarJob = foundedPulsarApp && 
      foundedPulsarApp.jobs?.find((job) => job.jobid === query.jobid);

    // when the "selected app" is not in the "app options" anymore
    if (query.appid && !foundedPulsarApp) {
      // clear the app and job selection
      onChange({ ...query, appid: undefined, jobid: undefined });
    }

    // when the "selected job" is not in the "job options" anymore
    if (query.jobid && !foundedPulsarJob) {
      // clear the job selection
      onChange({ ...query, jobid: undefined });
    }
    
    // when the 3 dropdowns are selected and at least one has just changed value
    if (query.appid && query.jobid && query.metricType &&
         (prevProps.query.appid != query.appid ||
         prevProps.query.jobid != query.jobid ||
         prevProps.query.metricType != query.metricType)
    ) {
      // run a new query
      onRunQuery();
    }
  }

  render() {
    const { query, data, onChange } = this.props;

    const appJobOptions = data?.series && data?.series[0]?.meta?.custom as PulsarApp[];

    console.log(`...... Props of RefId ${this.props.query.refId}:`, this.props);

    return (
      <div
        className={css`
          width: 100%;
          display: flex;
            > * {
              width: 100%;
            }
            > * + * {
              margin-left: 8px;
            }
        `}
      >
        <Field label="App">
          <Select
            placeholder="Select a Pulsar App"
            options={appJobOptions?.map((app) => ({ label: app.name, value: app.appid }))}
            value={query.appid || null}
            onChange={(option) => onChange({
                ...query,
                appid: option?.value,
                jobid: query.appid !== option?.value ? undefined : query.jobid, // clear job if the app selection has changed
            })}
          />
        </Field>
        <Field label="Job">
          <Select
            placeholder="Select a Pulsar Job"
            options={appJobOptions?.find((app) => app.appid === query.appid)?.jobs
              ?.map((job) => ({ label: job.name, value: job.jobid }))}
            value={query.jobid || null}
            onChange={(option) => onChange({ ...query, jobid: option?.value })}
          />
        </Field> 
        <Field label="Metric">
          <Select
            placeholder="Select a metric type"
            options={Object.keys(metricTypeDisplayName)
              .map((key) => ({ label: metricTypeDisplayName[key as MetricType], value: key }))}
            value={query.metricType || null}
            onChange={(option) => onChange({ ...query, metricType: option?.value as MetricType })}
          />
        </Field> 
      </div>
    );
  }
}
