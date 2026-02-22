import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { PostHogQuery, PostHogDataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, PostHogQuery, PostHogDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
