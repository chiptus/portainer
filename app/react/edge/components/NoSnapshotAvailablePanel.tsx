import { Widget } from '@@/Widget';

export function NoSnapshotAvailablePanel() {
  return (
    <div className="row">
      <div className="col-sm-12">
        <Widget>
          <Widget.Body>No snapshot available for this environment.</Widget.Body>
        </Widget>
      </div>
    </div>
  );
}
