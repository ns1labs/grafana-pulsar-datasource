import { MetricType } from './types';

export const metricTypeDisplayName: Record<MetricType, string> = {
  [MetricType.PERFORMANCE]: 'Performance',
  [MetricType.AVAILABILITY]: 'Availability',
  [MetricType.DECISIONS]: 'Decisions',
}
