import { useEffect } from 'react';
import { useFormikContext } from 'formik';

import { Option } from '@@/form-components/Input/Select';

// If the value is not found in the new list of available options,
// then set the fieldName to an available value.
export function useSetAvailableOption<T extends string | number>(
  options: Option<T>[] | undefined,
  value: T,
  fieldName: string,
  defaultValue?: string
) {
  const { setFieldValue } = useFormikContext();

  useEffect(() => {
    if (options) {
      if (options.length > 0 && !valueFound(options, value)) {
        setFieldValue(fieldName, options[0].value || '');
      }
    } else if (defaultValue) {
      setFieldValue(fieldName, defaultValue);
    } else {
      setFieldValue(fieldName, '');
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [options, setFieldValue]);
}

function valueFound(
  options: Option<string | number>[],
  value: string | number
) {
  return options.find((option) => option.value === value);
}
