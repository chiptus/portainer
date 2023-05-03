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
        <div className="col-sm-4 p-0">Key</div>
        <div className="col-sm-8 p-0">Value</div>
      </div>
      {annotations.map((a) => (
        <div className="row border-top" key={a.Key}>
          <div className="col-sm-4 px-0 py-2">{a.Key}</div>
          <div className="col-sm-8 px-0 py-2">{a.Value}</div>
        </div>
      ))}
    </>
  );
}
