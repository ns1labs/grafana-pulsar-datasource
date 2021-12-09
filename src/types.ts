import { DataQuery } from '@grafana/data';

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
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface SecureJsonData {
  apiKey?: string;
}
