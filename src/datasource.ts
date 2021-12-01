import { DataSourceInstanceSettings } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { PulsarDataSourceOptions, PulsarQuery } from './types';

export class DataSource extends DataSourceWithBackend<PulsarQuery, PulsarDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<PulsarDataSourceOptions>) {
    super(instanceSettings);
  }
}
