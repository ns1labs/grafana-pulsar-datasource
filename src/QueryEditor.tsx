import { defaults } from 'lodash';

import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from './datasource';
import { defaultQuery, PulsarDataSourceOptions, PulsarQuery } from './types';

const { FormField } = LegacyForms;

type Props = QueryEditorProps<DataSource, PulsarQuery, PulsarDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {

  onAppIDChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onRunQuery } = this.props;
    this.setState({appId: event.target.value})
    // Refresh jobID
    onRunQuery();
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { appID } = query;

    return (
      <div className="gf-form">
        <FormField
          width={6}
          value={appID}
          onChange={this.onAppIDChange}
          label="App ID"
          type="string"
        />
      </div>
    );
  }
}
