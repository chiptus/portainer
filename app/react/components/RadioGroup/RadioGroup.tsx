import { useState, ChangeEvent } from 'react';

interface Props {
  options: { value: string; label: string }[];
  selectedOption: string;
  name: string;
  onOptionChange: (value: string) => void;
}

export function RadioGroup({
  options,
  selectedOption,
  name,
  onOptionChange,
}: Props) {
  const [selected, setSelected] = useState(selectedOption);

  function handleOptionChange(event: ChangeEvent<HTMLInputElement>) {
    setSelected(event.target.value);
    onOptionChange(event.target.value);
  }

  return (
    <div>
      {options.map((option) => (
        <span
          key={option.value}
          className="col-sm-3 col-lg-2 control-label !p-0 text-left"
        >
          <input
            type="radio"
            name={name}
            value={option.value}
            checked={selected === option.value}
            onChange={handleOptionChange}
            style={{ margin: '0 4px 0 0' }}
          />
          {option.label}
        </span>
      ))}
    </div>
  );
}
