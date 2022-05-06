import { useState, useEffect } from 'react';

import { Select } from '@/portainer/components/form-components/ReactSelect';
import { FormControl } from '@/portainer/components/form-components/FormControl';
import { Button } from '@/portainer/components/Button';
import { RegistryViewModel } from '@/portainer/models/registry';
import { Tooltip } from '@/portainer/components/Tip/Tooltip';
import './PrivateRegistryView.css';

interface Props {
  value: number;
  registries: RegistryViewModel[];
  onChange: () => void;
  forminvalid?: boolean;
  errorMessage: string;
  onSelect: () => void;
  isActive?: boolean;
  clearRegistries: () => void;
  method: string;
}

export function PrivateRegistryView({
  value,
  registries,
  onChange,
  forminvalid,
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
        <div className="col-sm-12">
          <label className="registry-label control-label text-left">
            {' '}
            Use Credentials <Tooltip message={tooltipMessage} />
          </label>
          <label className="switch">
            {' '}
            <input
              type="checkbox"
              name="toggle_registry"
              onChange={handleChange}
              disabled={forminvalid}
              defaultChecked={isActive}
            />
            <i />{' '}
          </label>
        </div>
      </div>
      {checked && (
        <>
          {method !== 'repository' && (
            <>
              <p className="text-muted small">
                <i
                  className="fa fa-info-circle blue-icon space-right"
                  aria-hidden="true"
                />
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
          {errorMessage && (
            <div className="error-message col-sm-12 small text-warning">
              <i className="fa fa-exclamation-triangle" />
              {errorMessage}
            </div>
          )}
        </>
      )}
    </>
  );
}
