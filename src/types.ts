import { DataQuery, DataSourceJsonData } from '@grafana/data';

export enum MetricType {
  PERF = 'performance',
  AVAIL = 'availability',
}

export interface PulsarQuery extends DataQuery {
  customerID: number;
  appID?: string;
  jobList: string[];
  metricType: MetricType;
}

export const defaultQuery: Partial<PulsarQuery> = {
  customerID: 0,
  appID: '',
  jobList: [],
  metricType: MetricType.PERF,
};

/**
 * These are options configured for each DataSource instance.
 */
export interface PulsarDataSourceOptions extends DataSourceJsonData {
  customerID: number;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface SecureJsonData {
  apiKey?: string;
}
