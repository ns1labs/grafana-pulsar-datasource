/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import React, { FC } from 'react';
import { GrafanaTheme } from '@grafana/data';
import { useStyles } from '@grafana/ui';
import { css } from '@emotion/css';

export const FieldRowGroup: FC = ({ children }) => {
  const styles = useStyles(getStyles);

  // A simple flex row with spacing among its children
  return <div className={styles}>{children}</div>;
};

const getStyles = (theme: GrafanaTheme) => css`
  width: 100%;
  display: flex;
  > * {
    flex-basis: 100%;
  }
  > * + * {
    margin-left: ${theme.spacing.sm};
  }
`;
