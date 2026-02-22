import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export interface PostHogQuery extends DataQuery {
  queryType: string;
  rawHogQL: string;
}

export const DEFAULT_QUERY: Partial<PostHogQuery> = {
  queryType: 'hogql',
  rawHogQL: `SELECT
  toStartOfDay(timestamp) AS day,
  count() AS event_count
FROM events
WHERE timestamp >= $__timeFrom AND timestamp < $__timeTo
GROUP BY day
ORDER BY day`,
};

export interface PostHogDataSourceOptions extends DataSourceJsonData {
  posthogUrl: string;
  projectId: string;
}

export interface PostHogSecureJsonData {
  apiKey?: string;
}
