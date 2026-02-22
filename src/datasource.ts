import { DataSourceInstanceSettings, CoreApp, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';

import { PostHogQuery, PostHogDataSourceOptions, DEFAULT_QUERY } from './types';

export class DataSource extends DataSourceWithBackend<PostHogQuery, PostHogDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<PostHogDataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<PostHogQuery> {
    return DEFAULT_QUERY;
  }

  applyTemplateVariables(query: PostHogQuery, scopedVars: ScopedVars) {
    return {
      ...query,
      rawHogQL: getTemplateSrv().replace(query.rawHogQL, scopedVars),
    };
  }

  filterQuery(query: PostHogQuery): boolean {
    return !!query.rawHogQL;
  }
}
