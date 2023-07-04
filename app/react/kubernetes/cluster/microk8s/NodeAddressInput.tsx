import { TextArea } from '@@/form-components/Input/Textarea';

type Props = {
  nodeIPValues: string[];
  onChange: (e: React.ChangeEvent<HTMLTextAreaElement>) => void;
};

// this input is part of a formik form. When the text area is changed, the textbox string is separated into an array of strings by new line and comma separators.
export function NodeAddressInput({ nodeIPValues, onChange }: Props) {
  return (
    <div>
      <TextArea
        className="min-h-[150px] resize-y"
        value={nodeIPValues.join('\n')} // display the text area as a string with each new ip address/entry on a new line
        onChange={onChange}
      />
    </div>
  );
}
