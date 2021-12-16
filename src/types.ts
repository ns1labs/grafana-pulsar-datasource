import { DataQuery } from '@grafana/data';

export enum MetricType {
  PERFORMANCE = 'performance',
  AVAILABILITY = 'availability',
  DECISIONS = 'decisions',
}

export interface PulsarApp {
  name: string;
  appid: string;
  jobs?: PulsarJob[];
}

export interface PulsarJob {
  name: string;
  jobid: string;
}

export interface PulsarQuery extends DataQuery {
  appid?: string;
  jobid?: string;
  metricType?: MetricType;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface SecureJsonData {
  apiKey?: string;
}
