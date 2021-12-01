import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './ConfigEditor';
import { QueryEditor } from './QueryEditor';
import { PulsarQuery, PulsarDataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, PulsarQuery, PulsarDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
