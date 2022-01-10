/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import React from 'react';
import { Select as GrafanaSelect, SelectCommonProps } from '@grafana/ui';

/*
 * Select
 *  Uses the Grafana Select and its props
 *  Forces the menu position/height (in order to be correctly viewed/navigated in the UI)
 */
export function Select<T>({ value, ...rest }: SelectCommonProps<T>) {
  return (
    <GrafanaSelect
      {...rest}
      value={value || null} // when the value is undefined, it should be turned to null, in order to reset the Select
      menuPosition="fixed"
      maxMenuHeight={200}
    />
  );
}
