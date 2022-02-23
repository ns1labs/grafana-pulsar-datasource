/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import { countries } from 'countries-list';
import { MetricType, AggType, Geo } from './types';

/**
 * Object that maps a display name for each metric type
 */
export const metricTypeDisplayName: Record<MetricType, string> = {
  [MetricType.PERFORMANCE]: 'Performance',
  [MetricType.AVAILABILITY]: 'Availability',
};

/**
 * Object that maps a display name for each aggregation type
 */
export const aggTypeDisplayName: Record<AggType, string> = {
  [AggType.AVG]: 'avg',
  [AggType.MAX]: 'max',
  [AggType.MIN]: 'min',
  [AggType.P50]: 'p50',
  [AggType.P75]: 'p75',
  [AggType.P90]: 'p90',
  [AggType.P95]: 'p95',
  [AggType.P99]: 'p99',
};

/**
 * Returns the US subdivisions (states) with ISO code.
 * Source: https://www.bls.gov/respondents/mwr/electronic-data-interchange/appendix-d-usps-state-abbreviations-and-fips-codes.htm
 * The data was made public domain by the source.
 */
export const getUsSubdivisions = () => {
  return [
    { name: 'Alabama', code: 'AL' },
    { name: 'Alaska', code: 'AK' },
    { name: 'Arizona', code: 'AZ' },
    { name: 'Arkansas', code: 'AR' },
    { name: 'California', code: 'CA' },
    { name: 'Colorado', code: 'CO' },
    { name: 'Connecticut', code: 'CT' },
    { name: 'Delaware', code: 'DE' },
    { name: 'District of Columbia', code: 'DC' },
    { name: 'Florida', code: 'FL' },
    { name: 'Georgia', code: 'GA' },
    { name: 'Hawaii', code: 'HI' },
    { name: 'Idaho', code: 'ID' },
    { name: 'Illinois', code: 'IL' },
    { name: 'Indiana', code: 'IN' },
    { name: 'Iowa', code: 'IA' },
    { name: 'Kansas', code: 'KS' },
    { name: 'Kentucky', code: 'KY' },
    { name: 'Louisiana', code: 'LA' },
    { name: 'Maine', code: 'ME' },
    { name: 'Maryland', code: 'MD' },
    { name: 'Massachusetts', code: 'MA' },
    { name: 'Michigan', code: 'MI' },
    { name: 'Minnesota', code: 'MN' },
    { name: 'Mississippi', code: 'MS' },
    { name: 'Missouri', code: 'MO' },
    { name: 'Montana', code: 'MT' },
    { name: 'Nebraska', code: 'NE' },
    { name: 'Nevada', code: 'NV' },
    { name: 'New Hampshire', code: 'NH' },
    { name: 'New Jersey', code: 'NJ' },
    { name: 'New Mexico', code: 'NM' },
    { name: 'New York', code: 'NY' },
    { name: 'North Carolina', code: 'NC' },
    { name: 'North Dakota', code: 'ND' },
    { name: 'Ohio', code: 'OH' },
    { name: 'Oklahoma', code: 'OK' },
    { name: 'Oregon', code: 'OR' },
    { name: 'Pennsylvania', code: 'PA' },
    { name: 'Puerto Rico', code: 'PR' },
    { name: 'Rhode Island', code: 'RI' },
    { name: 'South Carolina', code: 'SC' },
    { name: 'South Dakota', code: 'SD' },
    { name: 'Tennessee', code: 'TN' },
    { name: 'Texas', code: 'TX' },
    { name: 'Utah', code: 'UT' },
    { name: 'Vermont', code: 'VT' },
    { name: 'Virginia', code: 'VA' },
    { name: 'Virgin Islands', code: 'VI' },
    { name: 'Washington', code: 'WA' },
    { name: 'West Virginia', code: 'WV' },
    { name: 'Wisconsin', code: 'WI' },
    { name: 'Wyoming', code: 'WY' },
  ];
};

/**
 * Returns the Canada subdivisions (provinces/territories) with ISO code.
 * Source: https://www.iso.org/obp/ui/#iso:code:3166:CA
 */
export const getCanadaSubdivisions = () => {
  return [
    { name: 'Alberta', code: 'AB' },
    { name: 'British Columbia', code: 'BC' },
    { name: 'Manitoba', code: 'MB' },
    { name: 'New Brunswick', code: 'NB' },
    { name: 'Newfoundland and Labrador', code: 'NL' },
    { name: 'Northwest Territories', code: 'NT' },
    { name: 'Nova Scotia', code: 'NS' },
    { name: 'Nunavut', code: 'NU' },
    { name: 'Ontario', code: 'ON' },
    { name: 'Prince Edward Island', code: 'PE' },
    { name: 'Quebec', code: 'QC' },
    { name: 'Saskatchewan', code: 'SK' },
    { name: 'Yukon', code: 'YT' },
  ];
};

/**
 * Returns the geo list (countries + US and CA subdivisions).
 * Each geo contains name, code and flag.
 */
export const getGeoList = () => {
  // Get all countries (with name, code and flag)
  const countryList: Geo[] = Object.keys(countries).map((key) => ({
    name: countries[key as keyof typeof countries].name,
    code: key,
    flag: countries[key as keyof typeof countries].emoji,
  }));

  // Get US and Canada subdivisions
  const usSubdivisions: Geo[] = getUsSubdivisions().map((subdivision) => ({
    name: `${subdivision.name} (US)`,
    code: `US_${subdivision.code}`,
    flag: countries.US.emoji,
  }));

  const caSubdivisions: Geo[] = getCanadaSubdivisions().map((subdivision) => ({
    name: `${subdivision.name} (CA)`,
    code: `CA_${subdivision.code}`,
    flag: countries.CA.emoji,
  }));

  const list = countryList.concat(usSubdivisions).concat(caSubdivisions);

  // Sort the list alphabetically
  list.sort((a, b) => {
    const x = a.name.toLowerCase();
    const y = b.name.toLowerCase();

    if (x < y) {
      return -1;
    }

    if (x > y) {
      return 1;
    }

    return 0;
  });

  return list;
};
