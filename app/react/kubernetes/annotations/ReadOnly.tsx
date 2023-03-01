import { Annotation } from './types';

interface Props {
  annotations: Annotation[];
}

export function ReadOnly({ annotations }: Props) {
  if (!annotations.length) {
    return <div className="col-sm-12 p-0">None</div>;
  }
  return (
    <>
      <div className="row font-sm text-muted small font-medium">
        <div className="col-sm-3 col-md-1 p-0">Key</div>
        <div className="col-sm-9 col-md-6 p-0">Value</div>
      </div>
      {annotations.map((a) => (
        <div className="row border-top">
          <div className="col-sm-3 col-md-1 px-0 py-2">{a.Key}</div>
          <div className="col-sm-9 col-md-6 px-0 py-2">{a.Value}</div>
        </div>
      ))}
    </>
  );
}
