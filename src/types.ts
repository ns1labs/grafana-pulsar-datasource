/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import { DataQuery } from '@grafana/data';

export enum MetricType {
  PERFORMANCE = 'performance',
  AVAILABILITY = 'availability',
}

export enum AggType {
  AVG = 'avg',
  MAX = 'max',
  MIN = 'min',
  P50 = 'p50',
  P75 = 'p75',
  P90 = 'p90',
  P95 = 'p95',
  P99 = 'p99',
}

export enum QueryType {
  INITIAL_APPS_JOBS_FETCH = 'initialAppsJobsFetch',
  REGULAR = 'regular',
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
  agg?: string;
  geo?: string;
  asn?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface SecureJsonData {
  apiKey?: string;
}
