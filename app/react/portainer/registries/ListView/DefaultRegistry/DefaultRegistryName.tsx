import { useSettings } from 'Portainer/settings/queries';
import clsx from 'clsx';

export function DefaultRegistryName() {
  const settingsQuery = useSettings(
    (settings) => settings.DefaultRegistry.Hide
  );

  return (
    <span
      className={clsx({
        'cm-strikethrough': settingsQuery.isSuccess && settingsQuery.data,
      })}
    >
      Docker Hub (anonymous)
    </span>
  );
}
