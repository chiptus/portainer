import { useState, useEffect } from 'react';
import { Info } from 'lucide-react';

import { RegistryViewModel } from '@/portainer/models/registry';

import { Select } from '@@/form-components/ReactSelect';
import { FormControl } from '@@/form-components/FormControl';
import { Button } from '@@/buttons';
import { Tooltip } from '@@/Tip/Tooltip';
import { Icon } from '@@/Icon';
import { FormError } from '@@/form-components/FormError';

interface Props {
  value: number;
  registries: RegistryViewModel[];
  onChange: () => void;
  formInvalid?: boolean;
  errorMessage: string;
  onSelect: () => void;
  isActive?: boolean;
  clearRegistries: () => void;
  method: string;
}

export function PrivateRegistryFieldset({
  value,
  registries,
  onChange,
  formInvalid,
  errorMessage,
  onSelect,
  isActive,
  clearRegistries,
  method,
}: Props) {
  const [checked, setChecked] = useState(isActive);
  const [selected, setSelected] = useState(value);

  const tooltipMessage =
    'Use this when using a private registry that requires credentials';

  useEffect(() => {
    if (checked) {
      onChange();
    } else {
      clearRegistries();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [checked]);

  useEffect(() => {
    setSelected(value);
  }, [value]);

  function reload() {
    onChange();
    setSelected(value);
  }

  function handleChange() {
    setChecked(!checked);
  }

  return (
    <>
      <div className="col-sm-12 form-section-title"> Registry </div>
      <div className="form-group">
        <div className="col-sm-12 vertical-center">
          <label
            className="mr-5 control-label text-left !pt-0"
            htmlFor="toggle-registry-slider"
          >
            Use Credentials <Tooltip message={tooltipMessage} />
          </label>
          <label className="switch mb-0">
            <input
              type="checkbox"
              name="toggle_registry"
              onChange={handleChange}
              disabled={formInvalid}
              defaultChecked={isActive}
              id="toggle-registry-slider"
            />
            <span className="slider round" />
          </label>
        </div>
      </div>
      {checked && (
        <>
          {method !== 'repository' && (
            <>
              <p className="text-muted small">
                <Icon icon={Info} mode="primary" />
                If you make any changes to the image urls in your yaml, please
                reload or select registry manually
              </p>
              <Button onClick={reload}>Reload</Button>
            </>
          )}
          {!errorMessage && (
            <FormControl label="Registry" inputId="users-selector">
              <Select
                value={registries.filter(
                  (registry) => registry.Id === selected
                )}
                options={registries}
                getOptionLabel={(registry) => registry.Name}
                getOptionValue={(registry) => registry.Id}
                onChange={onSelect}
              />
            </FormControl>
          )}
          {errorMessage && <FormError>{errorMessage}</FormError>}
        </>
      )}
    </>
  );
}
