import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput, Select } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps, SelectableValue } from '@grafana/data';
import { PostHogDataSourceOptions, PostHogSecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<PostHogDataSourceOptions, PostHogSecureJsonData> {}

const URL_OPTIONS: Array<SelectableValue<string>> = [
  { label: 'US Cloud', value: 'https://us.posthog.com', description: 'PostHog US Cloud (us.posthog.com)' },
  { label: 'EU Cloud', value: 'https://eu.posthog.com', description: 'PostHog EU Cloud (eu.posthog.com)' },
  { label: 'Custom', value: 'custom', description: 'Self-hosted or custom URL' },
];

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const isCustomUrl = !URL_OPTIONS.some((opt) => opt.value === jsonData.posthogUrl && opt.value !== 'custom');

  const selectedUrlOption = isCustomUrl
    ? URL_OPTIONS.find((opt) => opt.value === 'custom')
    : URL_OPTIONS.find((opt) => opt.value === jsonData.posthogUrl);

  const onUrlOptionChange = (value: SelectableValue<string>) => {
    if (value.value === 'custom') {
      onOptionsChange({
        ...options,
        jsonData: { ...jsonData, posthogUrl: '' },
      });
    } else {
      onOptionsChange({
        ...options,
        jsonData: { ...jsonData, posthogUrl: value.value || '' },
      });
    }
  };

  const onCustomUrlChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: { ...jsonData, posthogUrl: event.target.value },
    });
  };

  const onProjectIdChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: { ...jsonData, projectId: event.target.value },
    });
  };

  const onAPIKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: { apiKey: event.target.value },
    });
  };

  const onResetAPIKey = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: { ...options.secureJsonFields, apiKey: false },
      secureJsonData: { ...options.secureJsonData, apiKey: '' },
    });
  };

  return (
    <>
      <InlineField label="PostHog Instance" labelWidth={20} tooltip="Select your PostHog deployment">
        <Select
          options={URL_OPTIONS}
          value={selectedUrlOption}
          onChange={onUrlOptionChange}
          width={40}
        />
      </InlineField>

      {isCustomUrl && (
        <InlineField label="Custom URL" labelWidth={20} tooltip="Your PostHog instance URL (e.g. https://app.posthog.com)">
          <Input
            onChange={onCustomUrlChange}
            value={jsonData.posthogUrl || ''}
            placeholder="https://your-posthog-instance.com"
            width={40}
          />
        </InlineField>
      )}

      <InlineField label="Project ID" labelWidth={20} tooltip="Your PostHog project ID (found in Project Settings)">
        <Input
          onChange={onProjectIdChange}
          value={jsonData.projectId || ''}
          placeholder="12345"
          width={40}
        />
      </InlineField>

      <InlineField label="API Key" labelWidth={20} tooltip="Personal API key (not the project API key). Create one at Settings → Personal API Keys.">
        <SecretInput
          required
          isConfigured={secureJsonFields.apiKey}
          value={secureJsonData?.apiKey || ''}
          placeholder="phx_..."
          width={40}
          onReset={onResetAPIKey}
          onChange={onAPIKeyChange}
        />
      </InlineField>
    </>
  );
}
