import { MetricType, AggType } from './types';

export const metricTypeDisplayName: Record<MetricType, string> = {
  [MetricType.PERFORMANCE]: 'Performance',
  [MetricType.AVAILABILITY]: 'Availability',
  [MetricType.DECISIONS]: 'Decisions',
}

export const aggTypeDisplayName: Record<AggType, string> = {
  [AggType.AVG]: 'avg',
  [AggType.MAX]: 'max',
  [AggType.MIN]: 'min',
  [AggType.P50]: 'p50',
  [AggType.P75]: 'p75',
  [AggType.P90]: 'p90',
  [AggType.P95]: 'p95',
  [AggType.P99]: 'p99',
}
