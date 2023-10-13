import { AlertTriangle } from 'lucide-react';

export function HttpsWarning() {
  return (
    <div className="vertical-center">
      <AlertTriangle className="icon icon-warning" />
      <span className="text-warning">
        Your Portainer server is currently running under HTTP, change to a HTTPS
        connection to be able to use this feature to handle secrets.
      </span>
    </div>
  );
}

export function isHttps() {
  return window.location.protocol === 'https:';
}
