import React from 'react';
import { Stack, InlineField, CodeEditor, Alert } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { PostHogDataSourceOptions, PostHogQuery } from '../types';

type Props = QueryEditorProps<DataSource, PostHogQuery, PostHogDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onHogQLChange = (value: string) => {
    onChange({ ...query, rawHogQL: value });
  };

  const onEditorBlur = () => {
    onRunQuery();
  };

  return (
    <Stack direction="column" gap={1}>
      <InlineField label="HogQL Query" labelWidth={16} grow tooltip="Write your HogQL query">
        <CodeEditor
          language="sql"
          value={query.rawHogQL || ''}
          height={200}
          showLineNumbers={true}
          showMiniMap={false}
          onBlur={onEditorBlur}
          onSave={onEditorBlur}
          onEditorDidMount={(editor) => {
            editor.onDidChangeModelContent(() => {
              onHogQLChange(editor.getValue());
            });
          }}
        />
      </InlineField>

      <Alert title="Available macros" severity="info">
        <ul>
          <li><code>$__timeFrom</code> — Start of the selected time range (e.g. &apos;2024-01-15 10:30:00&apos;)</li>
          <li><code>$__timeTo</code> — End of the selected time range</li>
        </ul>
      </Alert>
    </Stack>
  );
}
