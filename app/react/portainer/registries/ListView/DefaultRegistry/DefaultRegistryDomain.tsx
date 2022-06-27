import clsx from 'clsx';
import { useSettings } from 'Portainer/settings/queries';

export function DefaultRegistryDomain() {
  const settingsQuery = useSettings(
    (settings) => settings.DefaultRegistry.Hide
  );

  return (
    <span
      className={clsx({
        'cm-strikethrough': settingsQuery.isSuccess && settingsQuery.data,
      })}
    >
      docker.io
    </span>
  );
}
