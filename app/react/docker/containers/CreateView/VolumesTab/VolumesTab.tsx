import { InputList } from '@@/form-components/InputList';
import { ArrayError } from '@@/form-components/InputList/InputList';

import { Values, Volume } from './types';
import { InputContext } from './context';
import { Item } from './Item';

export function VolumesTab({
  onChange,
  values,
  allowBindMounts,
  errors,
}: {
  onChange: (values: Values) => void;
  values: Values;
  allowBindMounts: boolean;
  errors?: ArrayError<Values>;
}) {
  const volumes = values;

  return (
    <InputContext.Provider value={allowBindMounts}>
      <InputList
        errors={errors}
        label="Volume mapping"
        onChange={(volumes) => onChange(volumes)}
        value={volumes}
        addLabel="map additional volume"
        item={Item}
        itemBuilder={() =>
          ({
            containerPath: '',
            type: 'volume',
            name: '',
            readOnly: false,
          } as Volume)
        }
      />
    </InputContext.Provider>
  );
}
